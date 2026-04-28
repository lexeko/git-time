package calc

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/lexeko/git-time/internal/model"
)

func Estimate(commits []model.Commit, aliases map[string]string, sessionGap, sessionOffset time.Duration, includeSessions bool) model.Result {
	byAuthor := make(map[string][]model.Commit)
	for _, commit := range commits {
		email := NormalizeEmail(commit.AuthorEmail, aliases)
		commit.AuthorEmail = email
		byAuthor[email] = append(byAuthor[email], commit)
	}

	authors := make([]string, 0, len(byAuthor))
	for email := range byAuthor {
		authors = append(authors, email)
	}
	sort.Strings(authors)

	result := model.Result{}
	for _, email := range authors {
		authorCommits := byAuthor[email]
		sort.Slice(authorCommits, func(i, j int) bool {
			return authorCommits[i].Time.Before(authorCommits[j].Time)
		})

		author := model.AuthorResult{Email: email, CommitCount: len(authorCommits)}
		sessions := buildSessions(email, authorCommits, sessionGap, sessionOffset)
		for _, session := range sessions {
			author.TotalMinutes += session.Minutes
		}
		author.TotalHours = roundHours(author.TotalMinutes)
		if includeSessions {
			author.Sessions = sessions
		}

		result.CommitCount += author.CommitCount
		result.TotalMinutes += author.TotalMinutes
		result.Authors = append(result.Authors, author)
	}

	result.TotalHours = roundHours(result.TotalMinutes)
	return result
}

func NormalizeEmail(email string, aliases map[string]string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	seen := map[string]bool{}
	for {
		if seen[normalized] {
			return normalized
		}
		seen[normalized] = true

		main, ok := aliases[normalized]
		if !ok {
			return normalized
		}
		normalized = strings.ToLower(strings.TrimSpace(main))
	}
}

func buildSessions(author string, commits []model.Commit, sessionGap, sessionOffset time.Duration) []model.Session {
	if len(commits) == 0 {
		return nil
	}

	var sessions []model.Session
	start := commits[0].Time
	end := commits[0].Time
	count := 1

	for i := 1; i < len(commits); i++ {
		current := commits[i].Time
		if current.Sub(end) <= sessionGap {
			end = current
			count++
			continue
		}

		sessions = append(sessions, makeSession(author, start, end, count, sessionOffset))
		start = current
		end = current
		count = 1
	}

	sessions = append(sessions, makeSession(author, start, end, count, sessionOffset))
	return sessions
}

func makeSession(author string, start, end time.Time, commits int, sessionOffset time.Duration) model.Session {
	minutes := int(end.Sub(start).Minutes() + sessionOffset.Minutes())
	return model.Session{
		AuthorEmail: author,
		Start:       start,
		End:         end,
		Commits:     commits,
		Minutes:     minutes,
	}
}

func roundHours(minutes int) int {
	return int(math.Round(float64(minutes) / 60))
}
