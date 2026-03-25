---
name: outlookcli
description: Operate Outlook mail and calendar via the outlookcli CLI (Microsoft Graph). Use when the user wants to read/send mail, inbox, search, calendar, scheduling, or run outlookcli commands.
version: 1.0.0
author: ysu03zyy
---

# outlookcli Skill

Drive Outlook / Microsoft 365 mail and calendar through the **`outlookcli`** command-line tool (same Graph permissions as the Outlook MCP setup: `~/.outlook-mcp`).

## When to use

- User mentions **outlookcli**, **Outlook CLI**, or wants **mail/calendar** automation in the terminal.
- Prefer **`outlookcli ...`** over raw `curl` to Microsoft Graph.

## Prerequisites

- Binary **`outlookcli`** on `PATH` (see project `README.md`: `go install`, Homebrew tap, or build from source).
- **OAuth files** (unless using `--access-token` only):
  - `~/.outlook-mcp/config.json` — `client_id`, `client_secret`
  - `~/.outlook-mcp/credentials.json` — refresh / access tokens (after setup)

Initial app registration and consent can follow the same Azure steps as the Outlook skill; see `references/setup.md`.

## Global flags (before subcommands)

| Flag / env | Purpose |
|------------|---------|
| `--config-dir` / `OUTLOOK_CONFIG_DIR` | Override config directory (default `~/.outlook-mcp`) |
| `--access-token` / `OUTLOOK_ACCESS_TOKEN` | Use a Graph access token directly (multi-user; **no** auto-refresh from files; ~1h lifetime) |
| `-j` / `--json` | Machine-readable JSON on stdout |
| `-z` / `--timezone` / `OUTLOOK_TIMEZONE` | IANA timezone for calendar (e.g. `Asia/Shanghai`) |

Example:

```bash
outlookcli -j --timezone Asia/Shanghai mail inbox -n 5
```

## Token

```bash
outlookcli token test              # connectivity / inbox counts
outlookcli token refresh           # force refresh (needs config files)
outlookcli token get               # print token (from OAuth flow or echo --access-token)
```

With **`--access-token`**, do **not** use `token refresh` (it requires `config.json`); omit the flag to use file-based OAuth.

## Mail

| Action | Command |
|--------|---------|
| Latest inbox | `outlookcli mail inbox [-n COUNT]` |
| Unread | `outlookcli mail unread [-n COUNT]` |
| Search | `outlookcli mail search "query" [-n COUNT]` |
| From sender | `outlookcli mail from sender@example.com [-n COUNT]` |
| Read body | `outlookcli mail read <id>` |
| Folders | `outlookcli mail folders` |
| Inbox stats | `outlookcli mail stats` |
| Send | `outlookcli mail send -t TO -s "Subject" -b "Body"` |
| Reply | `outlookcli mail reply <id> "body text"` |
| Mark read / unread | `outlookcli mail mark-read <id>` / `mail mark-unread <id>` |
| Flag / unflag | `outlookcli mail flag <id>` / `mail unflag <id>` |
| Trash / archive | `outlookcli mail delete <id>` / `mail archive <id>` |
| Move folder | `outlookcli mail move <id> <folder-name>` |

**Message IDs:** list output shows a **short id** (suffix of the Graph id). Use that same value with `read`, `reply`, `move`, etc.

## Calendar (`calendar` or `cal`)

| Action | Command |
|--------|---------|
| Upcoming events | `outlookcli calendar events [-n COUNT]` |
| Today | `outlookcli calendar today` |
| This week | `outlookcli calendar week` |
| Event detail | `outlookcli calendar read <id>` |
| Create | `outlookcli calendar create "Subject" "2026-03-25T10:00" "2026-03-25T11:00" [-l "Location"]` |
| Quick 1h event | `outlookcli calendar quick "Subject" [-s START]` |
| Update field | `outlookcli calendar update <id> <field> <value>` — field: `subject`, `location`, `start`, or `end` |
| Delete | `outlookcli calendar delete <id>` |
| List calendars | `outlookcli calendar calendars` |
| Free/busy | `outlookcli calendar free <start> <end>` |

**Datetime format:** `YYYY-MM-DDTHH:MM` (local interpretation uses `--timezone` / `OUTLOOK_TIMEZONE`).

## Multi-user / CI (token only)

```bash
export OUTLOOK_ACCESS_TOKEN="eyJ..."
outlookcli -j mail inbox -n 3
```

No `credentials.json` refresh; renew the token externally when it expires.

## JSON output

Add **`-j`** for stable JSON (e.g. automation, agents). Plain text mode prints one JSON object per line for list-style results.

## Troubleshooting

| Symptom | Suggestion |
|---------|------------|
| Auth / refresh errors | Run `outlookcli token test`; check `~/.outlook-mcp` files and Azure app permissions |
| `401` / expired | File mode: run `outlookcli token refresh` or rely on automatic refresh; token-only mode: supply a new `--access-token` |
| Message not found | Id may be outside the recent window used for resolution; try `mail search` first |

## Changelog

### 1.0.0

- Initial skill for `outlookcli` (token, mail, calendar, global flags).
