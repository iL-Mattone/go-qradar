package qradar

import (
	"context"
	"fmt"
	"net/http"
)

// SearchResultsWindow is a default window for scrolling results of the query.
var SearchResultsWindow = 50

// SearchResultsScroller represents a scroller for the results of the query.
type SearchResultsScroller struct {
	Count             int
	client            *Client
	searchID          string
	startIdx, currIdx int
	window            int
	events            []Event
}

// NewSearchResultsScroller initializes struct to scroll the records.
func (a *ArielService) NewSearchResultsScroller(ctx context.Context, searchID string) (*SearchResultsScroller, error) {
	_, num, err := a.SearchStatus(ctx, searchID)

	srs := &SearchResultsScroller{
		Count:    num,
		window:   SearchResultsWindow,
		client:   a.client,
		searchID: searchID,
	}

	err = srs.getEvents(ctx)
	if err != nil {
		return nil, err
	}

	return srs, nil
}

// Next returns true if an event is still available to be consumed by the
// Result() method.
func (s *SearchResultsScroller) Next(ctx context.Context) bool {
	if s.currIdx-s.startIdx == len(s.events) && len(s.events) < s.window {
		return false
	}

	if s.currIdx-s.startIdx > len(s.events) && len(s.events) == s.window {
		s.startIdx += s.window
		err := s.getEvents(ctx)
		if err != nil {
			return false
		}
	}

	return true
}

func (s *SearchResultsScroller) getEvents(ctx context.Context) error {
	req, err := s.client.NewRequest("GET", fmt.Sprintf("%s/%s/results", arielSearchAPIPrefix, s.searchID), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Range", fmt.Sprintf("items=%d-%d", s.startIdx, s.startIdx+s.window))

	var r SearchResult
	resp, err := s.client.Do(ctx, req, &r)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SearchResultScroller failed: code %s", resp.Status)
	}

	s.events = r.Events

	return nil
}

// Result returns the event iterated by the Next.
func (s *SearchResultsScroller) Result() Event {
	event := s.events[s.currIdx-s.startIdx]
	s.currIdx++
	return event

}
