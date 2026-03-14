package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ok":true}`)
}

// handleListBlocks serves GET /api/blocks — lists available analysis stems.
func handleListBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := filepath.Glob("out/*.json")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "glob_error", err.Error())
		return
	}

	var stems []string
	for _, path := range entries {
		base := filepath.Base(path)
		stems = append(stems, strings.TrimSuffix(base, ".json"))
	}
	if stems == nil {
		stems = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "blocks": stems})
}

// handleBlocksDispatch routes /api/blocks/<stem>/... sub-paths.
func handleBlocksDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/blocks/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) == 0 || parts[0] == "" {
		jsonError(w, http.StatusBadRequest, "missing_stem", "block stem required")
		return
	}
	stem := parts[0]

	data, result, err := loadAnalysis(stem)
	if err != nil {
		jsonError(w, http.StatusNotFound, "not_found", fmt.Sprintf("no analysis for %q: %v", stem, err))
		return
	}

	if len(parts) == 1 {
		// GET /api/blocks/<stem> → full JSON
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
		return
	}

	sub := parts[1]
	switch sub {
	case "summary":
		// GET /api/blocks/<stem>/summary
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":               true,
			"file":             result["file"],
			"block_count":      result["block_count"],
			"analysis_summary": result["analysis_summary"],
		})

	case "blocks":
		// GET /api/blocks/<stem>/blocks/<index>
		if len(parts) < 3 {
			jsonError(w, http.StatusBadRequest, "missing_index", "block index required")
			return
		}
		idx, err := strconv.Atoi(parts[2])
		if err != nil || idx < 0 {
			jsonError(w, http.StatusBadRequest, "invalid_index", "index must be a non-negative integer")
			return
		}
		blocks, _ := result["blocks"].([]interface{})
		if idx >= len(blocks) {
			jsonError(w, http.StatusNotFound, "index_out_of_range",
				fmt.Sprintf("index %d out of range (block_count=%d)", idx, len(blocks)))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "block": blocks[idx]})

	case "transactions":
		// GET /api/blocks/<stem>/transactions?block=N&limit=M&offset=O
		blockIdx := intQuery(r, "block", 0)
		limit := intQuery(r, "limit", 50)
		offset := intQuery(r, "offset", 0)

		blocks, _ := result["blocks"].([]interface{})
		if blockIdx < 0 || blockIdx >= len(blocks) {
			jsonError(w, http.StatusNotFound, "index_out_of_range",
				fmt.Sprintf("block index %d out of range", blockIdx))
			return
		}
		blockObj, _ := blocks[blockIdx].(map[string]interface{})
		txs, _ := blockObj["transactions"].([]interface{})

		total := len(txs)
		if offset >= total {
			txs = []interface{}{}
		} else {
			txs = txs[offset:]
			if limit > 0 && limit < len(txs) {
				txs = txs[:limit]
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":           true,
			"block_index":  blockIdx,
			"total":        total,
			"offset":       offset,
			"limit":        limit,
			"transactions": txs,
		})

	default:
		jsonError(w, http.StatusNotFound, "unknown_route", fmt.Sprintf("unknown sub-resource %q", sub))
	}
}

// loadAnalysis reads and parses out/<stem>.json.
func loadAnalysis(stem string) ([]byte, map[string]interface{}, error) {
	path := filepath.Join("out", stem+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, fmt.Errorf("parse JSON: %w", err)
	}
	return data, result, nil
}

// intQuery reads an integer query parameter with a default fallback.
func intQuery(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// jsonError writes a structured JSON error response.
func jsonError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}
