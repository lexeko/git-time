package gitlog

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/lexeko/git-time/internal/model"
)

type Options struct {
	Path          string
	Branch        string
	AllBranches   bool
	IncludeMerges bool
	Since         *time.Time
	Until         *time.Time
}

func Collect(ctx context.Context, opts Options) ([]model.Commit, error) {
	path := opts.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "log", "--pretty=format:%H%x09%an%x09%ae%x09%at"}
	if opts.Since != nil {
		args = append(args, "--since=@"+strconv.FormatInt(opts.Since.Unix(), 10))
	}
	if opts.Until != nil {
		args = append(args, "--until=@"+strconv.FormatInt(opts.Until.Unix(), 10))
	}
	if !opts.IncludeMerges {
		args = append(args, "--no-merges")
	}
	if opts.AllBranches {
		args = append(args, "--all")
	} else if opts.Branch != "" {
		args = append(args, opts.Branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git log failed: %s", msg)
	}

	return parseLog(out)
}

func parseLog(out []byte) ([]model.Commit, error) {
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}

	lines := strings.Split(text, "\n")
	commits := make([]model.Commit, 0, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			return nil, fmt.Errorf("unexpected git log line %d: expected 4 fields, got %d", i+1, len(parts))
		}

		epoch, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp on git log line %d: %w", i+1, err)
		}

		commits = append(commits, model.Commit{
			Hash:        parts[0],
			AuthorName:  parts[1],
			AuthorEmail: parts[2],
			Time:        time.Unix(epoch, 0),
		})
	}

	return commits, nil
}
