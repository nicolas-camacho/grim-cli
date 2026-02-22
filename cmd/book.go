package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nicolas-camacho/grim-cli/store"
	"github.com/nicolas-camacho/grim-cli/ui"
	"github.com/spf13/cobra"
)

// bookAddCmd launches an interactive huh form that collects book details
// sequentially and persists the new entry via the store.
var bookAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a book to your reading list",
	RunE: func(cmd *cobra.Command, args []string) error {
		var title, pageStr, totalPagesStr string
		var readToday bool

		// numValidate is reused for both numeric inputs.
		numValidate := func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be a number")
			}
			return nil
		}

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

		s, err := store.New()
		if err != nil {
			return err
		}
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
	},
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
	if pct > 100 {
		pct = 100
	}
	const width = 10
	filled := width * pct / 100
	bar := ui.Title.Render(repeatStr("█", filled)) + ui.Muted.Render(repeatStr("░", width-filled))
	return fmt.Sprintf("%s %d%%", bar, pct)
}

// repeatStr returns s repeated n times.
func repeatStr(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}

// bookListCmd loads all books from the store and renders them in a styled table.
// The "Read Today" column is evaluated live so it resets at midnight automatically.
var bookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all books and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No books yet. Add one with: grim add"))
			return nil
		}

		headers := []string{"Title", "Page", "Progress", "Last Read", "Session", "Pages Read", "Read Today"}
		rows := make([][]string, len(s.Books))
		for i, b := range s.Books {
			readStatus := ui.Danger.Render("✗ not yet")
			if b.WasReadToday() {
				readStatus = ui.Success.Render("✓ yes")
			}

			lastRead := ui.Muted.Render("never")
			if b.LastReadDate != "" {
				lastRead = b.LastReadDate
			}

			// Session and pages read are only meaningful when the book has been read at least once.
			session := ui.Muted.Render("—")
			pagesRead := ui.Muted.Render("—")
			if b.LastReadDate != "" {
				session = fmt.Sprintf("%d → %d", b.PreviousPage, b.CurrentPage)
				pagesRead = fmt.Sprintf("%d", b.CurrentPage-b.PreviousPage)
			}

			rows[i] = []string{
				b.Title,
				fmt.Sprintf("%d", b.CurrentPage),
				progressBar(b.CurrentPage, b.TotalPages),
				lastRead,
				session,
				pagesRead,
				readStatus,
			}
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#A673FF"))).
			Headers(headers...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A673FF"))
				}
				// Alternate row shading for readability.
				if row%2 == 0 {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
				}
				return lipgloss.NewStyle()
			}).
			Rows(rows...)

		fmt.Println(ui.Title.Render("Reading List"))
		fmt.Println(t)
		return nil
	},
}

// bookDeleteCmd presents an interactive selector of all books, then asks for
// confirmation before removing the chosen entry from the store.
var bookDeleteCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete a book from your reading list",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		if len(s.Books) == 0 {
			fmt.Println(ui.Muted.Render("No books yet. Add one with: grim add"))
			return nil
		}

		// Build the option list from current books.
		options := make([]huh.Option[string], len(s.Books))
		for i, b := range s.Books {
			options[i] = huh.NewOption(b.Title, b.Title)
		}

		var selected string
		var confirmed bool

		// Step 1: pick a book.
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which book do you want to delete?").
					Options(options...).
					Value(&selected),
			),
		)

		if err := selectForm.Run(); err != nil {
			return err
		}

		// Step 2: confirm with the book title in the message.
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
			fmt.Println(ui.Muted.Render("Cancelled. No books were deleted."))
			return nil
		}

		if err := s.DeleteBook(selected); err != nil {
			return err
		}

		fmt.Println(ui.Success.Render("✓ ") + ui.Bold.Render(selected) + " removed from your reading list.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bookAddCmd)
	rootCmd.AddCommand(bookListCmd)
	rootCmd.AddCommand(bookDeleteCmd)
}
