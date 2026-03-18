package golfcourseapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type golfCourseAPIResponse struct {
	Courses []golfCourseResult `json:"courses"`
}

type golfCourseResult struct {
	ID         int                `json:"id"`
	ClubName   string             `json:"club_name"`
	CourseName string             `json:"course_name"`
	Location   golfCourseLocation `json:"location"`
}

type golfCourseLocation struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Client implements port.GolfCourseSearcher by calling the Golf Course API.
type Client struct {
	apiKey string
}

// NewClient returns a Client configured with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

// Search queries the Golf Course API and returns matching courses.
// It returns an empty slice when query is blank.
func (c *Client) Search(ctx context.Context, query string) ([]entity.Course, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []entity.Course{}, nil
	}

	apiURL := fmt.Sprintf("https://api.golfcourseapi.com/v1/search?search_query=%s&limit=10",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("golfcourseapi: create request: %w", err)
	}
	req.Header.Set("Authorization", "Key "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("golfcourseapi: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("golfcourseapi: read response: %w", err)
	}

	var apiResp golfCourseAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("golfcourseapi: parse response: %w", err)
	}

	results := make([]entity.Course, 0, len(apiResp.Courses))
	for _, cr := range apiResp.Courses {
		name := cr.CourseName
		if name == "" {
			name = cr.ClubName
		}

		street := cr.Location.Address
		if idx := strings.Index(street, ", "+cr.Location.City); idx > 0 {
			street = street[:idx]
		}

		results = append(results, entity.Course{
			Name:   name,
			Street: street,
			City:   cr.Location.City,
			State:  cr.Location.State,
		})
	}

	return results, nil
}
