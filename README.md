```
   ██████╗ █████╗ ███╗   ██╗██╗   ██╗ █████╗ ███████╗
  ██╔════╝██╔══██╗████╗  ██║██║   ██║██╔══██╗██╔════╝
  ██║     ███████║██╔██╗ ██║██║   ██║███████║███████╗
  ██║     ██╔══██║██║╚██╗██║╚██╗ ██╔╝██╔══██║╚════██║
  ╚██████╗██║  ██║██║ ╚████║ ╚████╔╝ ██║  ██║███████║
   ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝  ╚═╝╚══════╝  CLI
```

A command-line client for **Canvas LMS (Experiencia21)** written in Go, built specifically for **Tecnologico de Monterrey** students.

I'm a student at Tec de Monterrey and I built this so I could access Canvas (Experiencia21) directly from [OpenClaw](https://openclaw.com) and the terminal, without needing to open a browser.

> **Important:** This version only works for Tec de Monterrey students using `experiencia21.tec.mx`. It handles the full Tec SAML SSO chain (amfs.tec.mx + aamfa.tec.mx + TOTP). If you're looking for a universal Canvas CLI that works with any institution, see [canvas-cli-general](https://github.com/jpnarchi/canvas-cli-general).

## Features

- Full Tec de Monterrey SAML SSO authentication (amfs.tec.mx IdP)
- Automatic device fingerprinting (JWS signed + JWE encrypted)
- TOTP/MFA support — prompts for your authenticator code
- Session caching — login once, reuse until session expires (no TOTP on every run)
- 20+ commands covering courses, assignments, grades, submissions, modules, discussions, files, calendar, and more
- Color-coded output with human-readable formatting
- `--json` flag on any command for scripting/piping
- Zero external Go dependencies

## Requirements

- Go 1.21+ (for building from source)
- A Tec de Monterrey student account (e.g., `a01234567@tec.mx`)
- Your TOTP authenticator app configured for Tec de Monterrey MFA

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
# 1. Configure with your Tec credentials
canvas-cli configure
# Canvas URL: https://experiencia21.tec.mx
# Username: a01234567@tec.mx
# Password: your password

# 2. Verify login (will prompt for TOTP code on first run)
canvas-cli whoami

# 3. List your courses
canvas-cli courses

# 4. Check your grades
canvas-cli grades

# 5. See what's due
canvas-cli todo
```

After the first successful login, your session is cached — subsequent commands won't ask for your TOTP code again until the session expires.

## Commands

### Setup

| Command | Description |
|---------|-------------|
| `canvas-cli configure` | Set up Canvas URL, username, and password |
| `canvas-cli whoami` | Show your profile info |
| `canvas-cli debug-login` | Test login flow with verbose output |
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

## Authentication Flow (Tec de Monterrey)

This CLI handles the full Tec SAML SSO chain automatically:

1. `experiencia21.tec.mx/login` redirects via SAML to `amfs.tec.mx` (NetIQ IdP)
2. Auto-submit intermediate form to load the credential page
3. Credentials submitted with Base64-encoded password (`itesm64` field)
4. Device fingerprinting handled (JWS HS256 signed + JWE A128CBC-HS256 encrypted)
5. OAuth2 redirect to `aamfa.tec.mx` for MFA
6. **TOTP code prompted** — enter from your authenticator app
7. Consent form auto-submitted
8. JavaScript redirect followed to generate SAMLResponse
9. SAMLResponse posted back to Canvas → session established
10. Session cookies saved to `~/.canvas-cli/config.json` for reuse

On subsequent runs, saved cookies are tested first. If still valid, no login is needed.

## Configuration

Config is stored at `~/.canvas-cli/config.json` with `0600` permissions:

```json
{
  "api_url": "https://experiencia21.tec.mx",
  "username": "a01234567@tec.mx",
  "password": "yourpassword",
  "cookies": [...]
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
│   ├── debug.go               # Debug login command
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
│   │   └── client.go          # HTTP client, SAML SSO, session management
│   ├── config/
│   │   └── config.go          # Configuration load/save
│   └── ui/
│       └── ui.go              # Colors, tables, formatting helpers
├── SKILL.md
└── README.md
```

## Canvas API Documentation

This CLI is built on top of the [Canvas LMS REST API](https://developerdocs.instructure.com/services/canvas).

## License

MIT
