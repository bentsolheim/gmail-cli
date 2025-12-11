# Gmail CLI - Product Requirements Document

## Overview

A read-only Gmail CLI tool built in Go, designed for agent/LLM consumption. Enables searching Gmail, listing results, and downloading complete email threads with attachments.

## Primary Use Case

Allow an LLM agent (e.g., Claude Code) to search for and retrieve specific emails on behalf of the user. Examples:
- "Find the email from Felipe about conversion factors"
- "Get the meeting notes from earlier this week"

## Core Features

### 1. Search (`gmail-cli search`)
- Pass-through Gmail search syntax (e.g., `from:felipe after:2025/12/10 subject:conversion`)
- Display up to 25 results in numbered list format
- Output format:
  ```
  [1] Dec 11 | Felipe Garcia | Re: Conversion factors (3 messages, 2 attachments)
  [2] Dec 10 | Felipe Garcia | Meeting notes (1 message)
  [3] Dec 9  | Felipe, Sarah | Project update (5 messages, 1 attachment)
  ```

### 2. Download (`gmail-cli download <thread-id>`)
- Download complete thread by thread ID
- Thread content (text/metadata) → stdout
- Attachments → saved to `--output-dir` (required when attachments present)
- Flag: `--no-attachments` to skip attachment download (default: include attachments)

**Stdout format:**
```
Subject: Re: Conversion factors
Participants: felipe@example.com, you@gmail.com
Date Range: Dec 9-11, 2025

--- Message 1 (Dec 9, 10:30 AM) ---
From: Felipe <felipe@example.com>
<body>

--- Message 2 (Dec 9, 2:15 PM) ---
From: You <you@gmail.com>
<body>

Attachments:
- conversion_factors.xlsx (saved to: /path/to/output/conversion_factors.xlsx)
```

### 3. Interactive Mode (`gmail-cli search <query> --interactive`)
- Search → display results → prompt for number → download selected thread
- Combines search and download in one flow

## Authentication

- Google OAuth2 with browser-based consent flow (first-time only)
- Tokens stored in XDG config directory (`~/.config/gmail-cli/`)
- Automatic token refresh
- Gmail API scope: read-only

## Non-Goals (Explicitly Out of Scope)

- Sending emails
- Modifying emails (labels, read status, delete)
- Multiple account support
- Google Workspace admin features

## Technical Decisions

- **Language:** Go (single binary distribution)
- **Gmail API:** Google's official Go client library
- **Config location:** `~/.config/gmail-cli/`

## Commands Summary

| Command | Description |
|---------|-------------|
| `gmail-cli search <query>` | Search and list matching threads |
| `gmail-cli search <query> --interactive` | Search, select, and download in one flow |
| `gmail-cli download <thread-id> --output-dir <path>` | Download thread with attachments |
| `gmail-cli download <thread-id> --no-attachments` | Download thread text only |
| `gmail-cli auth` | Trigger authentication flow (or re-auth) |

## Success Criteria

An agent can:
1. Search for emails using natural query translation to Gmail syntax
2. Review search results and identify the correct thread
3. Download the full thread content and read it into context
4. Access attachment contents when needed
