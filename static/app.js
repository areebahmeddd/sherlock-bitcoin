(() => {
  const TX_DETAILS_EMPTY_HTML =
    '<div class="tx-details-empty">No transaction selected.</div>';

  const stemSelect = document.getElementById("block-stem-select");
  const blockIndexInput = document.getElementById("block-index-input");
  const statusMessage = document.getElementById("status-message");
  const classificationFilters = document.getElementById(
    "classification-filters",
  );

  const summaryFile = document.getElementById("summary-file");
  const summaryBlockCount = document.getElementById("summary-block-count");
  const summaryTotalTxs = document.getElementById("summary-total-txs");
  const summaryFlagged = document.getElementById("summary-flagged");
  const summaryHeuristics = document.getElementById("summary-heuristics");
  const summaryFeeStats = document.getElementById("summary-fee-stats");
  const summaryScriptTypes = document.querySelector(
    "#summary-script-types tbody",
  );

  const blockHeightEl = document.getElementById("block-height");
  const blockHashEl = document.getElementById("block-hash");
  const blockTimestampEl = document.getElementById("block-timestamp");
  const blockTxCountEl = document.getElementById("block-tx-count");
  const blockFlaggedEl = document.getElementById("block-flagged");
  const blockScriptTypes = document.querySelector("#block-script-types tbody");
  const blockFeeStats = document.getElementById("block-fee-stats");

  const txTableBody = document.querySelector("#tx-table tbody");
  const txCountLabel = document.getElementById("tx-count-label");
  const txPageLabel = document.getElementById("tx-page-label");
  const txDetails = document.getElementById("tx-details");
  const heuristicFilters = document.getElementById("heuristic-filters");
  const heuristicFilterPills = document.getElementById(
    "heuristic-filter-pills",
  );

  const state = {
    currentStem: "",
    currentBlockIndex: 0,
    currentFilter: "all",
    currentHeuristicFilter: "",
    currentTransactions: [],
    pageLimit: 50,
  };

  function setStatus(message) {
    statusMessage.textContent = message;
  }

  async function fetchJSON(path) {
    const res = await fetch(path, { headers: { Accept: "application/json" } });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(
        `Request failed (${res.status} ${res.statusText}): ${text}`,
      );
    }
    return res.json();
  }

  async function loadBlockList() {
    try {
      const data = await fetchJSON("/api/blocks");
      const stems = Array.isArray(data.blocks) ? data.blocks : [];
      stemSelect.innerHTML = "";

      if (stems.length === 0) {
        const opt = document.createElement("option");
        opt.value = "";
        opt.textContent = "No analyses found in out/";
        stemSelect.appendChild(opt);
        setStatus(
          "No analysis JSON files found. Run ./cli.sh --block first to generate results.",
        );
        return;
      }

      const placeholder = document.createElement("option");
      placeholder.value = "";
      placeholder.textContent = "Choose a file…";
      stemSelect.appendChild(placeholder);

      stems.forEach((stem) => {
        const opt = document.createElement("option");
        opt.value = stem;
        opt.textContent = stem;
        stemSelect.appendChild(opt);
      });

      setStatus("");
    } catch (err) {
      console.error(err);
      setStatus("Failed to load analyzed block list.");
      stemSelect.innerHTML = "";
      const opt = document.createElement("option");
      opt.value = "";
      opt.textContent = "Error loading blocks";
      stemSelect.appendChild(opt);
    }
  }

  const SCRIPT_TYPE_TOOLTIPS = {
    p2wpkh: "Pay to Witness Public Key Hash (SegWit, native bech32)",
    p2tr: "Pay to Taproot (Schnorr, bech32m)",
    p2pkh: "Pay to Public Key Hash (legacy)",
    p2sh: "Pay to Script Hash",
    p2wsh: "Pay to Witness Script Hash (SegWit)",
    multisig: "Multisignature script",
    op_return: "Provably unspendable / null data output",
    unknown: "Unrecognized script type",
  };

  function scriptTypeTooltip(script) {
    return SCRIPT_TYPE_TOOLTIPS[script] || `Script type: ${script}`;
  }

  function renderScriptTypeTable(tbody, dist) {
    tbody.innerHTML = "";
    if (!dist || typeof dist !== "object") {
      return;
    }
    const entries = Object.entries(dist);
    if (entries.length === 0) return;
    entries
      .sort((a, b) => b[1] - a[1])
      .forEach(([script, count]) => {
        const tr = document.createElement("tr");
        const tdType = document.createElement("td");
        const tdCount = document.createElement("td");
        tdType.textContent = script;
        tdType.title = scriptTypeTooltip(script);
        tdCount.textContent = count;
        tdCount.className = "cell-num";
        tr.appendChild(tdType);
        tr.appendChild(tdCount);
        tbody.appendChild(tr);
      });
  }

  const FEE_STAT_TOOLTIPS = {
    min: "Minimum fee rate (sat/vB) in this set",
    median: "Median fee rate (sat/vB)",
    max: "Maximum fee rate (sat/vB) in this set",
    mean: "Average fee rate (sat/vB)",
  };

  const HEURISTIC_TOOLTIPS = {
    cioh: "Common Input Ownership: multiple inputs likely from the same entity.",
    change_detection:
      "Identifies the likely change output (e.g. script type match, round number).",
    address_reuse: "Same address in inputs and/or outputs; weakens privacy.",
    coinjoin:
      "Multiple owners, equal-value outputs; obscures the transaction graph.",
    consolidation: "Many inputs merged into 1–2 outputs (wallet maintenance).",
    self_transfer: "Inputs and outputs appear to belong to the same entity.",
    peeling_chain:
      "Large input → small payment + large change, repeated pattern.",
    op_return:
      "Null data / provably unspendable output (e.g. Omni, timestamps).",
    round_number_payment:
      "Outputs with round BTC amounts (more likely payments).",
  };

  function renderFeeStats(dl, stats) {
    dl.innerHTML = "";
    if (!stats || typeof stats !== "object") {
      return;
    }
    const entries = [
      ["min", stats.min_sat_vb],
      ["median", stats.median_sat_vb],
      ["max", stats.max_sat_vb],
      ["mean", stats.mean_sat_vb],
    ];
    entries.forEach(([label, value]) => {
      if (value === undefined || value === null) return;
      const dt = document.createElement("dt");
      const dd = document.createElement("dd");
      dt.textContent = label;
      dt.title = FEE_STAT_TOOLTIPS[label] || label;
      dd.textContent = String(value);
      dl.appendChild(dt);
      dl.appendChild(dd);
    });
  }

  function renderTags(container, values) {
    container.innerHTML = "";
    if (!Array.isArray(values)) return;
    values.forEach((v) => {
      const span = document.createElement("span");
      span.className = "tag";
      span.textContent = v;
      container.appendChild(span);
    });
  }

  function isFlaggedTx(tx) {
    if (!tx || !tx.heuristics) return false;
    return Object.values(tx.heuristics).some(
      (h) => h && typeof h === "object" && h.detected,
    );
  }

  function formatTimestamp(unix) {
    if (!unix || unix <= 0) return "–";
    const d = new Date(unix * 1000);
    return d
      .toISOString()
      .replace("T", " ")
      .replace(/\.\d+Z$/, " UTC");
  }

  function classificationClass(name) {
    switch (name) {
      case "coinjoin":
        return "tag-class tag-class-coinjoin";
      case "consolidation":
        return "tag-class tag-class-consolidation";
      case "self_transfer":
        return "tag-class tag-class-self";
      case "batch_payment":
        return "tag-class tag-class-batch";
      case "simple_payment":
        return "tag-class tag-class-simple";
      default:
        return "tag-class tag-class-unknown";
    }
  }

  function renderFileSummary(data) {
    summaryFile.textContent = data.file || "–";
    summaryBlockCount.textContent =
      data.block_count != null ? String(data.block_count) : "–";

    const summary = data.analysis_summary || {};
    summaryTotalTxs.textContent =
      summary.total_transactions_analyzed != null
        ? String(summary.total_transactions_analyzed)
        : "–";
    summaryFlagged.textContent =
      summary.flagged_transactions != null
        ? String(summary.flagged_transactions)
        : "–";

    renderTags(summaryHeuristics, summary.heuristics_applied || []);
    renderFeeStats(summaryFeeStats, summary.fee_rate_stats || {});
    renderScriptTypeTable(
      summaryScriptTypes,
      summary.script_type_distribution || {},
    );
  }

  function renderBlockSummary(block) {
    if (!block) {
      blockHeightEl.textContent = "–";
      blockHashEl.textContent = "–";
      blockTimestampEl.textContent = "–";
      blockTxCountEl.textContent = "–";
      blockFlaggedEl.textContent = "–";
      blockScriptTypes.innerHTML = "";
      blockFeeStats.innerHTML = "";
      return;
    }
    blockHeightEl.textContent =
      block.block_height != null ? String(block.block_height) : "–";
    blockHashEl.textContent = block.block_hash || "–";
    blockTimestampEl.textContent = formatTimestamp(block.block_timestamp);
    blockTxCountEl.textContent =
      block.tx_count != null ? String(block.tx_count) : "–";

    const summary = block.analysis_summary || {};
    blockFlaggedEl.textContent =
      summary.flagged_transactions != null
        ? String(summary.flagged_transactions)
        : "–";
    renderScriptTypeTable(
      blockScriptTypes,
      summary.script_type_distribution || {},
    );
    renderFeeStats(blockFeeStats, summary.fee_rate_stats || {});
  }

  function shortenTxid(txid) {
    if (!txid || typeof txid !== "string") return "";
    if (txid.length <= 16) return txid;
    return txid.slice(0, 10) + "…" + txid.slice(-6);
  }

  function renderTransactionsTable() {
    txTableBody.innerHTML = "";
    txDetails.innerHTML = TX_DETAILS_EMPTY_HTML;

    const filtered = state.currentTransactions.filter((tx) => {
      let classMatch = true;
      if (state.currentFilter === "flagged") classMatch = isFlaggedTx(tx);
      else if (state.currentFilter !== "all")
        classMatch = tx.classification === state.currentFilter;

      let heurMatch = true;
      if (state.currentHeuristicFilter) {
        const h = tx.heuristics && tx.heuristics[state.currentHeuristicFilter];
        heurMatch = !!(h && h.detected);
      }

      return classMatch && heurMatch;
    });

    txCountLabel.textContent = `${
      filtered.length
    } transactions (page size ${state.pageLimit})`;
    txPageLabel.textContent = `showing offset 0 in block ${state.currentBlockIndex}`;

    if (filtered.length === 0) {
      const tr = document.createElement("tr");
      const td = document.createElement("td");
      td.colSpan = 3;
      td.className = "cell-empty";
      td.textContent = "No transactions for this view.";
      tr.appendChild(td);
      txTableBody.appendChild(tr);
      txDetails.innerHTML = TX_DETAILS_EMPTY_HTML;
      return;
    }

    filtered.forEach((tx) => {
      const tr = document.createElement("tr");
      if (isFlaggedTx(tx)) {
        tr.classList.add("tx-row-flagged");
      }

      const tdTxid = document.createElement("td");
      const tdClass = document.createElement("td");
      const tdHeur = document.createElement("td");

      tdTxid.textContent = shortenTxid(tx.txid);

      const classSpan = document.createElement("span");
      classSpan.className = classificationClass(tx.classification);
      classSpan.textContent = tx.classification || "unknown";
      tdClass.appendChild(classSpan);

      const heurContainer = document.createElement("div");
      heurContainer.className = "tag-row";
      const heuristics = tx.heuristics || {};
      Object.entries(heuristics).forEach(([id, res]) => {
        const tag = document.createElement("span");
        const active = res && res.detected;
        tag.className = active ? "tag tag-heur-active" : "tag tag-heur";
        tag.textContent = id;
        heurContainer.appendChild(tag);
      });
      tdHeur.appendChild(heurContainer);

      tr.appendChild(tdTxid);
      tr.appendChild(tdClass);
      tr.appendChild(tdHeur);

      tr.addEventListener("click", () => {
        renderTxDetails(tx);
      });

      txTableBody.appendChild(tr);
    });
  }

  function renderTxGraph(tx) {
    const CLS_COLORS = {
      coinjoin: "#7c3aed",
      consolidation: "#c2410c",
      self_transfer: "#0284c7",
      batch_payment: "#a16207",
      simple_payment: "#059669",
      unknown: "#9c9691",
    };

    const h = tx.heuristics || {};
    const cls = tx.classification || "unknown";
    const clsColor = CLS_COLORS[cls] || CLS_COLORS.unknown;

    // Derive implied input structure from heuristics
    let inputCount = 1;
    if (
      (h.consolidation && h.consolidation.detected) ||
      (h.coinjoin && h.coinjoin.detected)
    ) {
      inputCount = 4; // show "many"
    } else if (h.cioh && h.cioh.detected) {
      inputCount = 2;
    }
    const showMoreInputs = inputCount > 3;
    const visibleInCount = showMoreInputs ? 3 : inputCount;
    const inputNodes = [];
    for (let i = 0; i < visibleInCount; i++) {
      const isEllipsis = showMoreInputs && i === 2;
      inputNodes.push({
        label: isEllipsis ? "\u2026" : `In\u202f${i + 1}`,
        isEllipsis,
      });
    }

    // Derive implied output structure from heuristics
    const cd = h.change_detection;
    const changeIdx =
      cd &&
      cd.detected &&
      cd.likely_change_index != null &&
      cd.likely_change_index >= 0
        ? cd.likely_change_index
        : -1;

    const outputNodes = [];
    if (cls === "batch_payment" || (h.coinjoin && h.coinjoin.detected)) {
      outputNodes.push({
        label: "Out\u202f1",
        isChange: false,
        isEllipsis: false,
      });
      outputNodes.push({
        label: "Out\u202f2",
        isChange: false,
        isEllipsis: false,
      });
      outputNodes.push({ label: "\u2026", isChange: false, isEllipsis: true });
    } else if (h.consolidation && h.consolidation.detected) {
      outputNodes.push({
        label: changeIdx === 0 ? "Change" : "Output",
        isChange: changeIdx === 0,
        isEllipsis: false,
      });
      outputNodes.push({
        label: changeIdx === 1 ? "Change" : "Output",
        isChange: changeIdx === 1,
        isEllipsis: false,
      });
    } else if (changeIdx >= 0) {
      outputNodes.push({
        label: changeIdx === 0 ? "Change" : "Payment",
        isChange: changeIdx === 0,
        isEllipsis: false,
      });
      outputNodes.push({
        label: changeIdx === 1 ? "Change" : "Payment",
        isChange: changeIdx === 1,
        isEllipsis: false,
      });
    } else {
      outputNodes.push({ label: "Output", isChange: false, isEllipsis: false });
    }

    // SVG layout
    const NODE_R = 14;
    const NODE_SPACING = 44;
    const SVG_W = 400;
    const maxNodes = Math.max(inputNodes.length, outputNodes.length);
    const SVG_H = maxNodes * NODE_SPACING + 44;
    const centerY = SVG_H / 2;

    const TX_W = 96;
    const TX_H = 44;
    const TX_CX = SVG_W / 2;
    const TX_X = TX_CX - TX_W / 2;
    const TX_Y = centerY - TX_H / 2;

    const IN_X = 28;
    const OUT_X = SVG_W - 28;

    function nodeY(idx, count) {
      if (count === 1) return centerY;
      const totalSpan = (count - 1) * NODE_SPACING;
      return centerY - totalSpan / 2 + idx * NODE_SPACING;
    }

    const parts = [];

    // Edges: inputs → TX
    inputNodes.forEach((_, i) => {
      const iy = nodeY(i, inputNodes.length);
      parts.push(
        `<path d="M ${IN_X + NODE_R} ${iy} C ${IN_X + 70} ${iy}, ${TX_X - 20} ${centerY}, ${TX_X} ${centerY}" fill="none" stroke="#e6e4e2" stroke-width="1.5"/>`,
      );
    });

    // Edges: TX → outputs
    outputNodes.forEach((out, i) => {
      const oy = nodeY(i, outputNodes.length);
      const edgeColor = out.isChange ? "#0d9488" : "#e6e4e2";
      parts.push(
        `<path d="M ${TX_X + TX_W} ${centerY} C ${TX_X + TX_W + 20} ${centerY}, ${OUT_X - 70} ${oy}, ${OUT_X - NODE_R} ${oy}" fill="none" stroke="${edgeColor}" stroke-width="1.5"/>`,
      );
    });

    // TX rect (fill bg-alt, classification-coloured border)
    parts.push(
      `<rect x="${TX_X}" y="${TX_Y}" width="${TX_W}" height="${TX_H}" rx="6" fill="#f3f2f0" stroke="${clsColor}" stroke-width="1.5"/>`,
    );

    // TX classification label
    const clsLabel = cls.replace(/_/g, "\u00a0");
    parts.push(
      `<text x="${TX_CX}" y="${centerY + 4}" text-anchor="middle" font-size="10" font-weight="600" font-family="Source Sans 3,sans-serif" fill="${clsColor}">${clsLabel}</text>`,
    );

    // "inputs" column header
    const inFirstY = nodeY(0, inputNodes.length);
    parts.push(
      `<text x="${IN_X}" y="${inFirstY - NODE_R - 5}" text-anchor="middle" font-size="8" font-family="Source Sans 3,sans-serif" fill="#9c9691">inputs</text>`,
    );

    // Input nodes
    inputNodes.forEach((inp, i) => {
      const iy = nodeY(i, inputNodes.length);
      parts.push(
        `<circle cx="${IN_X}" cy="${iy}" r="${NODE_R}" fill="#f3f2f0" stroke="#e6e4e2" stroke-width="1.5"/>`,
      );
      parts.push(
        `<text x="${IN_X}" y="${iy + 4}" text-anchor="middle" font-size="${inp.isEllipsis ? "12" : "8"}" font-weight="500" font-family="Source Sans 3,sans-serif" fill="${inp.isEllipsis ? "#9c9691" : "#6b6560"}">${inp.label}</text>`,
      );
    });

    // "outputs" column header
    const outFirstY = nodeY(0, outputNodes.length);
    parts.push(
      `<text x="${OUT_X}" y="${outFirstY - NODE_R - 5}" text-anchor="middle" font-size="8" font-family="Source Sans 3,sans-serif" fill="#9c9691">outputs</text>`,
    );

    // Output nodes
    outputNodes.forEach((out, i) => {
      const oy = nodeY(i, outputNodes.length);
      const stroke = out.isChange ? "#0d9488" : "#e6e4e2";
      const fill = out.isChange ? "#0d94881a" : "#f3f2f0";
      const textFill = out.isChange
        ? "#0d9488"
        : out.isEllipsis
          ? "#9c9691"
          : "#6b6560";
      parts.push(
        `<circle cx="${OUT_X}" cy="${oy}" r="${NODE_R}" fill="${fill}" stroke="${stroke}" stroke-width="1.5"/>`,
      );
      parts.push(
        `<text x="${OUT_X}" y="${oy + 4}" text-anchor="middle" font-size="${out.isEllipsis ? "12" : "8"}" font-weight="500" font-family="Source Sans 3,sans-serif" fill="${textFill}">${out.label}</text>`,
      );
    });

    const wrapper = document.createElement("div");
    wrapper.className = "tx-graph";

    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("viewBox", `0 0 ${SVG_W} ${SVG_H}`);
    svg.setAttribute("width", "100%");
    svg.setAttribute("height", String(SVG_H));
    svg.setAttribute("aria-hidden", "true");
    svg.innerHTML = parts.join("\n");

    const caption = document.createElement("p");
    caption.className = "tx-graph__caption";
    caption.textContent = "Inferred from heuristics \u00b7 not raw UTXO data";

    wrapper.appendChild(svg);
    wrapper.appendChild(caption);

    return wrapper;
  }

  function renderTxDetails(tx) {
    if (!tx) {
      txDetails.innerHTML = TX_DETAILS_EMPTY_HTML;
      return;
    }

    const container = document.createElement("div");
    container.className = "tx-details-body";

    const heading = document.createElement("div");
    heading.className = "tx-details-heading";
    heading.innerHTML = `<div class="tx-details-label">Transaction</div><div class="tx-details-txid">${tx.txid}</div>`;
    container.appendChild(heading);
    container.appendChild(renderTxGraph(tx));

    const heuristics = tx.heuristics || {};

    const list = document.createElement("div");
    list.className = "tx-details-heuristics";

    Object.entries(heuristics).forEach(([id, res]) => {
      const item = document.createElement("div");
      item.className = "tx-details-item";

      const title = document.createElement("div");
      title.className = "tx-details-item-title";
      const detected = res && res.detected;
      title.textContent = id;
      title.title = HEURISTIC_TOOLTIPS[id] || id;
      if (detected) {
        const badge = document.createElement("span");
        badge.className = "badge badge-ok badge-inline";
        badge.textContent = "detected";
        title.appendChild(badge);
      } else {
        const badge = document.createElement("span");
        badge.className = "badge badge-muted badge-inline";
        badge.textContent = "not detected";
        title.appendChild(badge);
      }

      const body = document.createElement("div");
      body.className = "tx-details-item-body";

      if (res && typeof res === "object") {
        const entries = Object.entries(res).filter(
          ([key]) => key !== "detected",
        );
        if (entries.length === 0) {
          body.textContent = "No additional details.";
        } else {
          const dl = document.createElement("dl");
          entries.forEach(([k, v]) => {
            const dt = document.createElement("dt");
            const dd = document.createElement("dd");
            dt.textContent = k;
            dd.textContent = String(v);
            dl.appendChild(dt);
            dl.appendChild(dd);
          });
          body.appendChild(dl);
        }
      } else {
        body.textContent = "No additional details.";
      }

      item.appendChild(title);
      item.appendChild(body);
      list.appendChild(item);
    });

    container.appendChild(list);
    txDetails.innerHTML = "";
    txDetails.appendChild(container);
  }

  async function loadFileSummary(stem) {
    const data = await fetchJSON(
      `/api/blocks/${encodeURIComponent(stem)}/summary`,
    );
    renderFileSummary(data);
    return data;
  }

  async function loadBlockAndTransactions(stem, blockIndex) {
    const blockData = await fetchJSON(
      `/api/blocks/${encodeURIComponent(stem)}/blocks/${blockIndex}`,
    );
    const block = blockData.block;
    renderBlockSummary(block);

    const txData = await fetchJSON(
      `/api/blocks/${encodeURIComponent(
        stem,
      )}/transactions?block=${blockIndex}&limit=${state.pageLimit}&offset=0`,
    );
    const txs = Array.isArray(txData.transactions) ? txData.transactions : [];
    state.currentTransactions = txs;
    renderTransactionsTable();
  }

  async function onStemChange() {
    const stem = stemSelect.value;
    state.currentStem = stem;
    state.currentBlockIndex = 0;
    state.currentHeuristicFilter = "";
    blockIndexInput.value = "0";
    state.currentTransactions = [];
    clearHeuristicPills();
    renderTransactionsTable();
    renderBlockSummary(null);

    if (!stem) {
      renderFileSummary({ file: "", block_count: null, analysis_summary: {} });
      setStatus("");
      return;
    }

    try {
      const summary = await loadFileSummary(stem);
      const blockCount =
        typeof summary.block_count === "number" ? summary.block_count : 1;
      blockIndexInput.max = String(Math.max(blockCount - 1, 0));
      setStatus("");
      await loadBlockAndTransactions(stem, state.currentBlockIndex);
    } catch (err) {
      console.error(err);
      setStatus("Failed to load analysis for selected block file.");
    }
  }

  async function onBlockIndexChange() {
    if (!state.currentStem) {
      return;
    }
    const value = parseInt(blockIndexInput.value, 10);
    if (Number.isNaN(value) || value < 0) {
      blockIndexInput.value = String(state.currentBlockIndex);
      return;
    }
    state.currentBlockIndex = value;
    try {
      await loadBlockAndTransactions(state.currentStem, value);
      setStatus("");
    } catch (err) {
      console.error(err);
      setStatus("Failed to load that block index. Check the range.");
    }
  }

  function onFilterClick(evt) {
    const btn = evt.target.closest("button[data-filter]");
    if (!btn) return;
    const filter = btn.getAttribute("data-filter");
    if (!filter) return;

    state.currentFilter = filter;
    Array.from(
      classificationFilters.querySelectorAll("button[data-filter]"),
    ).forEach((b) => {
      b.classList.toggle(
        "pill-active",
        b.getAttribute("data-filter") === filter,
      );
    });
    renderTransactionsTable();
  }

  const ALL_HEURISTIC_IDS = [
    "address_reuse",
    "change_detection",
    "cioh",
    "coinjoin",
    "consolidation",
    "op_return",
    "peeling_chain",
    "round_number_payment",
    "self_transfer",
  ];

  function buildHeuristicPills() {
    heuristicFilterPills.innerHTML = "";

    const allBtn = document.createElement("button");
    allBtn.type = "button";
    allBtn.setAttribute("data-heuristic", "");
    allBtn.className = "pill pill-active";
    allBtn.textContent = "All";
    heuristicFilterPills.appendChild(allBtn);

    ALL_HEURISTIC_IDS.forEach((id) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.setAttribute("data-heuristic", id);
      btn.className = "pill";
      btn.title = HEURISTIC_TOOLTIPS[id] || id;
      btn.textContent = id;
      heuristicFilterPills.appendChild(btn);
    });
  }

  function clearHeuristicPills() {
    state.currentHeuristicFilter = "";
    Array.from(
      heuristicFilterPills.querySelectorAll("button[data-heuristic]"),
    ).forEach((b) => {
      b.classList.toggle("pill-active", b.getAttribute("data-heuristic") === "");
    });
  }

  function onHeuristicFilterClick(evt) {
    const btn = evt.target.closest("button[data-heuristic]");
    if (!btn) return;
    const filter = btn.getAttribute("data-heuristic");
    if (filter === null) return;

    state.currentHeuristicFilter = filter;
    Array.from(
      heuristicFilterPills.querySelectorAll("button[data-heuristic]"),
    ).forEach((b) => {
      b.classList.toggle(
        "pill-active",
        b.getAttribute("data-heuristic") === filter,
      );
    });
    renderTransactionsTable();
  }

  function initEvents() {
    stemSelect.addEventListener("change", onStemChange);
    blockIndexInput.addEventListener("change", onBlockIndexChange);
    classificationFilters.addEventListener("click", onFilterClick);
    heuristicFilters.addEventListener("click", onHeuristicFilterClick);
  }

  async function init() {
    buildHeuristicPills();
    await loadBlockList();
    renderTransactionsTable();
    renderBlockSummary(null);
  }

  window.addEventListener("DOMContentLoaded", () => {
    initEvents();
    init().catch((err) => {
      console.error(err);
      setStatus("Failed to initialize UI.");
    });
  });
})();
