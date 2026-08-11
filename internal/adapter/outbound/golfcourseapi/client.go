package golfcourseapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

const maxLoggedBodyBytes = 512

type golfCourseAPIResponse struct {
	Courses []golfCourseResult `json:"courses"`
}

type golfCourseResult struct {
	ID         json.RawMessage    `json:"id"`
	ClubName   string             `json:"club_name"`
	CourseName string             `json:"course_name"`
	Location   golfCourseLocation `json:"location"`
}

type golfCourseLocation struct {
	Address   string   `json:"address"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Country   string   `json:"country"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// Client implements port.GolfCourseSearcher by calling the Golf Course API.
type Client struct {
	apiKey string
}

// NewClient returns a Client configured with the given API key.
// Trims surrounding whitespace so a stray newline in an env var or Secret Manager
// version can't break the Authorization header.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: strings.TrimSpace(apiKey)}
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

	if c.apiKey == "" {
		return nil, fmt.Errorf("golfcourseapi: API key is empty")
	}

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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("golfcourseapi non-2xx",
			"status", resp.StatusCode,
			"url", apiURL,
			"body", truncate(body, maxLoggedBodyBytes),
		)
		return nil, fmt.Errorf("golfcourseapi: unexpected status %d", resp.StatusCode)
	}

	var apiResp golfCourseAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		slog.Warn("golfcourseapi parse failed",
			"url", apiURL,
			"body", truncate(body, maxLoggedBodyBytes),
			"err", err,
		)
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

		country := cr.Location.Country
		if country == "" {
			country = "US"
		}

		results = append(results, entity.Course{
			Name:       name,
			Street:     street,
			City:       cr.Location.City,
			State:      cr.Location.State,
			Country:    country,
			Latitude:   cr.Location.Latitude,
			Longitude:  cr.Location.Longitude,
			Provider:   entity.ProviderGolfCourseAPI,
			ExternalID: parseExternalID(cr.ID),
		})
	}

	return results, nil
}

func parseExternalID(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}
	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber.String()
	}
	var asInt int64
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return strconv.FormatInt(asInt, 10)
	}
	return strings.Trim(string(raw), `"`)
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...(truncated)"
}
