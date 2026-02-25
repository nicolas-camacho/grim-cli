package store

import (
	"path/filepath"
	"testing"
	"time"
)

// newTestStore creates a Store backed by a temporary directory so tests
// never touch ~/.grim and are fully isolated from each other.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	return &Store{path: filepath.Join(t.TempDir(), "books.json")}
}

// ---------------------------------------------------------------------------
// WasReadToday
// ---------------------------------------------------------------------------

func TestWasReadToday_True(t *testing.T) {
	b := Book{LastReadDate: time.Now().Format("2006-01-02")}
	if !b.WasReadToday() {
		t.Error("expected WasReadToday to return true for today's date")
	}
}

func TestWasReadToday_False_OldDate(t *testing.T) {
	b := Book{LastReadDate: "2000-01-01"}
	if b.WasReadToday() {
		t.Error("expected WasReadToday to return false for a past date")
	}
}

func TestWasReadToday_False_Empty(t *testing.T) {
	b := Book{}
	if b.WasReadToday() {
		t.Error("expected WasReadToday to return false when LastReadDate is empty")
	}
}

// ---------------------------------------------------------------------------
// AddBook
// ---------------------------------------------------------------------------

func TestAddBook_AppendsToList(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddBook("Dune", 100, 688, false); err != nil {
		t.Fatalf("AddBook returned unexpected error: %v", err)
	}

	if len(s.Books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(s.Books))
	}

	b := s.Books[0]
	if b.Title != "Dune" {
		t.Errorf("expected title %q, got %q", "Dune", b.Title)
	}
	if b.CurrentPage != 100 {
		t.Errorf("expected CurrentPage 100, got %d", b.CurrentPage)
	}
	if b.TotalPages != 688 {
		t.Errorf("expected TotalPages 688, got %d", b.TotalPages)
	}
	if b.PreviousPage != 0 {
		t.Errorf("expected PreviousPage 0 on creation, got %d", b.PreviousPage)
	}
}

func TestAddBook_ReadToday_SetsLastReadDate(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddBook("Dune", 100, 688, true); err != nil {
		t.Fatalf("AddBook returned unexpected error: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if s.Books[0].LastReadDate != today {
		t.Errorf("expected LastReadDate %q, got %q", today, s.Books[0].LastReadDate)
	}
}

func TestAddBook_NotReadToday_EmptyLastReadDate(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddBook("Dune", 100, 688, false); err != nil {
		t.Fatalf("AddBook returned unexpected error: %v", err)
	}

	if s.Books[0].LastReadDate != "" {
		t.Errorf("expected empty LastReadDate, got %q", s.Books[0].LastReadDate)
	}
}

func TestAddBook_MultipleBooks(t *testing.T) {
	s := newTestStore(t)

	titles := []string{"Book A", "Book B", "Book C"}
	for _, title := range titles {
		if err := s.AddBook(title, 0, 100, false); err != nil {
			t.Fatalf("AddBook(%q) returned unexpected error: %v", title, err)
		}
	}

	if len(s.Books) != 3 {
		t.Fatalf("expected 3 books, got %d", len(s.Books))
	}
}

// ---------------------------------------------------------------------------
// AddManga
// ---------------------------------------------------------------------------

func TestAddManga_AppendsToList(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddMangaComic("One Piece", "volume", 5, 110, false); err != nil {
		t.Fatalf("AddManga returned unexpected error: %v", err)
	}

	if len(s.Books) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(s.Books))
	}

	b := s.Books[0]
	if b.Title != "One Piece" {
		t.Errorf("expected title %q, got %q", "One Piece", b.Title)
	}
	if !b.IsMangaComic {
		t.Error("expected IsManga to be true")
	}
	if b.TrackingUnit != "volume" {
		t.Errorf("expected TrackingUnit %q, got %q", "volume", b.TrackingUnit)
	}
	if b.CurrentPage != 5 {
		t.Errorf("expected CurrentPage 5, got %d", b.CurrentPage)
	}
	if b.TotalPages != 110 {
		t.Errorf("expected TotalPages 110, got %d", b.TotalPages)
	}
}

