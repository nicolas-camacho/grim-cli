package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	xterm "github.com/charmbracelet/x/term"
	"github.com/nicolas-camacho/grim-cli/store"
	"github.com/nicolas-camacho/grim-cli/ui"
	"github.com/spf13/cobra"
)

// bookAddCmd asks whether the user wants to add a book or a manga, then
// collects the appropriate details and persists the new entry via the store.
var bookAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a book, manga or comic to your reading list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// numValidate is reused for all numeric inputs.
		numValidate := func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be a number")
			}
			return nil
		}

		// Step 1: choose entry type.
		var entryType string
		typeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What do you want to add?").
					Options(
						huh.NewOption("Book", "book"),
						huh.NewOption("Manga/Comic", "manga"),
					).
					Value(&entryType),
			),
		)
		if err := typeForm.Run(); err != nil {
			return err
		}

		s, err := store.New()
		if err != nil {
			return err
		}

		if entryType == "book" {
			var title, pageStr, totalPagesStr string
			var readToday bool

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Book title").
						Placeholder("The Go Programming Language").
						Value(&title),
					huh.NewInput().
						Title("Total pages").
						Placeholder("350").
						Validate(numValidate).
						Value(&totalPagesStr),
					huh.NewInput().
						Title("Current page").
						Placeholder("0").
						Validate(numValidate).
						Value(&pageStr),
					huh.NewConfirm().
						Title("Did you read it today?").
						Value(&readToday),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}

			page, _ := strconv.Atoi(pageStr)
			totalPages, _ := strconv.Atoi(totalPagesStr)

			if err := s.AddBook(title, page, totalPages, readToday); err != nil {
				return err
			}

			readStatus := ui.Danger.Render("✗ not yet")
			if readToday {
				readStatus = ui.Success.Render("✓ yes")
			}

			fmt.Println(ui.Box.Render(
				ui.Title.Render("Book added") + "\n\n" +
					ui.Bold.Render("Title:        ") + title + "\n" +
					ui.Bold.Render("Current page: ") + fmt.Sprintf("%d", page) + "\n" +
					ui.Bold.Render("Read today:   ") + readStatus,
			))
			return nil
		}

		// Manga flow — two sequential forms so the second can use the chosen unit label.
		var title, trackingUnit string
		metaForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Manga/Comic title").
					Placeholder("One Piece").
					Value(&title),
				huh.NewSelect[string]().
					Title("Track progress by").
					Options(
						huh.NewOption("Volume", "volume"),
						huh.NewOption("Chapter", "chapter"),
					).
					Value(&trackingUnit),
			),
		)
		if err := metaForm.Run(); err != nil {
			return err
		}

		unitSingular := "volume"
		if trackingUnit == "chapter" {
			unitSingular = "chapter"
		}

		var totalStr, currentStr string
		var readToday bool
		progressForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Total %ss", unitSingular)).
					Placeholder("100").
					Validate(numValidate).
					Value(&totalStr),
				huh.NewInput().
					Title(fmt.Sprintf("Current %s", unitSingular)).
					Placeholder("0").
					Validate(numValidate).
					Value(&currentStr),
				huh.NewConfirm().
					Title("Did you read it today?").
					Value(&readToday),
			),
		)
		if err := progressForm.Run(); err != nil {
			return err
		}

		total, _ := strconv.Atoi(totalStr)
		current, _ := strconv.Atoi(currentStr)

		if err := s.AddMangaComic(title, trackingUnit, current, total, readToday); err != nil {
			return err
		}

		readStatus := ui.Danger.Render("✗ not yet")
		if readToday {
			readStatus = ui.Success.Render("✓ yes")
		}

		capUnit := strings.ToUpper(unitSingular[:1]) + unitSingular[1:]
		fmt.Println(ui.Box.Render(
			ui.Title.Render("Manga added") + "\n\n" +
				ui.Bold.Render("Title:          ") + title + "\n" +
				ui.Bold.Render(fmt.Sprintf("Current %-8s", capUnit+":")) + fmt.Sprintf("%d", current) + "\n" +
				ui.Bold.Render("Read today:     ") + readStatus,
		))
		return nil
	},
}

// mangaComicUnitLabel returns "Vol." or "Ch." for a manga entry, or "" for a book.
func mangaComicUnitLabel(b store.Book) string {
	if !b.IsMangaComic {
		return ""
	}
	if b.TrackingUnit == "chapter" {
		return "Ch."
	}
	return "Vol."
}

