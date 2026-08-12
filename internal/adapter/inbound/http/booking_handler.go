package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type teeTimeResponse struct {
	ID             int64   `json:"id"`
	StartsAt       string  `json:"starts_at"`
	Holes          string  `json:"holes,omitempty"`
	AvailableSlots *int32  `json:"available_slots,omitempty"`
	PriceCents     *int32  `json:"price_cents,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	LastSyncedAt   *string `json:"last_synced_at,omitempty"`
}

type reservationPlayerResponse struct {
	PlayerID  *int64 `json:"player_id,omitempty"`
	GuestName string `json:"guest_name,omitempty"`
}

type reservationResponse struct {
	ID               int64                       `json:"id"`
	TeeTimeID        int64                       `json:"tee_time_id"`
	Status           string                      `json:"status"`
	PartySize        int32                       `json:"party_size"`
	QuotedPriceCents *int32                      `json:"quoted_price_cents,omitempty"`
	QuotedCurrency   string                      `json:"quoted_currency,omitempty"`
	HoldExpiresAt    *string                     `json:"hold_expires_at,omitempty"`
	Players          []reservationPlayerResponse `json:"players"`
}

type beginReservationRequest struct {
	TeeTimeID int64 `json:"tee_time_id"`
	Players   []struct {
		PlayerID  *int64 `json:"player_id"`
		GuestName string `json:"guest_name"`
	} `json:"players"`
}

func mapTeeTime(t entity.TeeTime) teeTimeResponse {
	out := teeTimeResponse{
		ID:             t.ID,
		StartsAt:       t.StartsAt.UTC().Format(time.RFC3339),
		Holes:          t.Holes,
		AvailableSlots: t.AvailableSlots,
		PriceCents:     t.PriceCents,
		Currency:       t.Currency,
	}
	if t.LastSyncedAt != nil {
		s := t.LastSyncedAt.UTC().Format(time.RFC3339)
		out.LastSyncedAt = &s
	}
	return out
}

func mapReservation(res *entity.Reservation, players []entity.ReservationPlayer) reservationResponse {
	out := reservationResponse{
		ID:               res.ID,
		TeeTimeID:        res.TeeTimeID,
		Status:           res.Status,
		PartySize:        res.PartySize,
		QuotedPriceCents: res.QuotedPriceCents,
		QuotedCurrency:   res.QuotedCurrency,
		Players:          make([]reservationPlayerResponse, 0, len(players)),
	}
	if res.HoldExpiresAt != nil {
		s := res.HoldExpiresAt.UTC().Format(time.RFC3339)
		out.HoldExpiresAt = &s
	}
	for _, p := range players {
		out.Players = append(out.Players, reservationPlayerResponse{
			PlayerID:  p.PlayerID,
			GuestName: p.GuestName,
		})
	}
	return out
}

func (h *Handler) respondReservation(w http.ResponseWriter, r *http.Request, status int, res *entity.Reservation) {
	players, err := h.booking.ListReservationPlayers(r.Context(), res.ID)
	if err != nil {
		respondInternalError(w, r, err, "Failed to load reservation players")
		return
	}
	respondJSON(w, status, mapReservation(res, players))
}

func (h *Handler) ListCourseTeeTimes(w http.ResponseWriter, r *http.Request) {
	if h.booking == nil {
		respondError(w, http.StatusNotImplemented, "not_implemented", "Booking is not configured")
		return
	}
	if _, ok := mw.PlayerIDFromContext(r.Context()); !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil || courseID <= 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid course id")
		return
	}
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "from and to query params are required")
		return
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "from must be RFC3339")
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "to must be RFC3339")
		return
	}
	var minPlayers int32
	if p := r.URL.Query().Get("players"); p != "" {
		n, err := strconv.ParseInt(p, 10, 32)
		if err != nil || n < 0 {
			respondError(w, http.StatusBadRequest, "validation_error", "players must be a non-negative integer")
			return
		}
		minPlayers = int32(n)
	}

	tees, err := h.booking.SearchAvailability(r.Context(), courseID, from, to, minPlayers)
	if err != nil {
		writeBookingError(w, r, err)
		return
	}
	items := make([]teeTimeResponse, len(tees))
	for i, t := range tees {
		items[i] = mapTeeTime(t)
	}
	respondJSON(w, http.StatusOK, map[string]any{"tee_times": items})
}

func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	if h.booking == nil {
		respondError(w, http.StatusNotImplemented, "not_implemented", "Booking is not configured")
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	var req beginReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid JSON body")
		return
	}
	if req.TeeTimeID <= 0 || len(req.Players) == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "tee_time_id and players are required")
		return
	}
	players := make([]entity.ReservationPlayer, 0, len(req.Players))
	for _, p := range req.Players {
		if p.PlayerID == nil && p.GuestName == "" {
			respondError(w, http.StatusBadRequest, "validation_error", "each player needs player_id or guest_name")
			return
		}
		players = append(players, entity.ReservationPlayer{PlayerID: p.PlayerID, GuestName: p.GuestName})
	}

	out, err := h.booking.BeginBooking(r.Context(), port.BeginBookingInput{
		ActorID: actorID, TeeTimeID: req.TeeTimeID, Players: players,
	})
	if err != nil {
		writeBookingError(w, r, err)
		return
	}
	status := http.StatusOK
	if out.Created {
		status = http.StatusCreated
	}
	h.respondReservation(w, r, status, out.Reservation)
}

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	h.mutateReservation(w, r, func(ctx context.Context, actorID, id int64) (*entity.Reservation, error) {
		return h.booking.ConfirmBooking(ctx, actorID, id)
	})
}

func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	h.mutateReservation(w, r, func(ctx context.Context, actorID, id int64) (*entity.Reservation, error) {
		return h.booking.CancelBooking(ctx, actorID, id)
	})
}

func (h *Handler) mutateReservation(
	w http.ResponseWriter,
	r *http.Request,
	fn func(ctx context.Context, actorID, id int64) (*entity.Reservation, error),
) {
	if h.booking == nil {
		respondError(w, http.StatusNotImplemented, "not_implemented", "Booking is not configured")
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid reservation id")
		return
	}
	res, err := fn(r.Context(), actorID, id)
	if err != nil {
		writeBookingError(w, r, err)
		return
	}
	h.respondReservation(w, r, http.StatusOK, res)
}

func writeBookingError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, booking.ErrProviderOutcomeUnknown):
		respondLoggedError(w, r, http.StatusServiceUnavailable, "provider_outcome_unknown", "Provider outcome unknown; retry the same reservation", err)
	case errors.Is(err, booking.ErrProviderRejected):
		respondError(w, http.StatusConflict, "provider_rejected", "Provider rejected the booking request")
	case errors.Is(err, entity.ErrReservationForbidden):
		respondError(w, http.StatusForbidden, "forbidden", "Not allowed to modify this reservation")
	case errors.Is(err, entity.ErrReservationNotFound), errors.Is(err, booking.ErrTeeTimeNotFound), errors.Is(err, booking.ErrCourseNotFound):
		respondError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, entity.ErrActiveReservationExists),
		errors.Is(err, entity.ErrInvalidReservationTransition),
		errors.Is(err, booking.ErrReservationConflict):
		respondError(w, http.StatusConflict, "conflict", "Reservation conflict")
	case errors.Is(err, booking.ErrInvalidParty), errors.Is(err, booking.ErrProviderLinkMissing):
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid booking request")
	case errors.Is(err, booking.ErrProviderRequired):
		respondError(w, http.StatusNotImplemented, "not_implemented", "Booking provider is not configured")
	default:
		respondInternalError(w, r, err, "Booking request failed")
	}
}
