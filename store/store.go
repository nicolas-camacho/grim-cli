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

// Book represents a single entry in the reading list.
type Book struct {
	Title        string    `json:"title"`
	PreviousPage int       `json:"previous_page"` // page before the last update
	CurrentPage  int       `json:"current_page"`
	TotalPages   int       `json:"total_pages"`
	ReadToday    bool      `json:"read_today"`
	LastReadDate string    `json:"last_read_date"` // format: "2006-01-02"
	AddedAt      time.Time `json:"added_at"`
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
	lastReadDate := ""
	if readToday {
		lastReadDate = time.Now().Format("2006-01-02")
	}
	s.Books = append(s.Books, Book{
		Title:        title,
		PreviousPage: 0,
		CurrentPage:  page,
		TotalPages:   totalPages,
		ReadToday:    readToday,
		LastReadDate: lastReadDate,
		AddedAt:      time.Now(),
	})
	return s.save()
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