// currentPosition formats the current progress value for display.
// Books show "pg. X"; manga show "Vol. 3" or "Ch. 25".
func currentPosition(b store.Book) string {
	if b.IsMangaComic {
		return fmt.Sprintf("%s%d", mangaComicUnitLabel(b), b.CurrentPage)
	}
	return fmt.Sprintf("pg.%d", b.CurrentPage)
}

// progressBar renders a 10-character block bar followed by a percentage.
// Returns a muted dash when total is 0 to avoid division by zero.
//
// Example output:  ████░░░░░░ 40%
func progressBar(current, total int) string {
	if total == 0 {
		return ui.Muted.Render("—")
	}
	pct := int(float64(current) / float64(total) * 100)
	pct = min(pct, 100)
	const width = 10
	filled := width * pct / 100
	bar := ui.Title.Render(repeatStr("█", filled)) + ui.Muted.Render(repeatStr("░", width-filled))
	return fmt.Sprintf("%s %d%%", bar, pct)
}

// truncateTitle shortens s to at most maxWidth visible columns, appending "..."
// when truncation occurs. Handles multi-byte Unicode correctly.
func truncateTitle(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return "..."
	}
	limit := maxWidth - 3
	width := 0
	var buf strings.Builder
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if width+rw > limit {
			buf.WriteString("...")
			return buf.String()
		}
		buf.WriteRune(r)
		width += rw
	}
	return s
}

// repeatStr returns s repeated n times.
func repeatStr(s string, n int) string {
	var sb strings.Builder
	for range n {
		sb.WriteString(s)
	}
	return sb.String()
}

// bookListCmd loads all entries from the store and renders them in a styled table.
// The "Read Today" column is evaluated live so it resets at midnight automatically.
var bookListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all books and manga and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No entries yet. Add one with: grim add"))
			return nil
		}

		headers := []string{"Title", "Type", "Status", "Progress", "Last Read", "Session", "Read Today"}
		rows := make([][]string, len(s.Books))
		for i, b := range s.Books {
			readStatus := ui.Danger.Render("✗ not yet")
			if b.Completed {
				readStatus = ui.Warning.Render("★ completed")
			} else if b.WasReadToday() {
				readStatus = ui.Success.Render("✓ yes")
			}

			lastRead := ui.Muted.Render("never")
			if b.LastReadDate != "" {
				lastRead = b.LastReadDate
			}

			// Session is only meaningful when the entry has been read at least once.
			session := ui.Muted.Render("—")
			if b.LastReadDate != "" {
				session = fmt.Sprintf("%d → %d", b.PreviousPage, b.CurrentPage)
			}

			entryType := ui.Muted.Render("book")
			if b.IsMangaComic {
				entryType = ui.Muted.Render("manga/comic")
			}

			rows[i] = []string{
				b.Title,
				entryType,
				currentPosition(b),
				progressBar(b.CurrentPage, b.TotalPages),
				lastRead,
				session,
				readStatus,
			}
		}

		buildTable := func(r [][]string) *table.Table {
			return table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#A673FF"))).
				Headers(headers...).
				StyleFunc(func(row, col int) lipgloss.Style {
					base := lipgloss.NewStyle()
					if col == 0 {
						base = base.PaddingLeft(2).PaddingRight(2)
					}
					if row == table.HeaderRow {
						return base.Bold(true).Foreground(lipgloss.Color("#A673FF"))
					}
					// Alternate row shading for readability.
					if row%2 == 0 {
						return base.Foreground(lipgloss.Color("#E5E7EB"))
					}
					return base
				}).
				Rows(r...)
		}

		t := buildTable(rows)

		termWidth, _, err := xterm.GetSize(os.Stdout.Fd())
		if err == nil && termWidth > 0 {
			tableWidth := lipgloss.Width(t.String())
			if tableWidth > termWidth {
				// Find the longest title to know the current title column width.
				// The column = max(longestTitle, len("Title")) + 4 padding.
				maxTitleContent := lipgloss.Width("Title")
				for _, row := range rows {
					if w := lipgloss.Width(row[0]); w > maxTitleContent {
						maxTitleContent = w
					}
				}
				titleColWidth := maxTitleContent + 4
				fixedWidth := tableWidth - titleColWidth

				// Space available for title text inside the column.
				availableContent := max(termWidth-fixedWidth-4, 3)

				// Check each title individually and truncate only those that exceed the limit.
				for i, row := range rows {
					if lipgloss.Width(row[0]) > availableContent {
						rows[i][0] = truncateTitle(row[0], availableContent)
					}
				}
				t = buildTable(rows)
			}
		}

		fmt.Println(ui.Title.Render("Reading List"))
		fmt.Println(t)
		return nil
	},
}

