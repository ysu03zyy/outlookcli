# Outlook / Microsoft Graph setup (for outlookcli)

`outlookcli` expects the same OAuth application and token storage layout as the **Outlook MCP** flow:

- **Directory:** `~/.outlook-mcp/` (or `OUTLOOK_CONFIG_DIR`)
- **`config.json`:** `client_id`, `client_secret` from an Azure App Registration
- **`credentials.json`:** OAuth tokens (including `refresh_token` for file-based mode)

## Options to obtain credentials

1. **Azure CLI automation** — use a community `outlook-setup.sh`-style script (see original Outlook skill) to create the app and save files under `~/.outlook-mcp/`.
2. **Azure Portal (manual)** — register an app, add Microsoft Graph delegated permissions (`Mail.ReadWrite`, `Mail.Send`, `Calendars.ReadWrite`, `offline_access`, `User.Read` as needed), create a client secret, set redirect URI, complete consent, then place tokens in `credentials.json` using your OAuth flow or device code.

## Permissions (typical)

- `Mail.ReadWrite`, `Mail.Send`, `Calendars.ReadWrite`, `offline_access`

## Repository

Project home: `https://github.com/ysu03zyy/outlookcli` — install and full docs in `README.md`.
