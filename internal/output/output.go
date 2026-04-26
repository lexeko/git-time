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
		_, err := fmt.Fprintf(w, "%d\n", result.TotalHours)
		return err
	}

	if _, err := fmt.Fprintf(w, "total: %d hours (%d minutes)\n", result.TotalHours, result.TotalMinutes); err != nil {
		return err
	}
	for _, author := range result.Authors {
		if _, err := fmt.Fprintf(w, "%s: %d hours (%d minutes)\n", author.Email, author.TotalHours, author.TotalMinutes); err != nil {
			return err
		}
		for _, session := range author.Sessions {
			if _, err := fmt.Fprintf(
				w,
				"  %s - %s: %d minutes, %d commits\n",
				formatTime(session.Start),
				formatTime(session.End),
				session.Minutes,
				session.Commits,
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