// bookDeleteCmd presents an interactive selector of all entries, then asks for
// confirmation before removing the chosen entry from the store.
var bookDeleteCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete a book or manga/comic from your reading list",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No entries yet. Add one with: grim add"))
			return nil
		}

		// Build the option list from current entries.
		options := make([]huh.Option[string], len(s.Books))
		for i, b := range s.Books {
			options[i] = huh.NewOption(b.Title, b.Title)
		}

		var selected string
		var confirmed bool

		// Step 1: pick an entry.
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which entry do you want to delete?").
					Options(options...).
					Value(&selected),
			),
		)

		if err := selectForm.Run(); err != nil {
			return err
		}

		// Step 2: confirm with the title in the message.
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Are you sure you want to delete %q?", selected)).
					Value(&confirmed),
			),
		)

		if err := confirmForm.Run(); err != nil {
			return err
		}

		if !confirmed {
			fmt.Println(ui.Muted.Render("Cancelled. No entries were deleted."))
			return nil
		}

		if err := s.DeleteBook(selected); err != nil {
			return err
		}

		fmt.Println(ui.Success.Render("✓ ") + ui.Bold.Render(selected) + " removed from your reading list.")
		return nil
	},
}

// bookReadCmd lets the user mark a reading session for today. It presents
// a selector to pick an entry, then asks for the position they stopped at
// (page for books, volume or chapter for manga).
var bookReadCmd = &cobra.Command{
	Use:     "read",
	Aliases: []string{"rd"},
	Short:   "Log today's reading session",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No entries yet. Add one with: grim add"))
			return nil
		}

		options := make([]huh.Option[string], len(s.Books))
		for i, b := range s.Books {
			options[i] = huh.NewOption(b.Title, b.Title)
		}

		var selected string
		var completed bool

		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What did you read today?").
					Options(options...).
					Value(&selected),
			),
		)

		if err := selectForm.Run(); err != nil {
			return err
		}

		// Find the selected entry to adapt labels and prompts.
		var book store.Book
		for _, b := range s.Books {
			if b.Title == selected {
				book = b
				break
			}
		}

		statusForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[bool]().
					Title("How's it going?").
					Options(
						huh.NewOption("Still reading", false),
						huh.NewOption("Completed", true),
					).
					Value(&completed),
			),
		)

		if err := statusForm.Run(); err != nil {
			return err
		}

		if completed {
			if err := s.CompleteBook(selected); err != nil {
				return err
			}

			unitsRead := book.TotalPages - book.CurrentPage
			unitsLabel := "Pages read:  "
			if book.IsMangaComic {
				if book.TrackingUnit == "chapter" {
					unitsLabel = "Chapters:    "
				} else {
					unitsLabel = "Volumes:     "
				}
			}

			fmt.Println(ui.Box.Render(
				ui.Title.Render("Session logged") + "\n\n" +
					ui.Bold.Render("Title:       ") + selected + "\n" +
					ui.Bold.Render("Session:     ") + fmt.Sprintf("%d → %d", book.CurrentPage, book.TotalPages) + "\n" +
					ui.Bold.Render(unitsLabel) + ui.Success.Render(fmt.Sprintf("%d", unitsRead)) + "\n" +
					ui.Bold.Render("Progress:    ") + progressBar(book.TotalPages, book.TotalPages) + "\n" +
					ui.Bold.Render("Status:      ") + ui.Warning.Render("★ completed"),
			))

			return nil
		}

		// Build prompt label based on entry type.
		posPrompt := "What page did you finish on?"
		if book.IsMangaComic {
			if book.TrackingUnit == "chapter" {
				posPrompt = "What chapter did you finish on?"
			} else {
				posPrompt = "What volume did you finish on?"
			}
		}

		var pageStr string
		pageForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(posPrompt).
					Placeholder("0").
					Validate(func(s string) error {
						if _, err := strconv.Atoi(s); err != nil {
							return fmt.Errorf("must be a number")
						}
						return nil
					}).
					Value(&pageStr),
			),
		)

		if err := pageForm.Run(); err != nil {
			return err
		}

		newPage, _ := strconv.Atoi(pageStr)

		if err := s.UpdateBook(selected, newPage); err != nil {
			return err
		}

		// Find the updated entry to show session info in the confirmation.
		var updated store.Book
		for _, b := range s.Books {
			if b.Title == selected {
				updated = b
				break
			}
		}

		unitsRead := newPage - updated.PreviousPage
		unitsLabel := "Pages read:  "
		if book.IsMangaComic {
			if book.TrackingUnit == "chapter" {
				unitsLabel = "Chapters:    "
			} else {
				unitsLabel = "Volumes:     "
			}
		}

		fmt.Println(ui.Box.Render(
			ui.Title.Render("Session logged") + "\n\n" +
				ui.Bold.Render("Title:       ") + selected + "\n" +
				ui.Bold.Render("Session:     ") + fmt.Sprintf("%d → %d", updated.PreviousPage, newPage) + "\n" +
				ui.Bold.Render(unitsLabel) + ui.Success.Render(fmt.Sprintf("%d", unitsRead)) + "\n" +
				ui.Bold.Render("Progress:    ") + progressBar(newPage, updated.TotalPages),
		))

		return nil
	},
}

