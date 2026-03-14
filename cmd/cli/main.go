package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sherlock/internal/analysis"
	"sherlock/internal/report"
)

func errorJSON(code, message string) {
	out, _ := json.Marshal(map[string]interface{}{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
	fmt.Println(string(out))
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 || args[0] != "--block" {
		errorJSON("INVALID_ARGS", "Usage: cli --block <blk.dat> <rev.dat> <xor.dat>")
		os.Exit(1)
	}

	if len(args) < 4 {
		errorJSON("INVALID_ARGS", "--block mode requires: <blk.dat> <rev.dat> <xor.dat>")
		os.Exit(1)
	}

	blkFile := args[1]
	revFile := args[2]
	xorFile := args[3]

	for _, f := range []string{blkFile, revFile, xorFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			errorJSON("FILE_NOT_FOUND", "File not found: "+f)
			os.Exit(1)
		}
	}

	if err := os.MkdirAll("out", 0o755); err != nil {
		errorJSON("IO_ERROR", "Cannot create out/ directory: "+err.Error())
		os.Exit(1)
	}

	result, err := analysis.AnalyzeFile(blkFile, revFile, xorFile)
	if err != nil {
		errorJSON("ANALYSIS_ERROR", err.Error())
		os.Exit(1)
	}

	// Derive output stem from blk filename
	blkBase := filepath.Base(blkFile)
	blkStem := strings.TrimSuffix(blkBase, ".dat")

	// Write JSON output
	jsonPath := filepath.Join("out", blkStem+".json")
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		errorJSON("JSON_ERROR", err.Error())
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		errorJSON("IO_ERROR", "Cannot write JSON: "+err.Error())
		os.Exit(1)
	}

	// Write Markdown output
	mdPath := filepath.Join("out", blkStem+".md")
	mdContent := report.Generate(result)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		errorJSON("IO_ERROR", "Cannot write Markdown: "+err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Analysis complete: %d blocks, %d transactions\n",
		result.BlockCount, result.AnalysisSummary.TotalTxsAnalyzed)
}
