package aso

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONNoExclude(t *testing.T) {
	data := map[string]any{"keyword": "camera", "popularity": 72, "topApps": []string{"app1"}}
	var buf bytes.Buffer
	if err := writeJSON(&buf, data, nil); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["topApps"]; !ok {
		t.Error("topApps should be present when no exclude")
	}
}

func TestWriteJSONExcludeObject(t *testing.T) {
	data := map[string]any{"keyword": "camera", "popularity": 72, "topApps": []string{"app1"}}
	var buf bytes.Buffer
	if err := writeJSON(&buf, data, []string{"topApps"}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["topApps"]; ok {
		t.Error("topApps should be excluded")
	}
	if got["keyword"] != "camera" {
		t.Errorf("keyword = %v, want camera", got["keyword"])
	}
}

func TestWriteJSONExcludeArray(t *testing.T) {
	data := []map[string]any{
		{"keyword": "camera", "topApps": []string{"a"}, "relatedSearches": []string{"b"}},
		{"keyword": "photo", "topApps": []string{"c"}, "relatedSearches": []string{"d"}},
	}
	var buf bytes.Buffer
	if err := writeJSON(&buf, data, []string{"topApps", "relatedSearches"}); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	for i, item := range got {
		if _, ok := item["topApps"]; ok {
			t.Errorf("item[%d]: topApps should be excluded", i)
		}
		if _, ok := item["relatedSearches"]; ok {
			t.Errorf("item[%d]: relatedSearches should be excluded", i)
		}
		if _, ok := item["keyword"]; !ok {
			t.Errorf("item[%d]: keyword should be present", i)
		}
	}
}

func TestParseExclude(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"topApps", 1},
		{"topApps,relatedSearches", 2},
		{"topApps, relatedSearches , foo", 3},
		{",,,", 0},
	}
	for _, tt := range tests {
		got := parseExclude(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseExclude(%q) = %v (len %d), want len %d", tt.input, got, len(got), tt.want)
		}
	}
}
