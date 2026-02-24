```
   ██████╗ █████╗ ███╗   ██╗██╗   ██╗ █████╗ ███████╗
  ██╔════╝██╔══██╗████╗  ██║██║   ██║██╔══██╗██╔════╝
  ██║     ███████║██╔██╗ ██║██║   ██║███████║███████╗
  ██║     ██╔══██║██║╚██╗██║╚██╗ ██╔╝██╔══██║╚════██║
  ╚██████╗██║  ██║██║ ╚████║ ╚████╔╝ ██║  ██║███████║
   ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝  ╚═╝╚══════╝  CLI
```

A command-line client for **Canvas LMS** written in Go. Works with **any Canvas instance** — Instructure-hosted, self-hosted, or institutional deployments.

Manage your courses, assignments, grades, discussions, files, and more — entirely from the terminal.

## Features

- Works with any Canvas LMS instance (API token authentication)
- 20+ commands covering courses, assignments, grades, submissions, modules, discussions, files, calendar, and more
- Color-coded output with human-readable formatting
- `--json` flag on any command for scripting/piping
- Zero external Go dependencies

## Requirements

- Go 1.21+ (for building from source)
- A Canvas LMS account
- A Canvas API access token (generated from Canvas → Account → Settings → New Access Token)

## Installation

```bash
# Clone and build
git clone <repo-url> canvas-cli
cd canvas-cli
go build -o canvas-cli .

# Optional: add to PATH
sudo ln -s $(pwd)/canvas-cli /usr/local/bin/canvas-cli
```

## Quick Start

```bash
# 1. Configure with your Canvas URL and API token
canvas-cli configure
# Canvas URL: https://myschool.instructure.com
# API Token: your_access_token_here

# 2. Verify your identity
canvas-cli whoami

# 3. List your courses
canvas-cli courses

# 4. Check your grades
canvas-cli grades

# 5. See what's due
canvas-cli todo
```

## Getting Your API Token

1. Log into your Canvas instance in a browser
2. Go to **Account** → **Settings**
3. Scroll to **Approved Integrations**
4. Click **+ New Access Token**
5. Give it a purpose (e.g. "canvas-cli") and click **Generate Token**
6. Copy the token — you won't be able to see it again

## Commands

### Setup

| Command | Description |
|---------|-------------|
| `canvas-cli configure` | Set up Canvas URL and API token |
| `canvas-cli whoami` | Show your profile info |
| `canvas-cli version` | Show CLI version |

### Courses

| Command | Description |
|---------|-------------|
| `canvas-cli courses` | List your active courses |
| `canvas-cli courses <id>` | Show course details (syllabus, term, grade) |
| `canvas-cli courses <id> users` | List users enrolled in a course |

### Assignments & Grades

| Command | Description |
|---------|-------------|
| `canvas-cli assignments <course>` | List assignments with due dates and status |
| `canvas-cli assignments <course> <id>` | Show assignment details and your submission |
| `canvas-cli grades` | Grades overview across all courses |
| `canvas-cli grades <course>` | Detailed grades for a specific course |
| `canvas-cli submissions <course> <assign>` | View your submission, comments, and rubric |
| `canvas-cli submit <course> <assign> --text "..."` | Submit text entry |
| `canvas-cli submit <course> <assign> --url <url>` | Submit a URL |

### Productivity

| Command | Description |
|---------|-------------|
| `canvas-cli todo` | Pending to-do items |
| `canvas-cli upcoming` | Upcoming events and assignments |
| `canvas-cli missing` | Missing/late submissions |
| `canvas-cli calendar` | Calendar events (next 30 days) |
| `canvas-cli calendar --start 2026-01-01 --end 2026-02-01` | Events in a date range |

### Content

| Command | Description |
|---------|-------------|
| `canvas-cli modules <course>` | List modules in a course |
| `canvas-cli modules <course> <id>` | List items in a module |
| `canvas-cli discussions <course>` | List discussion topics |
| `canvas-cli discussions <course> <id>` | View a discussion thread |
| `canvas-cli discussions <course> <id> --reply "msg"` | Post a reply |
| `canvas-cli announcements` | Recent announcements (all courses) |
| `canvas-cli announcements <course>` | Announcements for a specific course |

### Files

| Command | Description |
|---------|-------------|
| `canvas-cli files <course>` | List course files with sizes |
| `canvas-cli download <file_id>` | Download a file |
| `canvas-cli download <file_id> -o ~/path` | Download to a specific path |

### Other

| Command | Description |
|---------|-------------|
| `canvas-cli notifications` | Activity stream (announcements, messages, submissions) |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output raw JSON (works on any command) |
| `--per-page <n>` | Results per page, default 50 |
| `-h, --help` | Show help |

## Authentication

This CLI uses Canvas API token authentication via the `Authorization: Bearer <token>` header. Your token is stored locally and never sent anywhere other than your Canvas instance.

## Configuration

Config is stored at `~/.canvas-cli/config.json` with `0600` permissions:

```json
{
  "api_url": "https://myschool.instructure.com",
  "api_token": "your_access_token"
}
```

## Project Structure

```
canvas-cli/
├── main.go                    # Entry point
├── go.mod                     # Go module (zero external deps)
├── cmd/
│   ├── root.go                # Command router, global flags, help
│   ├── announcements.go       # Announcements command
│   ├── assignments.go         # Assignments command
│   ├── calendar.go            # Calendar command
│   ├── courses.go             # Courses command
│   ├── discussions.go         # Discussions command
│   ├── files.go               # Files & download commands
│   ├── grades.go              # Grades command
│   ├── modules.go             # Modules command
│   ├── notifications.go       # Notifications command
│   ├── submissions.go         # Submissions & submit commands
│   ├── todo.go                # Todo, upcoming, missing commands
│   └── whoami.go              # Whoami command
├── internal/
│   ├── api/
│   │   └── client.go          # HTTP client, API token auth
│   ├── config/
│   │   └── config.go          # Configuration load/save
│   └── ui/
│       └── ui.go              # Colors, tables, formatting helpers
├── SKILL.md
└── README.md
```

## License

MIT
