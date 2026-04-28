package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/lexeko/git-time/internal/model"
)

func Write(w io.Writer, result model.Result, format string, showSessions bool) error {
	switch format {
	case "text":
		return writeText(w, result, showSessions)
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func writeText(w io.Writer, result model.Result, showSessions bool) error {
	if !showSessions {
		_, err := fmt.Fprintf(w, "%d hours, %d %s\n", result.TotalHours, result.CommitCount, plural(result.CommitCount, "commit", "commits"))
		return err
	}

	if _, err := fmt.Fprintf(w, "total: %d hours (%d minutes), %d %s\n", result.TotalHours, result.TotalMinutes, result.CommitCount, plural(result.CommitCount, "commit", "commits")); err != nil {
		return err
	}
	for _, author := range result.Authors {
		if _, err := fmt.Fprintf(w, "%s: %d hours (%d minutes), %d %s\n", author.Email, author.TotalHours, author.TotalMinutes, author.CommitCount, plural(author.CommitCount, "commit", "commits")); err != nil {
			return err
		}
		for _, session := range author.Sessions {
			if _, err := fmt.Fprintf(
				w,
				"  %s - %s: %d minutes, %d %s\n",
				formatTime(session.Start),
				formatTime(session.End),
				session.Minutes,
				session.Commits,
				plural(session.Commits, "commit", "commits"),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
}

func plural(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
