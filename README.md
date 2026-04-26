# git-time

Estimate how much time a developer probably spent working, using only Git
commit history.

`git-time` looks at when commits were made and groups nearby commits into work
sessions. If 2 commits are close together, it assumes the time between them was
active work. If they are far apart, it assumes the developer stopped and later
started a new session.

Because Git only records commits, the result is an estimate. It cannot see time
spent reading code, debugging before the first commit, taking breaks, or working
after the last commit. To account for some of that invisible setup time, each
session gets a fixed offset.

For example, with the default settings:

```text
10:00  commit
11:00  commit
15:00  commit
```

The 10:00 and 11:00 commits are treated as one session: 1 hour between commits
plus a 2-hour session offset, for three estimated hours. The 15:00 commit is
far enough away to start a new session, so it gets another 2-hour offset. The
total estimate is 5 hours.

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
