package aso

import (
	"encoding/json"
	"io"
	"strings"
)

// parseExclude splits a comma-separated exclude string into a slice of field names.
func parseExclude(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, f := range strings.Split(raw, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// writeJSON encodes v as indented JSON to w, optionally removing top-level
// fields listed in exclude from each object in the output.
func writeJSON(w io.Writer, v any, exclude []string) error {
	if len(exclude) == 0 {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Build exclude set.
	drop := make(map[string]bool, len(exclude))
	for _, k := range exclude {
		drop[k] = true
	}

	// Try as array of objects.
	var arr []json.RawMessage
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		filtered := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			var obj map[string]any
			if err := json.Unmarshal(item, &obj); err != nil {
				continue
			}
			for k := range drop {
				delete(obj, k)
			}
			filtered = append(filtered, obj)
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}

	// Try as single object.
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		for k := range drop {
			delete(obj, k)
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(obj)
	}

	// Fallback: write unfiltered.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
