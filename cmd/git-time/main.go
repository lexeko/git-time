package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/lexeko/git-time/internal/calc"
	"github.com/lexeko/git-time/internal/gitlog"
	"github.com/lexeko/git-time/internal/output"
)

type aliasFlags map[string]string

func (a aliasFlags) String() string {
	return fmt.Sprint(map[string]string(a))
}

func (a aliasFlags) Set(value string) error {
	other, main, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(other) == "" || strings.TrimSpace(main) == "" {
		return fmt.Errorf("expected other=main")
	}
	a[strings.ToLower(strings.TrimSpace(other))] = strings.ToLower(strings.TrimSpace(main))
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "git-time:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	var (
		repoPath      string
		sessionGap    int
		sessionOffset int
		sinceValue    string
		untilValue    string
		branch        string
		allBranches   bool
		merges        bool
		noMerges      bool
		format        string
		showSessions  bool
		aliases       = aliasFlags{}
	)

	flags := flag.NewFlagSet("git-time", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&repoPath, "path", ".", "path to the Git repository")
	flags.StringVar(&repoPath, "p", ".", "path to the Git repository")
	flags.IntVar(&sessionGap, "session-gap", 120, "maximum idle minutes within one session")
	flags.IntVar(&sessionGap, "g", 120, "maximum idle minutes within one session")
	flags.IntVar(&sessionOffset, "session-offset", 120, "minutes credited at the beginning of each session")
	flags.IntVar(&sessionOffset, "o", 120, "minutes credited at the beginning of each session")
	flags.StringVar(&sinceValue, "since", "all", "only include commits on or after this date")
	flags.StringVar(&sinceValue, "s", "all", "only include commits on or after this date")
	flags.StringVar(&untilValue, "until", "all", "only include commits on or before this date")
	flags.StringVar(&untilValue, "u", "all", "only include commits on or before this date")
	flags.StringVar(&branch, "branch", "", "restrict analysis to a branch")
	flags.StringVar(&branch, "b", "", "restrict analysis to a branch")
	flags.BoolVar(&allBranches, "all-branches", false, "analyze all branches")
	flags.BoolVar(&allBranches, "a", false, "analyze all branches")
	flags.BoolVar(&merges, "merges", false, "include merge commits")
	flags.BoolVar(&merges, "m", false, "include merge commits")
	flags.BoolVar(&noMerges, "no-merges", false, "exclude merge commits")
	flags.Var(aliases, "email-alias", "treat one email address as another, in other=main form")
	flags.Var(aliases, "e", "treat one email address as another, in other=main form")
	flags.StringVar(&format, "format", "text", "output format: text or json")
	flags.StringVar(&format, "f", "text", "output format: text or json")
	flags.BoolVar(&showSessions, "sessions", false, "show session breakdown")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if sessionGap < 0 {
		return errors.New("--session-gap must be non-negative")
	}
	if sessionOffset < 0 {
		return errors.New("--session-offset must be non-negative")
	}
	if format != "text" && format != "json" {
		return errors.New("--format must be text or json")
	}
	if allBranches && branch != "" {
		return errors.New("--branch cannot be used with --all-branches")
	}
	if noMerges {
		merges = false
	}

	now := time.Now()
	since, err := resolveDate(sinceValue, true, now)
	if err != nil {
		return fmt.Errorf("--since: %w", err)
	}
	until, err := resolveDate(untilValue, false, now)
	if err != nil {
		return fmt.Errorf("--until: %w", err)
	}
	if since != nil && until != nil && since.After(*until) {
		return errors.New("--since must be on or before --until")
	}

	commits, err := gitlog.Collect(context.Background(), gitlog.Options{
		Path:          repoPath,
		Branch:        branch,
		AllBranches:   allBranches,
		IncludeMerges: merges,
		Since:         since,
		Until:         until,
	})
	if err != nil {
		return err
	}

	result := calc.Estimate(
		commits,
		aliases,
		time.Duration(sessionGap)*time.Minute,
		time.Duration(sessionOffset)*time.Minute,
		showSessions,
	)

	return output.Write(stdout, result, format, showSessions)
}

func resolveDate(value string, startBound bool, now time.Time) (*time.Time, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "all" {
		return nil, nil
	}

	loc := now.Location()
	dayStart := func(t time.Time) time.Time {
		y, m, d := t.In(loc).Date()
		return time.Date(y, m, d, 0, 0, 0, 0, loc)
	}
	dayEnd := func(t time.Time) time.Time {
		return dayStart(t).AddDate(0, 0, 1).Add(-time.Second)
	}
	choose := func(start, end time.Time) time.Time {
		if startBound {
			return start
		}
		return end
	}

	today := dayStart(now)
	switch value {
	case "today":
		t := choose(today, dayEnd(today))
		return &t, nil
	case "yesterday":
		yesterday := today.AddDate(0, 0, -1)
		t := choose(yesterday, dayEnd(yesterday))
		return &t, nil
	case "thisweek":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := today.AddDate(0, 0, -(weekday - 1))
		t := choose(start, start.AddDate(0, 0, 7).Add(-time.Second))
		return &t, nil
	case "lastweek":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisWeekStart := today.AddDate(0, 0, -(weekday - 1))
		start := thisWeekStart.AddDate(0, 0, -7)
		t := choose(start, thisWeekStart.Add(-time.Second))
		return &t, nil
	case "thismonth":
		y, m, _ := today.Date()
		start := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		t := choose(start, start.AddDate(0, 1, 0).Add(-time.Second))
		return &t, nil
	case "lastmonth":
		y, m, _ := today.Date()
		thisMonthStart := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		start := thisMonthStart.AddDate(0, -1, 0)
		t := choose(start, thisMonthStart.Add(-time.Second))
		return &t, nil
	default:
		parsed, err := time.ParseInLocation("2006-01-02", value, loc)
		if err != nil {
			return nil, fmt.Errorf("unsupported date %q", value)
		}
		t := choose(parsed, dayEnd(parsed))
		return &t, nil
	}
}
