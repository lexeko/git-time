# git-time

Estimate working hours from Git commit history.

`git-time` is a small Go CLI that shells out to the installed `git` command and
uses commit timestamps as a proxy for active work. It is an approximation, not a
time tracker: Git history cannot know about reading, planning, breaks inside a
session, or work after the last commit.

## Install

```sh
go install github.com/lexeko/git-time/cmd/git-time@latest
```

From a local checkout:

```sh
go install ./cmd/git-time
```

## Usage

Analyze the current branch in the current repository:

```sh
git-time
```

Analyze a different repository:

```sh
git-time --path ~/src/project
```

Analyze all branches without double-counting commits reachable from multiple
branches:

```sh
git-time --all-branches
```

Show session details:

```sh
git-time --sessions
```

Write JSON:

```sh
git-time --format json
```

Write JSON with session details:

```sh
git-time --format json --sessions
```

Filter by date:

```sh
git-time --since thisweek --until today
git-time --since 2026-01-01 --until 2026-01-31
```

Treat multiple emails as one person:

```sh
git-time --email-alias old@example.com=new@example.com
```

## Options

```text
--path, -p <path>                 Git repository path (default: current directory)
--session-gap, -g <minutes>       Max gap between commits in one session (default: 120)
--session-offset, -o <minutes>    Minutes credited at the start of each session (default: 120)
--since, -s <date>                all, today, yesterday, thisweek, lastweek, thismonth, lastmonth, YYYY-MM-DD
--until, -u <date>                all, today, yesterday, thisweek, lastweek, thismonth, lastmonth, YYYY-MM-DD
--branch, -b <name>               Analyze one branch (default: current branch)
--all-branches, -a                Analyze all branches
--merges, -m                      Include merge commits
--no-merges                       Exclude merge commits (default)
--email-alias, -e <other=main>    Repeatable email alias mapping
--format, -f <text|json>          Output format (default: text)
--sessions                        Show session breakdown
--help, -h                        Show help
```

## How Sessions Work

`session-gap` is the largest idle gap, in minutes, that still counts as one
continuous work session. With the default `--session-gap 120`, commits at 10:00,
11:00, and 12:00 are one session, but commits at 10:00 and 15:00 are separate
sessions.

`session-offset` estimates invisible work before the first commit in each
session. With the default `--session-offset 120`, a single isolated commit counts
as two hours.

The final total is rounded to the nearest whole hour.

## JSON Example

```json
{
  "total_minutes": 180,
  "total_hours": 3,
  "authors": [
    {
      "email": "dev@example.com",
      "total_minutes": 180,
      "total_hours": 3,
      "sessions": [
        {
          "author_email": "dev@example.com",
          "start": "2026-01-02T10:00:00Z",
          "end": "2026-01-02T11:00:00Z",
          "commits": 2,
          "minutes": 180
        }
      ]
    }
  ]
}
```

## Inspiration

This project was inspired by git-hours by Kimmo Brunfeldt
(https://github.com/kimmobrunfeldt/git-hours), but is a clean reimplementation
using the Git CLI instead of nodegit.
