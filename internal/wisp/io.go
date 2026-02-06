package wisp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir ensures the .beads directory exists in the given root.
func EnsureDir(root string) (string, error) {
	dir := filepath.Join(root, WispDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create beads dir: %w", err)
	}
	return dir, nil
}

// WispPath returns the full path to a file in the beads directory.
func WispPath(root, filename string) string {
	return filepath.Join(root, WispDir, filename)
}

// writeJSON is a helper to write JSON files atomically.
func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	// Write to temp file then rename for atomicity
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil { //nolint:gosec // G306: wisp messages are non-sensitive operational data
		return fmt.Errorf("write temp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp) // cleanup on failure
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// WispCreateJSON is the JSON structure returned by bd mol wisp --json.
type WispCreateJSON struct {
	NewEpicID string `json:"new_epic_id"`
	RootID    string `json:"root_id"`
	ResultID  string `json:"result_id"`
}

// ParseWispIDFromJSON extracts the wisp ID from bd mol wisp --json output.
// It checks for new_epic_id, root_id, and result_id fields in that order.
func ParseWispIDFromJSON(jsonOutput []byte) (string, error) {
	var result WispCreateJSON
	if err := json.Unmarshal(jsonOutput, &result); err != nil {
		return "", fmt.Errorf("parsing wisp JSON: %w (output: %s)", err, trimJSONForError(jsonOutput))
	}

	switch {
	case result.NewEpicID != "":
		return result.NewEpicID, nil
	case result.RootID != "":
		return result.RootID, nil
	case result.ResultID != "":
		return result.ResultID, nil
	default:
		return "", fmt.Errorf("wisp JSON missing id field (expected one of new_epic_id, root_id, result_id); output: %s", trimJSONForError(jsonOutput))
	}
}

// trimJSONForError truncates JSON output for error messages.
func trimJSONForError(jsonOutput []byte) string {
	const maxLen = 500
	s := string(jsonOutput)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
