package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nicolas-camacho/grim-cli/store"
	"github.com/nicolas-camacho/grim-cli/ui"
	"github.com/spf13/cobra"
)

// bookLogCmd lists all entries that have at least one reading session logged.
// The user picks one and sees the full session history in a styled table.
var bookLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Show reading session history for a book or manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}

		// Collect entries with at least one session.
		var withSessions []store.Book
		for _, b := range s.Books {
			if len(b.Sessions) > 0 {
				withSessions = append(withSessions, b)
			}
		}

		if len(withSessions) == 0 {
			fmt.Println(ui.Muted.Render("No sessions logged yet. Use: grim read"))
			return nil
		}

		options := make([]huh.Option[string], len(withSessions))
		for i, b := range withSessions {
			label := b.Title
			if b.IsMangaComic {
				label += ui.Muted.Render(" (manga/comic)")
			}
			options[i] = huh.NewOption(label, b.Title)
		}

		var selected string
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which entry do you want to see the log for?").
					Options(options...).
					Value(&selected),
			),
		)
		if err := selectForm.Run(); err != nil {
			return err
		}

		var book store.Book
		for _, b := range withSessions {
			if b.Title == selected {
				book = b
				break
			}
		}

		// Build unit label for the "Read" column header.
		readHeader := "Pages Read"
		if book.IsMangaComic {
			if book.TrackingUnit == "chapter" {
				readHeader = "Chapters Read"
			} else {
				readHeader = "Volumes Read"
			}
		}

		// Build rows — newest session first.
		sessions := book.Sessions
		rows := make([][]string, len(sessions))
		for i := range sessions {
			sess := sessions[len(sessions)-1-i] // reverse order
			delta := sess.To - sess.From
			deltaStr := ui.Success.Render(fmt.Sprintf("+%d", delta))
			if delta <= 0 {
				deltaStr = ui.Muted.Render("—")
			}
			rows[i] = []string{
				sess.Date,
				fmt.Sprintf("%d → %d", sess.From, sess.To),
				deltaStr,
			}
		}

		headers := []string{"Date", "Session", readHeader}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#A673FF"))).
			Headers(headers...).
			StyleFunc(func(row, col int) lipgloss.Style {
				base := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
				if row == table.HeaderRow {
					return base.Bold(true).Foreground(lipgloss.Color("#A673FF"))
				}
				if row%2 == 0 {
					return base.Foreground(lipgloss.Color("#E5E7EB"))
				}
				return base
			}).
			Rows(rows...)

		fmt.Println(ui.Title.Render("Reading Log") + "  " + ui.Subtitle.Render(book.Title))
		fmt.Println(ui.Muted.Render(fmt.Sprintf("%d session(s) logged", len(sessions))))
		fmt.Println()
		fmt.Println(t)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bookLogCmd)
}
