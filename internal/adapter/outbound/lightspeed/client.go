package lightspeed

import (
	"context"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// ErrNotConfigured is returned until Lightspeed credentials and HTTP mapping are wired.
var ErrNotConfigured = errors.New("lightspeed booking provider is not configured")

// Client is the Lightspeed tee-sheet adapter. It implements port.BookingProvider.
// HTTP calls are intentionally stubbed until API credentials and contract are ready;
// domain booking flows can still be tested with fakes.
type Client struct {
	baseURL string
	apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) ProviderName() string {
	return entity.ProviderLightspeed
}

func (c *Client) configured() bool {
	return c != nil && c.baseURL != "" && c.apiKey != ""
}

func (c *Client) SearchAvailability(ctx context.Context, courseExternalID string, from, to time.Time) ([]port.BookingSlot, error) {
	if !c.configured() {
		return nil, ErrNotConfigured
	}
	_ = ctx
	_ = courseExternalID
	_ = from
	_ = to
	return nil, ErrNotConfigured
}

func (c *Client) Hold(ctx context.Context, req port.HoldRequest) (*port.HoldResult, error) {
	if !c.configured() {
		return nil, ErrNotConfigured
	}
	_ = ctx
	_ = req
	return nil, ErrNotConfigured
}

func (c *Client) Confirm(ctx context.Context, req port.ConfirmRequest) (*port.ConfirmResult, error) {
	if !c.configured() {
		return nil, ErrNotConfigured
	}
	_ = ctx
	_ = req
	return nil, ErrNotConfigured
}

func (c *Client) Cancel(ctx context.Context, req port.CancelRequest) error {
	if !c.configured() {
		return ErrNotConfigured
	}
	_ = ctx
	_ = req
	return ErrNotConfigured
}

var _ port.BookingProvider = (*Client)(nil)
