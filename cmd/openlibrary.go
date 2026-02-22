package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// olSearchResponse mirrors the relevant fields from the Open Library
// search API: https://openlibrary.org/search.json?title=<title>&limit=1
type olSearchResponse struct {
	NumFound int      `json:"numFound"`
	Docs     []olDoc `json:"docs"`
}

type olDoc struct {
	Key              string   `json:"key"` // e.g. "/works/OL45804W"
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
}

// olRatingResponse mirrors the Open Library ratings endpoint.
// https://openlibrary.org/works/<id>/ratings.json
type olRatingResponse struct {
	Summary struct {
		Average float64 `json:"average"`
		Count   int     `json:"count"`
	} `json:"summary"`
}

// BookMeta holds the Open Library metadata returned by fetchBookMeta.
type BookMeta struct {
	WorkKey     string
	Author      string
	PublishYear int
}

var olClient = &http.Client{Timeout: 8 * time.Second}

// fetchBookMeta queries the Open Library search API and returns the work key,
// author, and publish year for the best matching result.
func fetchBookMeta(title string) (*BookMeta, error) {
	searchURL := "https://openlibrary.org/search.json?limit=1&language=eng&title=" + url.QueryEscape(title)

	resp, err := olClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result olSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if result.NumFound == 0 || len(result.Docs) == 0 {
		return nil, fmt.Errorf("no results found for %q", title)
	}

	doc := result.Docs[0]
	meta := &BookMeta{
		WorkKey:     doc.Key,
		PublishYear: doc.FirstPublishYear,
	}
	if len(doc.AuthorName) > 0 {
		meta.Author = doc.AuthorName[0]
	}

	return meta, nil
}

// fetchRating fetches the rating for a work key (e.g. "/works/OL45804W").
// Returns 0, 0 without error when the work has no ratings yet.
func fetchRating(workKey string) (average float64, count int, err error) {
	ratingURL := "https://openlibrary.org" + workKey + "/ratings.json"

	resp, err := olClient.Get(ratingURL)
	if err != nil {
		return 0, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var rating olRatingResponse
	if err := json.NewDecoder(resp.Body).Decode(&rating); err != nil {
		return 0, 0, fmt.Errorf("decode error: %w", err)
	}

	return rating.Summary.Average, rating.Summary.Count, nil
}

// starRating renders a 5-star string from a 0–5 average.
// e.g. "★★★★☆ 4.2 (1 234 ratings)"
func starRating(average float64, count int) string {
	if count == 0 {
		return "no ratings yet"
	}
	full := int(average + 0.5)
	bar := ""
	for i := 0; i < 5; i++ {
		if i < full {
			bar += "★"
		} else {
			bar += "☆"
		}
	}
	return fmt.Sprintf("%s %.1f (%s ratings)", bar, average, formatCount(count))
}

// formatCount formats an integer with spaces as thousands separators.
func formatCount(n int) string {
	s := fmt.Sprintf("%d", n)
	out := ""
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += " "
		}
		out += string(ch)
	}
	return out
}