// bookDetailCmd shows an entry selector, fetches enriched metadata from Open
// Library (books only), and renders a combined local + remote detail panel.
// The --refresh flag forces a new Open Library search even when cached metadata exists.
var bookDetailCmd = &cobra.Command{
	Use:   "dt",
	Short: "Show detailed information for a book or manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		refresh, _ := cmd.Flags().GetBool("refresh")
		useCustomSearch, _ := cmd.Flags().GetBool("search")
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No entries yet. Add one with: grim add"))
			return nil
		}

		options := make([]huh.Option[string], len(s.Books))
		for i, b := range s.Books {
			options[i] = huh.NewOption(b.Title, b.Title)
		}

		var selected string
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which entry do you want to inspect?").
					Options(options...).
					Value(&selected),
			),
		)

		if err := selectForm.Run(); err != nil {
			return err
		}

		var searchTitle string
		if useCustomSearch {
			searchForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Search Open Library under which title?").
						Placeholder(selected).
						Value(&searchTitle),
				),
			)
			if err := searchForm.Run(); err != nil {
				return err
			}
		}

		var book store.Book
		for _, b := range s.Books {
			if b.Title == selected {
				book = b
				break
			}
		}

		// Local reading stats
		readStatus := ui.Danger.Render("✗ not yet")
		if book.Completed {
			readStatus = ui.Warning.Render("★ completed")
		} else if book.WasReadToday() {
			readStatus = ui.Success.Render("✓ yes")
		}

		lastRead := ui.Muted.Render("never")
		if book.LastReadDate != "" {
			lastRead = book.LastReadDate
		}

		session := ui.Muted.Render("—")
		unitsRead := ui.Muted.Render("—")
		if book.LastReadDate != "" {
			session = fmt.Sprintf("%d → %d", book.PreviousPage, book.CurrentPage)
			unitsRead = ui.Success.Render(fmt.Sprintf("+%d", book.CurrentPage-book.PreviousPage))
		}

		addedAt := book.AddedAt.Format("2006-01-02")

		// Adapt labels based on entry type.
		detailTitle := "Book Details"
		positionLabel := "Current page: "
		unitsReadLabel := "Pages read:   "
		positionValue := fmt.Sprintf("%d / %d", book.CurrentPage, book.TotalPages)
		if book.IsMangaComic {
			detailTitle = "Manga Details"
			if book.TrackingUnit == "chapter" {
				positionLabel = "Current ch.:  "
				unitsReadLabel = "Chapters:     "
			} else {
				positionLabel = "Current vol.: "
				unitsReadLabel = "Volumes:      "
			}
			positionValue = fmt.Sprintf("%s%d / %d", mangaComicUnitLabel(book), book.CurrentPage, book.TotalPages)
		}

		// Open Library fetch — skipped for manga entries.
		var remoteSection string
		if !book.IsMangaComic {
			author := book.Author
			publishYear := book.PublishYear
			workKey := book.WorkKey

			query := book.Title
			if searchTitle != "" {
				query = searchTitle
			}

			if workKey == "" || refresh || searchTitle != "" {
				fmt.Print(ui.Muted.Render("Fetching book info from Open Library..."))
				meta, metaErr := fetchBookMeta(query)
				fmt.Print("\r\033[K") // clear the loading line
				if metaErr == nil {
					workKey = meta.WorkKey
					author = meta.Author
					publishYear = meta.PublishYear
					// Only persist when not using a custom search title.
					if searchTitle == "" {
						_ = s.UpdateBookMeta(book.Title, workKey, author, publishYear)
					}
				}
			}

			// Fetch the rating using the work key (always live, not cached).
			var ratingAvg float64
			var ratingCount int
			if workKey != "" {
				ratingAvg, ratingCount, _ = fetchRating(workKey)
			}

			if workKey == "" && author == "" {
				remoteSection = "\n" + ui.Muted.Render("Open Library: no results found") + "\n"
			} else {
				publishYearStr := ui.Muted.Render("unknown")
				if publishYear > 0 {
					publishYearStr = fmt.Sprintf("%d", publishYear)
				}
				authorStr := ui.Muted.Render("unknown")
				if author != "" {
					authorStr = author
				}
				remoteSection = "\n" +
					ui.Subtitle.Render("── Open Library ──") + "\n" +
					ui.Bold.Render("Author:       ") + authorStr + "\n" +
					ui.Bold.Render("Published:    ") + publishYearStr + "\n" +
					ui.Bold.Render("Rating:       ") + starRating(ratingAvg, ratingCount) + "\n"
			}
		}

		fmt.Println(ui.Box.Render(
			ui.Title.Render(detailTitle) + "\n\n" +
				ui.Bold.Render("Title:        ") + book.Title + "\n" +
				ui.Bold.Render(positionLabel) + positionValue + "\n" +
				ui.Bold.Render("Progress:     ") + progressBar(book.CurrentPage, book.TotalPages) + "\n" +
				ui.Bold.Render("Last session: ") + session + "\n" +
				ui.Bold.Render(unitsReadLabel) + unitsRead + "\n" +
				ui.Bold.Render("Last read:    ") + lastRead + "\n" +
				ui.Bold.Render("Added on:     ") + addedAt + "\n" +
				ui.Bold.Render("Read today:   ") + readStatus +
				remoteSection,
		))

		return nil
	},
}

