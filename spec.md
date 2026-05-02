# git-time

Build a Go CLI named `git-time`.

## Purpose

Estimate how many hours a developer worked using only Git commit timestamps.

The result is an approximation, not a precise time tracker.

## Core Assumptions

Commits act as a proxy for active work.

If two commits are close together, the developer was probably working
continuously between them.

If two commits are far apart, the developer likely stopped and later started a
new work session.

Work before the first commit in a session is invisible in Git history, so each
session receives a fixed session offset.

## Inputs

- Git repository path
- Commit timestamps collected from `git log`
- Commit author names and email addresses collected from `git log`
- Optional author email aliases
- Session gap threshold, in minutes
- Session offset, in minutes
- Optional since/until date filters
- Optional branch or all-branches selection
- Optional merge-commit inclusion/exclusion
- Output format and session detail options

## Output

Default text output:

```text
<hours> hours, <commit-count> commits
```

When `--sessions` is enabled, text output includes:

- Total estimated hours, minutes, and commit count
- Per-author estimated hours, minutes, and commit count
- Per-session start/end time, estimated minutes, and commit count

When `--format json` is enabled, output is an indented JSON encoding of the
result model.

## Algorithm

1. Collect commits using the installed `git` CLI.
2. Filter commits through `git log` by:
   - since date
   - until date
   - branch or all branches
   - merge policy
3. Normalize author emails using aliases.
4. Group commits by normalized author email.
5. Process authors in sorted email order.
6. For each author:
   - Sort commit timestamps from oldest to newest.
   - Split commits into sessions:
     - Start the first session at the first commit.
     - For each next commit:
       - If the gap from the previous commit is less than or equal to
         `session-gap`, keep it in the current session.
       - Otherwise, start a new session.
   - For each session:
     - Session duration is `last commit timestamp - first commit timestamp`.
     - Estimated session time is `session duration + session offset`.
   - Sum estimated session time across sessions.
   - Round the author's total estimated hours to the nearest whole hour.
7. Sum totals across authors.
8. Round total estimated hours to the nearest whole hour.

## Edge Cases

| Situation                          | Behavior                                                     |
| ---------------------------------- | ------------------------------------------------------------ |
| No commits                         | Returns 0 hours and 0 commits                                |
| One commit                         | Returns one session offset                                   |
| Two commits within gap             | Returns gap + one session offset                             |
| Two commits farther apart than gap | Returns two session offsets                                  |
| Gap exactly equal to session gap   | Keeps commits in the same session                            |
| All commits within one session     | Returns first-to-last duration + one session offset          |
| One commit per day                 | Each day is a separate session; each gets one session offset |
| Many rapid commits                 | Tiny gaps are summed literally inside the session            |
| Merge commits excluded             | Ignores merge commits before session grouping                |
| Multiple emails for one person     | Normalizes aliases before grouping                           |
| Alias cycle                        | Stops normalization and returns the repeated normalized email |

## What It Cannot Know

- Time spent reading, researching, or planning without committing
- Work done after the last commit of a session
- Actual breaks inside a session
- Whether a large gap means a break, sleep, or switching projects

## Implementation Requirements

Use Go.

Use the Go standard library where possible.

Do not use libgit2 or Go Git libraries.

Shell out to the installed `git` command.

The CLI command is:

```sh
git-time [options]
```

## Options

`--path <path>`, `-p <path>`

Path to the Git repository to analyze.

Default: current directory.

`--session-gap <minutes>`, `-g <minutes>`

Maximum idle time between commits that still counts as one continuous session.

Default: 120.

Must be non-negative.

`--session-offset <minutes>`, `-o <minutes>`

Estimated minutes of work credited at the beginning of each session.

Default: 120.

Must be non-negative.

`--since <date>`, `-s <date>`

Only include commits on or after this date.

Accepted values:

- `all`
- `today`
- `yesterday`
- `thisweek`
- `lastweek`
- `thismonth`
- `lastmonth`
- `YYYY-MM-DD`

Default: `all`.

`--until <date>`, `-u <date>`

Only include commits on or before this date.

Accepted values are the same as `--since`.

Default: `all`.

When both bounds are provided, `--since` must be on or before `--until`.

`--branch <name>`, `-b <name>`

Restrict analysis to a single branch.

Default: the current branch selected by `git log`.

Cannot be used with `--all-branches`.

`--all-branches`, `-a`

Analyze all branches by passing `--all` to `git log`.

`git log --all` lists commits, not branch appearances, so commits reachable
from multiple branches are not double-counted.

`--merges`, `-m`

Include merge commits.

`--no-merges`

Exclude merge commits.

Default: merge commits are excluded.

If both `--merges` and `--no-merges` are provided, `--no-merges` wins.

`--email-alias <other=main>`, `-e <other=main>`

Treat one email address as another.

Repeatable.

Example:

```sh
git-time --email-alias old@example.com=new@example.com
```

Aliases are normalized to lowercase and trimmed before use.

`--format <text|json>`, `-f <text|json>`

Output format.

Default: `text`.

`--sessions`

Show session breakdown in text output and include sessions in JSON output.

`--help`, `-h`

Show help.

## Git Command Guidance

Use `git log`.

Current format:

```sh
git log --pretty=format:%H%x09%an%x09%ae%x09%at
```

Fields:

- Commit hash
- Author name
- Author email
- Author timestamp as Unix epoch seconds

For excluding merge commits, add:

```sh
--no-merges
```

For including all branches, add:

```sh
--all
```

For a specific branch, pass the branch name.

For a repository path, run Git with:

```sh
git -C <path> ...
```

Date filters are passed as Unix timestamp bounds:

```sh
--since=@<epoch-seconds>
--until=@<epoch-seconds>
```

## Project Structure

```text
cmd/git-time/main.go
internal/calc/sessions.go
internal/gitlog/gitlog.go
internal/model/model.go
internal/output/output.go
```

## Result Model

The JSON result contains:

- `commit_count`
- `total_minutes`
- `total_hours`
- `authors`

Each author contains:

- `email`
- `commit_count`
- `total_minutes`
- `total_hours`
- `sessions`, when requested

Each session contains:

- `author_email`
- `start`
- `end`
- `commits`
- `minutes`

## Testing Requirements

Unit tests cover the session calculation logic.

Required cases:

1. No commits returns 0.
2. One commit returns one session offset.
3. Two commits 60 minutes apart with a 120-minute gap and 120-minute offset
   returns 3 hours.
4. Two commits 24 hours apart returns 4 hours.
5. Three commits at 10:00, 11:00, and 12:00 return 4 hours with a 120-minute
   offset.
6. Commits at 10:00, 11:00, and 15:00 return 5 hours with a 120-minute gap and
   120-minute offset.
   - Session 1: 10:00-11:00 + 2h = 3h.
   - Session 2: 15:00 + 2h = 2h.
7. A gap exactly equal to `session-gap` stays in the same session.
8. Alias normalization groups emails together.
9. JSON output is valid.
10. Default text output includes the total hours and commit count.
11. Help output is available through `--help` and `-h`.

## Non-Goals

- No daemon or background tracking
- No editor integration
- No GitHub API
- No automatic identity guessing beyond Git log author email
- No author inclusion/exclusion filter
- No database
- No dependency on Node.js

## README Requirements

Include:

- What the tool does
- Why it is an approximation
- Install instructions using `go install`
- Basic usage examples
- Explanation of `session-gap` and `session-offset`
- Example JSON output
- Inspiration note
