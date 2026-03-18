package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
)

// invite_status enum: 0=pending, 1=accepted, 2=declined, 3=closed
const (
	statusPending  int32 = 0
	statusAccepted int32 = 1
	statusDeclined int32 = 2
	statusClosed   int32 = 3
)

func (h *Handler) buildEventResponse(r *http.Request, eventID int64) (*model.EventResponse, error) {
	event, err := h.queries.GetEventByID(r.Context(), eventID)
	if err != nil {
		return nil, err
	}

	accepted, err := h.queries.ListPlayerIDsByEventAndStatus(r.Context(), store.ListPlayerIDsByEventAndStatusParams{
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: statusAccepted, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	declined, err := h.queries.ListPlayerIDsByEventAndStatus(r.Context(), store.ListPlayerIDsByEventAndStatusParams{
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: statusDeclined, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	pending, err := h.queries.ListPlayerIDsByEventAndStatus(r.Context(), store.ListPlayerIDsByEventAndStatusParams{
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: statusPending, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	closed, err := h.queries.ListPlayerIDsByEventAndStatus(r.Context(), store.ListPlayerIDsByEventAndStatusParams{
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: statusClosed, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	acceptedIDs := toInt64Slice(accepted)
	declinedIDs := toInt64Slice(declined)
	pendingIDs := toInt64Slice(pending)
	closedIDs := toInt64Slice(closed)

	remainingSpots := event.OpenSpots.Int32 - int32(len(acceptedIDs))

	return &model.EventResponse{
		ID:             event.ID,
		CourseName:     event.CourseName.String,
		Date:           event.Date.String,
		TeeTime:        event.TeeTime.String,
		OpenSpots:      event.OpenSpots.Int32,
		NumberOfHoles:  event.NumberOfHoles.String,
		Private:        event.Private.Bool,
		HostName:       event.HostName.String,
		HostID:         event.HostID.Int32,
		Accepted:       acceptedIDs,
		Declined:       declinedIDs,
		Pending:        pendingIDs,
		Closed:         closedIDs,
		RemainingSpots: remainingSpots,
	}, nil
}

func toInt64Slice(nullIDs []sql.NullInt64) []int64 {
	ids := make([]int64, 0, len(nullIDs))
	for _, nid := range nullIDs {
		if nid.Valid {
			ids = append(ids, nid.Int64)
		}
	}
	return ids
}

type eventRow interface {
	getID() int64
}

type listAllEventsRowAdapter struct{ store.ListAllEventsRow }

func (a listAllEventsRowAdapter) getID() int64 { return a.ID }

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	// Check for player_id path param (nested route: /api/v1/players/:player_id/events)
	playerIDStr := chi.URLParam(r, "player_id")

	// Check for query params
	privateParam := r.URL.Query().Get("private")
	playerIDQuery := r.URL.Query().Get("player_id")

	var eventIDs []int64

	if playerIDStr != "" {
		// Nested route: GET /api/v1/players/:player_id/events
		pid, err := strconv.ParseInt(playerIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
			return
		}
		events, err := h.queries.ListEventsByPlayerID(r.Context(), sql.NullInt64{Int64: pid, Valid: true})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch events")
			return
		}
		eventIDs = make([]int64, len(events))
		for i, e := range events {
			eventIDs[i] = e.ID
		}
	} else if playerIDQuery != "" {
		// Query param: GET /api/v1/events?player_id=X
		pid, err := strconv.ParseInt(playerIDQuery, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
			return
		}
		events, err := h.queries.ListEventsByPlayerID(r.Context(), sql.NullInt64{Int64: pid, Valid: true})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch events")
			return
		}
		eventIDs = make([]int64, len(events))
		for i, e := range events {
			eventIDs[i] = e.ID
		}
	} else if privateParam == "false" {
		// Query param: GET /api/v1/events?private=false
		events, err := h.queries.ListPublicEvents(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch events")
			return
		}
		eventIDs = make([]int64, len(events))
		for i, e := range events {
			eventIDs[i] = e.ID
		}
	} else {
		// Default: all events
		events, err := h.queries.ListAllEvents(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch events")
			return
		}
		eventIDs = make([]int64, len(events))
		for i, e := range events {
			eventIDs[i] = e.ID
		}
	}

	resp := make([]model.EventResponse, 0, len(eventIDs))
	for _, eid := range eventIDs {
		er, err := h.buildEventResponse(r, eid)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to build event response")
			return
		}
		resp = append(resp, *er)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid event ID")
		return
	}

	resp, err := h.buildEventResponse(r, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Event not found")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

type createEventRequest struct {
	CourseID      json.Number `json:"course_id"`
	Date          string      `json:"date"`
	TeeTime       string      `json:"tee_time"`
	OpenSpots     json.Number `json:"open_spots"`
	NumberOfHoles string      `json:"number_of_holes"`
	Private       bool        `json:"private"`
	HostID        int64       `json:"host_id"`
	Invitees      []int64     `json:"invitees"`
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	courseID, err := req.CourseID.Int64()
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Course can't be blank")
		return
	}
	if courseID == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Course can't be blank")
		return
	}

	openSpots, err := req.OpenSpots.Int64()
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Open spots can't be blank")
		return
	}

	if req.Date == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Date can't be blank")
		return
	}
	if req.TeeTime == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Tee time can't be blank")
		return
	}
	if req.NumberOfHoles == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Number of holes can't be blank")
		return
	}
	if req.HostID == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Host can't be blank")
		return
	}

	eventID, err := store.CreateEventWithInvites(r.Context(), h.db, h.queries, store.CreateEventWithInvitesParams{
		CourseID:      int32(courseID),
		Date:          req.Date,
		TeeTime:       req.TeeTime,
		OpenSpots:     int32(openSpots),
		NumberOfHoles: req.NumberOfHoles,
		Private:       req.Private,
		HostID:        int32(req.HostID),
		Invitees:      req.Invitees,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create event")
		return
	}

	resp, err := h.buildEventResponse(r, eventID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to build event response")
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) ListFriendsEvents(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "player_id")
	pid, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
		return
	}

	eventIDs, err := h.queries.ListFriendsAvailableEventIDs(r.Context(), store.ListFriendsAvailableEventIDsParams{
		FollowerID: sql.NullInt32{Int32: int32(pid), Valid: true},
		PlayerID:   sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch friends events")
		return
	}

	resp := make([]model.EventResponse, 0, len(eventIDs))
	for _, eid := range eventIDs {
		er, err := h.buildEventResponse(r, eid)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to build event response")
			return
		}
		resp = append(resp, *er)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid event ID")
		return
	}

	if err := h.queries.DeleteEvent(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Event not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