// bookModifyCmd lets the user update the title or total count of an entry.
// It presents an entry selector, a field selector, and then prompts for the new value.
// The field options are adapted based on whether the entry is a book or manga.
var bookModifyCmd = &cobra.Command{
	Use:     "modified",
	Aliases: []string{"mod"},
	Short:   "Update the title or total count of a book or manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No entries yet. Add one with: grim add"))
			return nil
		}

		options := make([]huh.Option[string], len(s.Books))
		for i, b := range s.Books {
			options[i] = huh.NewOption(b.Title, b.Title)
		}

		var selected string
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which entry do you want to modify?").
					Options(options...).
					Value(&selected),
			),
		)

		if err := selectForm.Run(); err != nil {
			return err
		}

		// Find the selected entry to set context-aware labels.
		var entry store.Book
		for _, b := range s.Books {
			if b.Title == selected {
				entry = b
				break
			}
		}

		totalLabel := "Total pages"
		if entry.IsMangaComic {
			if entry.TrackingUnit == "chapter" {
				totalLabel = "Total chapters"
			} else {
				totalLabel = "Total volumes"
			}
		}

		var field string
		fieldForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What do you want to change?").
					Options(
						huh.NewOption("Title", "title"),
						huh.NewOption(totalLabel, "total"),
					).
					Value(&field),
			),
		)

		if err := fieldForm.Run(); err != nil {
			return err
		}

		if field == "title" {
			var newTitle string

			titleForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("New title").
						Placeholder(selected).
						Value(&newTitle),
				),
			)

			if err := titleForm.Run(); err != nil {
				return err
			}

			if err := s.UpdateTitle(selected, newTitle); err != nil {
				return err
			}

			fmt.Println(ui.Box.Render(
				ui.Title.Render("Entry updated") + "\n\n" +
					ui.Bold.Render("Old title: ") + selected + "\n" +
					ui.Bold.Render("New title: ") + newTitle,
			))

			return nil
		}

		// field == "total"
		var totalStr string

		totalForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("New %s", strings.ToLower(totalLabel))).
					Placeholder("350").
					Validate(func(s string) error {
						if _, err := strconv.Atoi(s); err != nil {
							return fmt.Errorf("must be a number")
						}
						return nil
					}).
					Value(&totalStr),
			),
		)

		if err := totalForm.Run(); err != nil {
			return err
		}

		total, _ := strconv.Atoi(totalStr)

		if err := s.UpdateTotalPages(selected, total); err != nil {
			return err
		}

		fmt.Println(ui.Box.Render(
			ui.Title.Render("Entry updated") + "\n\n" +
				ui.Bold.Render("Title:        ") + selected + "\n" +
				ui.Bold.Render(totalLabel+": ") + fmt.Sprintf("%d", total),
		))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bookAddCmd)
	rootCmd.AddCommand(bookListCmd)
	rootCmd.AddCommand(bookDeleteCmd)
	rootCmd.AddCommand(bookReadCmd)
	rootCmd.AddCommand(bookDetailCmd)
	rootCmd.AddCommand(bookModifyCmd)

	bookDetailCmd.Flags().BoolP("refresh", "r", false, "Force a new Open Library search even if cached metadata exists")
	bookDetailCmd.Flags().BoolP("search", "s", false, "Prompt for a different title to search on Open Library (result is not saved)")
}
