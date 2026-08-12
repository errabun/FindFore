// Package fakebooking provides a configurable BookingProvider for failure-mode tests.
package fakebooking

import (
	"context"
	"sync"
	"time"

	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// Behavior controls Hold/Confirm/Cancel outcomes.
type Behavior int

const (
	BehaviorSuccess Behavior = iota
	BehaviorTimeout
	BehaviorReject
	BehaviorCancelFail
)

// Provider is a thread-safe fake BookingProvider.
type Provider struct {
	mu sync.Mutex

	Name string

	Slots []port.BookingSlot

	// SearchErr, when set, makes SearchAvailability return that error.
	SearchErr error

	HoldBehavior    Behavior
	ConfirmBehavior Behavior
	CancelBehavior  Behavior

	// ConfirmOnly makes Hold return ConfirmedImmediately (no separate hold step).
	ConfirmOnly bool

	// GoneExternalIDs causes Hold to reject for those tee-time external ids (stale inventory).
	GoneExternalIDs map[string]bool

	HoldCalls    int
	ConfirmCalls int
	CancelCalls  int

	// holdByKey remembers successful hold results for idempotent retries.
	holdByKey    map[string]*port.HoldResult
	confirmByKey map[string]*port.ConfirmResult
	cancelByKey  map[string]struct{}
}

func New(name string) *Provider {
	if name == "" {
		name = entity.ProviderLightspeed
	}
	return &Provider{
		Name:            name,
		GoneExternalIDs: map[string]bool{},
		holdByKey:       map[string]*port.HoldResult{},
		confirmByKey:    map[string]*port.ConfirmResult{},
		cancelByKey:     map[string]struct{}{},
	}
}

func (p *Provider) ProviderName() string {
	if p.Name == "" {
		return entity.ProviderLightspeed
	}
	return p.Name
}

func (p *Provider) SearchAvailability(context.Context, string, time.Time, time.Time) ([]port.BookingSlot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.SearchErr != nil {
		return nil, p.SearchErr
	}
	out := make([]port.BookingSlot, len(p.Slots))
	copy(out, p.Slots)
	return out, nil
}

func (p *Provider) Hold(_ context.Context, req port.HoldRequest) (*port.HoldResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.HoldCalls++

	if cached, ok := p.holdByKey[req.IdempotencyKey]; ok {
		cp := *cached
		return &cp, nil
	}

	if p.GoneExternalIDs[req.ExternalTeeTimeID] {
		return nil, booking.ErrProviderRejected
	}

	switch p.HoldBehavior {
	case BehaviorTimeout:
		return nil, booking.ErrProviderOutcomeUnknown
	case BehaviorReject:
		return nil, booking.ErrProviderRejected
	}

	exp := time.Now().UTC().Add(10 * time.Minute)
	result := &port.HoldResult{
		ExternalReservationID: "fake-hold-" + req.IdempotencyKey,
		HoldExpiresAt:         &exp,
		ConfirmedImmediately:  p.ConfirmOnly,
	}
	if p.ConfirmOnly {
		result.ExternalReservationID = "fake-conf-" + req.IdempotencyKey
		result.HoldExpiresAt = nil
	}
	cp := *result
	p.holdByKey[req.IdempotencyKey] = &cp
	return result, nil
}

func (p *Provider) Confirm(_ context.Context, req port.ConfirmRequest) (*port.ConfirmResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ConfirmCalls++

	if cached, ok := p.confirmByKey[req.IdempotencyKey]; ok {
		cp := *cached
		return &cp, nil
	}

	switch p.ConfirmBehavior {
	case BehaviorTimeout:
		return nil, booking.ErrProviderOutcomeUnknown
	case BehaviorReject:
		return nil, booking.ErrProviderRejected
	}

	id := req.ExternalReservationID
	if id == "" {
		id = "fake-conf-" + req.IdempotencyKey
	}
	result := &port.ConfirmResult{ExternalReservationID: id}
	cp := *result
	p.confirmByKey[req.IdempotencyKey] = &cp
	return result, nil
}

func (p *Provider) Cancel(_ context.Context, req port.CancelRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CancelCalls++

	if _, ok := p.cancelByKey[req.IdempotencyKey]; ok {
		return nil
	}

	switch p.CancelBehavior {
	case BehaviorTimeout:
		return booking.ErrProviderOutcomeUnknown
	case BehaviorCancelFail, BehaviorReject:
		return booking.ErrProviderRejected
	}

	p.cancelByKey[req.IdempotencyKey] = struct{}{}
	return nil
}

var _ port.BookingProvider = (*Provider)(nil)
