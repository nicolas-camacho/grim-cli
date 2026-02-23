# grim-cli

A terminal reading tracker built with Go and the [Charm](https://charm.sh) suite. Track your books, reading progress, and daily reading habits — all from the command line.

## Features

- Add books with title, total pages, current page, and daily reading status
- List all books in a styled table with a live reading progress bar
- Log reading sessions and mark books as completed
- View detailed information for a book, enriched with author, publish year, and rating from Open Library
- Edit a book's title or total pages interactively
- Delete books interactively
- Data persists locally in `~/.grim/books.json`

## Requirements

- [Go](https://go.dev/) 1.22 or later

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/nicolas-camacho/grim-cli
cd grim-cli
```

Then install it depending on your shell:

**PowerShell**
```powershell
go build -o $HOME\go\bin\grim.exe .
```

**CMD**
```cmd
go build -o %USERPROFILE%\go\bin\grim.exe .
```

**Git Bash**
```bash
go build -o $HOME/go/bin/grim.exe .
```

> **Note:** Avoid using `~` in the output path on Windows — it is not expanded automatically outside of Git Bash and will create a literal `~` folder inside your project directory.

Make sure `%USERPROFILE%\go\bin` is in your `PATH`. You can verify with:

```bash
grim --help
```

## Updating

After pulling new changes, rebuild and reinstall using the same command for your shell:

**PowerShell**
```powershell
git pull
go build -o $HOME\go\bin\grim.exe .
```

**CMD**
```cmd
git pull
go build -o %USERPROFILE%\go\bin\grim.exe .
```

**Git Bash**
```bash
git pull
go build -o $HOME/go/bin/grim.exe .
```

## Usage

### Shortcuts

Most commands have a short alias:

| Alias | Command |
|---|---|
| `grim a` | `grim add` |
| `grim ls` | `grim list` |
| `grim rd` | `grim read` |
| `grim mod` | `grim modified` |

### Add a book

Launches an interactive form that asks for the book details sequentially.

```bash
grim add
```

You will be prompted for:

| Field | Description |
|---|---|
| Book title | The name of the book |
| Total pages | Total number of pages in the book |
| Current page | The page you are currently on |
| Did you read it today? | Whether you read the book today |

### List all books

Displays a table with all tracked books and their current status.

```bash
grim list
```

The table includes:

| Column | Description |
|---|---|
| Title | Name of the book |
| Page | Current page |
| Progress | Visual progress bar and percentage |
| Last Read | Date the book was last read |
| Session | Page range of the last reading session (e.g. `120 → 180`) |
| Pages Read | Number of pages read in the last session |
| Read Today | Whether the book was read today |

The **Read Today** status is computed at runtime by comparing the stored date against today's date, so it resets automatically at midnight without modifying any data.

### Log a reading session

Lets you pick a book from your list, indicate whether you finished it or are still reading, and record your progress. The previous page and current page are updated automatically, and the book is marked as read today.

```bash
grim read
```

You will be prompted for:

| Step | Description |
|---|---|
| Which book did you read today? | Select a book from the list |
| How's it going? | Choose **Still reading** or **Completed** |
| What page did you finish on? | The page you stopped at *(only shown when still reading)* |

**Still reading** — records the page you stopped at and updates your progress bar.

**Completed** — automatically sets the current page to the last page and marks the book as completed. No manual page entry needed.

After confirming, a summary is shown with the session range, pages read, and updated progress bar. Completed books also display a `★ completed` status.

The **Read Today** column in `grim list` reflects three possible states:

| Status | Meaning |
|---|---|
| `★ completed` | The book has been fully read |
| `✓ yes` | A session was logged today |
| `✗ not yet` | No session logged today |

### View book details

Shows a detailed panel for a selected book. Local reading stats are combined with live metadata fetched from the [Open Library](https://openlibrary.org/) API.

```bash
grim dt
```

You will be prompted to select a book from your list. The detail panel includes:

| Field | Description |
|---|---|
| Title | Name of the book |
| Current page | Current page out of total pages |
| Progress | Visual progress bar and percentage |
| Last session | Page range of the last reading session (e.g. `95 → 120`) |
| Pages read | Pages read in the last session |
| Last read | Date the book was last read |
| Added on | Date the book was added to the list |
| Read today | Whether the book was read today |
| Author | Author name from Open Library |
| Published | Year of first publication from Open Library |
| Rating | Star rating and total count from Open Library |

**Flags:**

| Flag | Description |
|---|---|
| `-r`, `--refresh` | Force a new Open Library search even if cached metadata already exists. Updates stored metadata with the new result. |
| `-s`, `--search` | Prompt for a different title to use when searching Open Library. Useful when the stored title does not match what Open Library expects. The result is displayed but not saved. |

> **Note:** The Open Library lookup requires an internet connection. If no match is found, local data is still displayed normally.

### Edit a book

Lets you update the title or total pages of an existing book interactively.

```bash
grim modified
```

You will be prompted to:

| Step | Description |
|---|---|
| Which book do you want to modify? | Select a book from the list |
| What do you want to change? | Choose **Title** or **Total pages** |
| New value | Enter the new title or the new total page count |

After confirming, a summary is shown with the updated field.

### Delete a book

Launches an interactive selector to pick a book, then asks for confirmation before deleting.

```bash
grim del
```

### Print version

```bash
grim version
```

## Running tests

Run all tests across every package:

```bash
go test ./...
```

Run with verbose output to see each test case individually:

```bash
go test ./... -v
```

Run only a specific package:

```bash
go test ./store/...
go test ./cmd/...
```

Run a single test by name:

```bash
go test ./store/... -run TestDeleteBook_RemovesCorrectBook
```

Tests are fully isolated — they use temporary directories and never touch `~/.grim`.

## Data storage

All data is stored in a single JSON file:

```
~/.grim/books.json
```

You can back it up, copy it to another machine, or inspect it manually at any time.

## Project structure

```
grim-cli/
├── main.go          # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go      # Root Cobra command and Execute() function
│   ├── book.go        # add, list, del, read, dt, and modified commands
│   ├── openlibrary.go # Open Library API client
│   └── version.go     # version command
├── store/
│   └── store.go     # Data model and JSON persistence layer
└── ui/
    └── styles.go    # Lip Gloss styles shared across all commands
```

## Tech stack

| Library | Purpose |
|---|---|
| [Cobra](https://github.com/spf13/cobra) | CLI commands and flags |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [Huh](https://github.com/charmbracelet/huh) | Interactive forms |
| [Charm Log](https://github.com/charmbracelet/log) | Structured logging |
