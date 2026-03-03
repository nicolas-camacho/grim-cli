# grim-cli

A terminal reading tracker built with Go and the [Charm](https://charm.sh) suite. Track your books, manga, and comics — reading progress and daily reading habits — all from the command line.

## Features

- Add books or manga/comics with title, progress, and daily reading status
- Books are tracked by pages; manga/comics are tracked by volume or chapter
- List all entries in a styled table with a live reading progress bar
- Log reading sessions and mark entries as completed
- View the full reading session history for any entry
- View detailed information for a book, enriched with author, publish year, and rating from Open Library
- Edit an entry's title or total count interactively
- Delete entries interactively
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

### View session history

Shows the full reading session log for a selected entry. Only entries with at least one session logged appear in the selector.

```bash
grim log
```

| Column | Description |
|---|---|
| Date | Date the session was logged |
| Session | Position range covered in that session (e.g. `120 → 180`) |
| Pages Read / Volumes Read / Chapters Read | Units read in that session |

Sessions are displayed from most recent to oldest.

### Add an entry

Launches an interactive form. First you choose the type, then fill in the details.

```bash
grim add
```

**Book:**

| Field | Description |
|---|---|
| Book title | The name of the book |
| Total pages | Total number of pages in the book |
| Current page | The page you are currently on |
| Did you read it today? | Whether you read it today |

**Manga/Comic:**

| Field | Description |
|---|---|
| Manga/Comic title | The name of the manga or comic |
| Track progress by | Choose **Volume** or **Chapter** |
| Total volumes/chapters | Total number of volumes or chapters |
| Current volume/chapter | The volume or chapter you are currently on |
| Did you read it today? | Whether you read it today |

### List all entries

Displays a table with all tracked entries and their current status.

```bash
grim list
```

The table includes:

| Column | Description |
|---|---|
| Title | Name of the entry |
| Type | `book` or `manga/comic` |
| Status | Current position — `pg.X` for books, `Vol.X` or `Ch.X` for manga/comics |
| Progress | Visual progress bar and percentage |
| Last Read | Date the entry was last read |
| Session | Range of the last reading session (e.g. `120 → 180`) |
| Read Today | Whether the entry was read today |

The **Read Today** status is computed at runtime by comparing the stored date against today's date, so it resets automatically at midnight without modifying any data.

### Log a reading session

Lets you pick an entry from your list, indicate whether you finished it or are still reading, and record your progress.

```bash
grim read
```

| Step | Description |
|---|---|
| What did you read today? | Select an entry from the list |
| How's it going? | Choose **Still reading** or **Completed** |
| What page/volume/chapter did you finish on? | Your stopping point *(only shown when still reading)* |

**Still reading** — records the position you stopped at and updates your progress bar.

**Completed** — automatically sets the current position to the last page/volume/chapter and marks the entry as completed. No manual input needed.

After confirming, a summary is shown with the session range, units read, and updated progress bar. Completed entries display a `★ completed` status.

The **Read Today** column in `grim list` reflects three possible states:

| Status | Meaning |
|---|---|
| `★ completed` | The entry has been fully read |
| `✓ yes` | A session was logged today |
| `✗ not yet` | No session logged today |

### View details

Shows a detailed panel for a selected entry.

```bash
grim dt
```

**Books** — local reading stats combined with live metadata from the [Open Library](https://openlibrary.org/) API:

| Field | Description |
|---|---|
| Title | Name of the book |
| Current page | Current page out of total pages |
| Progress | Visual progress bar and percentage |
| Last session | Page range of the last reading session |
| Pages read | Pages read in the last session |
| Last read | Date last read |
| Added on | Date added to the list |
| Read today | Whether it was read today |
| Author | Author name from Open Library |
| Published | Year of first publication from Open Library |
| Rating | Star rating and total count from Open Library |

**Manga/Comics** — same local stats with labels adapted to volumes or chapters. Open Library is not queried.

| Field | Description |
|---|---|
| Title | Name of the manga/comic |
| Current vol./ch. | Current volume or chapter out of total |
| Progress | Visual progress bar and percentage |
| Last session | Range of the last session |
| Volumes/Chapters | Units read in the last session |
| Last read | Date last read |
| Added on | Date added to the list |
| Read today | Whether it was read today |

**Flags** *(books only)*:

| Flag | Description |
|---|---|
| `-r`, `--refresh` | Force a new Open Library search even if cached metadata already exists |
| `-s`, `--search` | Prompt for a different title to use when searching Open Library. Result is displayed but not saved |

> **Note:** The Open Library lookup requires an internet connection. If no match is found, local data is still displayed normally.

### Edit an entry

Lets you update the title or total count of an existing entry interactively.

```bash
grim modified
```

| Step | Description |
|---|---|
| Which entry do you want to modify? | Select an entry from the list |
| What do you want to change? | Choose **Title** or **Total pages / Total volumes / Total chapters** |
| New value | Enter the new title or the new total count |

### Delete an entry

Launches an interactive selector to pick an entry, then asks for confirmation before deleting.

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
│   ├── log.go         # log command
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
