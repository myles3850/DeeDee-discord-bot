# googleplatform

Google Sheets client. Provides read access to spreadsheet data via a service account.

## Setup

Two environment variables are required:

| Variable | Description |
|---|---|
| `GOOGLE_CREDENTIALS_JSON` | Full service account credentials JSON (not base64 — pass the raw JSON string) |
| `GOOGLE_SPREADSHEET_ID` | The spreadsheet ID from the Google Sheets URL: `https://docs.google.com/spreadsheets/d/<ID>/edit` |

The service account must have at least **Viewer** access to the target spreadsheet. Share the spreadsheet with the service account email from your credentials JSON.

## Usage

`NewSheet(ctx)` is called once at startup and the resulting `*Sheet` is passed into the Discord bot, stored on `Discord.Sheet`. Use `GetWorksheet` to read a tab by name:

```go
rows, err := sheet.GetWorksheet(ctx, "Sheet1")
// rows is [][]interface{} — each inner slice is one row, each element is a cell value
```

The sheet name must match the tab name exactly (case-sensitive).

## Status

`Discord.Sheet` isn't read by any handler yet — this client is wired up ahead of an unfinished feature (a "fun facts" type command, going by the root README's tagline). It's not dead code; the plumbing (env vars, startup fetch, struct field) is intentionally in place for whichever handler ends up using it. If you're picking this up: write a `GetWorksheet` call in a new file under `discord/`, wire it into a slash command, and this note can go.
