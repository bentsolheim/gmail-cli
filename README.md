# gmail-cli

A read-only Gmail CLI tool designed for agent/LLM consumption. Search Gmail, list results, and download complete email threads with attachments.

## Installation

```bash
go install github.com/bentsolheim/gmail-cli/cmd/gmail-cli@latest
```

Or build from source:

```bash
git clone https://github.com/bentsolheim/gmail-cli.git
cd gmail-cli
make install
```

## Setup

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select existing)
3. Enable the **Gmail API**:
   - Go to APIs & Services > Library
   - Search for "Gmail API"
   - Click Enable

### 2. Create OAuth Credentials

1. Go to APIs & Services > Credentials
2. Click "Create Credentials" > "OAuth client ID"
3. Select "Desktop app" as the application type
4. Name it (e.g., "gmail-cli")
5. Download the JSON file

### 3. Configure gmail-cli

```bash
mkdir -p ~/.config/gmail-cli
mv ~/Downloads/client_secret_*.json ~/.config/gmail-cli/credentials.json
```

### 4. Authenticate

```bash
gmail-cli auth
```

This opens a browser for Google authorization. After authorizing, the token is saved to `~/.config/gmail-cli/token.json`.

## Usage

### Search for emails

```bash
# Basic search
gmail-cli search "from:felipe subject:conversion"

# With Gmail search operators
gmail-cli search "after:2025/12/01 has:attachment"
gmail-cli search "is:unread from:me"
gmail-cli search "label:important newer_than:7d"
```

Output:
```
[1] Dec 11 | Felipe Garcia | Re: Conversion factors (3 messages, 2 attachments)
[2] Dec 10 | Felipe Garcia | Meeting notes (1 message)
[3] Dec 9  | Felipe, Sarah | Project update (5 messages, 1 attachment)
```

### Interactive search and download

```bash
gmail-cli search "from:felipe" --interactive
```

Search results are displayed, then you're prompted to select a thread to download.

### Download a thread

```bash
# Download with attachments
gmail-cli download <thread-id> --output-dir ./emails

# Download text only
gmail-cli download <thread-id> --no-attachments
```

Output:
```
Subject: Re: Conversion factors
Participants: felipe@example.com, you@gmail.com
Date Range: Dec 9-11, 2025

--- Message 1 (Dec 9, 10:30 AM) ---
From: Felipe <felipe@example.com>
Hey, here are the conversion factors...

--- Message 2 (Dec 9, 2:15 PM) ---
From: You <you@gmail.com>
Thanks! I'll review these.

Attachments:
- conversion_factors.xlsx (saved to: ./emails/conversion_factors.xlsx)
```

### Re-authenticate

```bash
gmail-cli auth
```

## Gmail Search Syntax

gmail-cli uses Gmail's native search syntax. Common operators:

| Operator | Example | Description |
|----------|---------|-------------|
| `from:` | `from:felipe` | Sender |
| `to:` | `to:me` | Recipient |
| `subject:` | `subject:meeting` | Subject line |
| `has:attachment` | `has:attachment` | Has attachments |
| `after:` | `after:2025/12/01` | After date |
| `before:` | `before:2025/12/31` | Before date |
| `newer_than:` | `newer_than:7d` | Within last N days |
| `older_than:` | `older_than:1m` | Older than N months |
| `is:unread` | `is:unread` | Unread messages |
| `label:` | `label:important` | Has label |

Combine operators: `from:felipe after:2025/12/01 has:attachment`

## Commands

| Command | Description |
|---------|-------------|
| `gmail-cli auth` | Authenticate with Gmail |
| `gmail-cli search <query>` | Search threads (up to 25 results) |
| `gmail-cli search <query> -i` | Interactive: search, select, download |
| `gmail-cli download <id> -o <dir>` | Download thread with attachments |
| `gmail-cli download <id> --no-attachments` | Download thread text only |

## Configuration

Configuration is stored in `~/.config/gmail-cli/`:

- `credentials.json` - OAuth client credentials (you provide)
- `token.json` - OAuth access/refresh tokens (auto-generated)

The tool uses read-only Gmail API scope (`gmail.readonly`).

## License

MIT
