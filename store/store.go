// Package store handles data persistence for grim-cli.
// Books are stored as JSON in ~/.grim/books.json and loaded into memory
// on each command invocation via Store.New().
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ReadingSession records a single reading session for an entry.
type ReadingSession struct {
	Date string `json:"date"` // format: "2006-01-02"
	From int    `json:"from"`
	To   int    `json:"to"`
}

// Book represents a single entry in the reading list.
// It handles both books (tracked by pages) and manga (tracked by volumes or chapters).
type Book struct {
	Title        string    `json:"title"`
	PreviousPage int       `json:"previous_page"` // page/vol/chapter before the last update
	CurrentPage  int       `json:"current_page"`
	TotalPages   int       `json:"total_pages"`
	ReadToday    bool      `json:"read_today"`
	LastReadDate string    `json:"last_read_date"` // format: "2006-01-02"
	Completed    bool      `json:"completed"`
	AddedAt      time.Time `json:"added_at"`

	// Manga-specific fields.
	IsMangaComic bool   `json:"is_manga,omitempty"`
	TrackingUnit string `json:"tracking_unit,omitempty"` // "volume" or "chapter"

	// Open Library metadata — populated on first `grim dt` call (books only).
	WorkKey     string `json:"work_key,omitempty"` // e.g. "/works/OL45804W"
	Author      string `json:"author,omitempty"`
	PublishYear int    `json:"publish_year,omitempty"`

	// Session history — appended on every read log.
	Sessions []ReadingSession `json:"sessions,omitempty"`
}

// WasReadToday returns true if the book's LastReadDate matches today's date.
// This is evaluated at runtime rather than trusting the stored ReadToday bool,
// so the status resets automatically at midnight without modifying any data.
func (b Book) WasReadToday() bool {
	return b.LastReadDate == time.Now().Format("2006-01-02")
}

// Store holds the full list of books and the path to the JSON file.
type Store struct {
	Books []Book `json:"books"`
	path  string
}

// New creates a Store, ensures the ~/.grim directory exists, and loads
// any existing data from books.json. If the file does not exist yet it
// returns an empty store without error.
func New() (*Store, error) {
	dir, err := grimDir()
	if err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "books.json")}
	return s, s.load()
}

// grimDir returns the path to ~/.grim and creates the directory if it
// does not already exist.
func grimDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".grim")
	return dir, os.MkdirAll(dir, 0755)
}

// load reads and unmarshals the JSON file into the store.
// It is a no-op when the file does not exist yet.
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil // fresh start, no file yet
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

// save marshals the store to indented JSON and writes it to disk.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// AddBook appends a new book to the list and persists the change.
// PreviousPage is always initialised to 0 on creation.
// LastReadDate is set to today only when readToday is true.
func (s *Store) AddBook(title string, page, totalPages int, readToday bool) error {
	today := time.Now().Format("2006-01-02")
	lastReadDate := ""
	var sessions []ReadingSession
	if readToday {
		lastReadDate = today
		sessions = []ReadingSession{{Date: today, From: 0, To: page}}
	}
	s.Books = append(s.Books, Book{
		Title:        title,
		PreviousPage: 0,
		CurrentPage:  page,
		TotalPages:   totalPages,
		ReadToday:    readToday,
		LastReadDate: lastReadDate,
		AddedAt:      time.Now(),
		Sessions:     sessions,
	})
	return s.save()
}

// AddManga appends a new manga to the list and persists the change.
// trackingUnit must be "volume" or "chapter".
func (s *Store) AddMangaComic(title, trackingUnit string, current, total int, readToday bool) error {
	today := time.Now().Format("2006-01-02")
	lastReadDate := ""
	var sessions []ReadingSession
	if readToday {
		lastReadDate = today
		sessions = []ReadingSession{{Date: today, From: 0, To: current}}
	}
	s.Books = append(s.Books, Book{
		Title:        title,
		PreviousPage: 0,
		CurrentPage:  current,
		TotalPages:   total,
		ReadToday:    readToday,
		LastReadDate: lastReadDate,
		AddedAt:      time.Now(),
		IsMangaComic: true,
		TrackingUnit: trackingUnit,
		Sessions:     sessions,
	})
	return s.save()
}

// UpdateBook marks a book as read today, shifting CurrentPage to PreviousPage
// and setting the new page as CurrentPage. LastReadDate is set to today.
func (s *Store) UpdateBook(title string, newPage int) error {
	today := time.Now().Format("2006-01-02")
	for i, b := range s.Books {
		if b.Title == title {
			s.Books[i].PreviousPage = b.CurrentPage
			s.Books[i].CurrentPage = newPage
			s.Books[i].LastReadDate = today
			s.Books[i].ReadToday = true
			s.Books[i].Sessions = append(s.Books[i].Sessions, ReadingSession{
				Date: today,
				From: b.CurrentPage,
				To:   newPage,
			})
			return s.save()
		}
	}
	return nil
}

// CompleteBook marks a book as finished: sets CurrentPage to TotalPages,
// records today as the last read date, and flags it as completed.
func (s *Store) CompleteBook(title string) error {
	today := time.Now().Format("2006-01-02")
	for i, b := range s.Books {
		if b.Title == title {
			s.Books[i].PreviousPage = b.CurrentPage
			s.Books[i].CurrentPage = b.TotalPages
			s.Books[i].LastReadDate = today
			s.Books[i].ReadToday = true
			s.Books[i].Completed = true
			s.Books[i].Sessions = append(s.Books[i].Sessions, ReadingSession{
				Date: today,
				From: b.CurrentPage,
				To:   b.TotalPages,
			})
			return s.save()
		}
	}
	return nil
}

// UpdateBookMeta stores the Open Library metadata for a book and persists the change.
// It is a no-op if no book with the given title is found.
func (s *Store) UpdateBookMeta(title, workKey, author string, publishYear int) error {
	for i, b := range s.Books {
		if b.Title == title {
			s.Books[i].WorkKey = workKey
			s.Books[i].Author = author
			s.Books[i].PublishYear = publishYear
			return s.save()
		}
	}
	return nil
}

// UpdateTotalPages sets a new total page count for the given book and persists the change.
// It is a no-op if no book with the given title is found.
func (s *Store) UpdateTotalPages(title string, totalPages int) error {
	for i, b := range s.Books {
		if b.Title == title {
			s.Books[i].TotalPages = totalPages
			return s.save()
		}
	}
	return nil
}

// UpdateTitle renames a book and persists the change.
// It is a no-op if no book with the given title is found.
func (s *Store) UpdateTitle(oldTitle, newTitle string) error {
	for i, b := range s.Books {
		if b.Title == oldTitle {
			s.Books[i].Title = newTitle
			return s.save()
		}
	}
	return nil
}

// DeleteBook removes the first book whose title matches and persists the change.
// It is a no-op if no match is found.
func (s *Store) DeleteBook(title string) error {
	for i, b := range s.Books {
		if b.Title == title {
			s.Books = append(s.Books[:i], s.Books[i+1:]...)
			return s.save()
		}
	}
	return nil
}
