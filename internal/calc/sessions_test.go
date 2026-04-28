package calc

import (
	"testing"
	"time"

	"github.com/lexeko/git-time/internal/model"
)

func TestEstimateSessions(t *testing.T) {
	base := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	gap := 120 * time.Minute
	offset := 120 * time.Minute

	tests := []struct {
		name      string
		minutes   []int
		wantHours int
	}{
		{
			name:      "no commits returns zero",
			minutes:   nil,
			wantHours: 0,
		},
		{
			name:      "one commit returns one offset",
			minutes:   []int{0},
			wantHours: 2,
		},
		{
			name:      "two commits within gap",
			minutes:   []int{0, 60},
			wantHours: 3,
		},
		{
			name:      "two commits far apart",
			minutes:   []int{0, 24 * 60},
			wantHours: 4,
		},
		{
			name:      "three commits in one session",
			minutes:   []int{0, 60, 120},
			wantHours: 4,
		},
		{
			name:      "split session after large gap",
			minutes:   []int{0, 60, 300},
			wantHours: 5,
		},
		{
			name:      "gap exactly equal stays in same session",
			minutes:   []int{0, 120},
			wantHours: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Estimate(commitsAt(base, "dev@example.com", tt.minutes...), nil, gap, offset, true)
			if result.TotalHours != tt.wantHours {
				t.Fatalf("TotalHours = %d, want %d", result.TotalHours, tt.wantHours)
			}
			if result.CommitCount != len(tt.minutes) {
				t.Fatalf("CommitCount = %d, want %d", result.CommitCount, len(tt.minutes))
			}
		})
	}
}

func TestEstimateNormalizesAliases(t *testing.T) {
	base := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	commits := []model.Commit{
		{AuthorEmail: "old@example.com", Time: base},
		{AuthorEmail: "new@example.com", Time: base.Add(time.Hour)},
	}

	result := Estimate(commits, map[string]string{"old@example.com": "new@example.com"}, 120*time.Minute, 120*time.Minute, true)
	if len(result.Authors) != 1 {
		t.Fatalf("authors = %d, want 1", len(result.Authors))
	}
	if result.Authors[0].Email != "new@example.com" {
		t.Fatalf("author email = %q, want new@example.com", result.Authors[0].Email)
	}
	if result.Authors[0].CommitCount != 2 {
		t.Fatalf("author commit count = %d, want 2", result.Authors[0].CommitCount)
	}
	if result.CommitCount != 2 {
		t.Fatalf("total commit count = %d, want 2", result.CommitCount)
	}
	if result.TotalHours != 3 {
		t.Fatalf("TotalHours = %d, want 3", result.TotalHours)
	}
}

func commitsAt(base time.Time, email string, minutes ...int) []model.Commit {
	commits := make([]model.Commit, 0, len(minutes))
	for _, minute := range minutes {
		commits = append(commits, model.Commit{
			AuthorEmail: email,
			Time:        base.Add(time.Duration(minute) * time.Minute),
		})
	}
	return commits
}
