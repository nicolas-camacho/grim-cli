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

// BookInfo holds the enriched metadata returned by Open Library.
type BookInfo struct {
	Author        string
	PublishYear   int
	RatingAverage float64
	RatingCount   int
}

var olClient = &http.Client{Timeout: 8 * time.Second}

// fetchBookInfo queries the Open Library search API with the given title,
// then fetches the work's rating in a second call.
func fetchBookInfo(title string) (*BookInfo, error) {
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

	info := &BookInfo{PublishYear: doc.FirstPublishYear}
	if len(doc.AuthorName) > 0 {
		info.Author = doc.AuthorName[0]
	}

	// Second call: fetch rating using the work key.
	if doc.Key != "" {
		ratingURL := "https://openlibrary.org" + doc.Key + "/ratings.json"
		rResp, err := olClient.Get(ratingURL)
		if err == nil && rResp.StatusCode == http.StatusOK {
			var rating olRatingResponse
			if json.NewDecoder(rResp.Body).Decode(&rating) == nil {
				info.RatingAverage = rating.Summary.Average
				info.RatingCount = rating.Summary.Count
			}
			rResp.Body.Close()
		}
	}

	return info, nil
}

// starRating renders a 5-star string from a 0–5 average, e.g. "★★★★☆ 4.2 (1 234 ratings)".
func starRating(average float64, count int) string {
	if count == 0 {
		return "no ratings yet"
	}
	full := int(average + 0.5) // round to nearest star
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

// formatCount formats an integer with a thin space as thousands separator.
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