func TestAddManga_ReadToday_SetsLastReadDate(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddMangaComic("Berserk", "chapter", 200, 364, true); err != nil {
		t.Fatalf("AddManga returned unexpected error: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if s.Books[0].LastReadDate != today {
		t.Errorf("expected LastReadDate %q, got %q", today, s.Books[0].LastReadDate)
	}
}

func TestAddManga_MixedWithBooks(t *testing.T) {
	s := newTestStore(t)

	_ = s.AddBook("Dune", 100, 688, false)
	_ = s.AddMangaComic("Vagabond", "volume", 10, 37, false)

	if len(s.Books) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(s.Books))
	}
	if s.Books[0].IsMangaComic {
		t.Error("expected first entry to not be manga")
	}
	if !s.Books[1].IsMangaComic {
		t.Error("expected second entry to be manga")
	}
}

// ---------------------------------------------------------------------------
// DeleteBook
// ---------------------------------------------------------------------------

func TestDeleteBook_RemovesCorrectBook(t *testing.T) {
	s := newTestStore(t)
	_ = s.AddBook("Keep Me", 0, 100, false)
	_ = s.AddBook("Delete Me", 0, 100, false)

	if err := s.DeleteBook("Delete Me"); err != nil {
		t.Fatalf("DeleteBook returned unexpected error: %v", err)
	}

	if len(s.Books) != 1 {
		t.Fatalf("expected 1 book after deletion, got %d", len(s.Books))
	}
	if s.Books[0].Title != "Keep Me" {
		t.Errorf("expected remaining book to be %q, got %q", "Keep Me", s.Books[0].Title)
	}
}

func TestDeleteBook_NoMatch_IsNoop(t *testing.T) {
	s := newTestStore(t)
	_ = s.AddBook("Existing Book", 0, 100, false)

	if err := s.DeleteBook("Non-existent Book"); err != nil {
		t.Fatalf("DeleteBook returned unexpected error for non-existent title: %v", err)
	}

	if len(s.Books) != 1 {
		t.Errorf("expected list to remain unchanged, got %d books", len(s.Books))
	}
}

func TestDeleteBook_EmptyStore_IsNoop(t *testing.T) {
	s := newTestStore(t)

	if err := s.DeleteBook("Anything"); err != nil {
		t.Fatalf("DeleteBook on empty store returned unexpected error: %v", err)
	}
}

func TestDeleteBook_OnlyOneBook_LeavesEmptyList(t *testing.T) {
	s := newTestStore(t)
	_ = s.AddBook("Solo Book", 0, 100, false)

	if err := s.DeleteBook("Solo Book"); err != nil {
		t.Fatalf("DeleteBook returned unexpected error: %v", err)
	}

	if len(s.Books) != 0 {
		t.Errorf("expected empty list after deleting last book, got %d books", len(s.Books))
	}
}

// ---------------------------------------------------------------------------
// Persistence (save → reload)
// ---------------------------------------------------------------------------

func TestPersistence_ReloadRestoresBooks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "books.json")

	s1 := &Store{path: path}
	_ = s1.AddBook("Persisted Book", 42, 300, true)

	// Create a second store pointing to the same file to simulate a new process.
	s2 := &Store{path: path}
	if err := s2.load(); err != nil {
		t.Fatalf("load returned unexpected error: %v", err)
	}

	if len(s2.Books) != 1 {
		t.Fatalf("expected 1 book after reload, got %d", len(s2.Books))
	}

	b := s2.Books[0]
	if b.Title != "Persisted Book" {
		t.Errorf("expected title %q after reload, got %q", "Persisted Book", b.Title)
	}
	if b.CurrentPage != 42 {
		t.Errorf("expected CurrentPage 42 after reload, got %d", b.CurrentPage)
	}
}

func TestPersistence_LoadMissingFile_ReturnsEmptyStore(t *testing.T) {
	s := &Store{path: filepath.Join(t.TempDir(), "nonexistent.json")}

	if err := s.load(); err != nil {
		t.Fatalf("load of missing file returned unexpected error: %v", err)
	}

	if len(s.Books) != 0 {
		t.Errorf("expected empty book list, got %d", len(s.Books))
	}
}
