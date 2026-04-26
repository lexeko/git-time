package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/lexeko/git-time/internal/model"
)

func TestWriteJSONIsValid(t *testing.T) {
	result := model.Result{
		TotalMinutes: 180,
		TotalHours:   3,
		Authors: []model.AuthorResult{
			{
				Email:        "dev@example.com",
				TotalMinutes: 180,
				TotalHours:   3,
				Sessions: []model.Session{
					{
						AuthorEmail: "dev@example.com",
						Start:       time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
						End:         time.Date(2026, 1, 2, 11, 0, 0, 0, time.UTC),
						Commits:     2,
						Minutes:     180,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Write(&buf, result, "json", true); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	var decoded model.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if decoded.TotalHours != 3 {
		t.Fatalf("decoded TotalHours = %d, want 3", decoded.TotalHours)
	}
}
