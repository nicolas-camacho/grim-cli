# grim-cli

A terminal reading tracker built with Go and the [Charm](https://charm.sh) suite. Track your books, reading progress, and daily reading habits — all from the command line.

## Features

- Add books with title, total pages, current page, and daily reading status
- List all books in a styled table with a live reading progress bar
- Delete books interactively
- Data persists locally in `~/.grim/books.json`

## Requirements

- [Go](https://go.dev/) 1.22 or later

## Installation

Clone the repository and install the binary:

```bash
git clone https://github.com/nicolas/grim-cli
cd grim-cli
go build -o ~/go/bin/grim.exe .
```

Make sure `~/go/bin` is in your `PATH`. You can verify with:

```bash
grim --help
```

## Updating

After pulling new changes, rebuild and reinstall with the same command:

```bash
git pull
go build -o ~/go/bin/grim.exe .
```

## Usage

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
│   ├── book.go      # add, list, and del commands
│   └── version.go   # version command
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
