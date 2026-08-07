package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func mapEventToResponse(e entity.EventWithDetails) EventResponse {
	accepted := e.Accepted
	if accepted == nil {
		accepted = []int64{}
	}
	declined := e.Declined
	if declined == nil {
		declined = []int64{}
	}
	pending := e.Pending
	if pending == nil {
		pending = []int64{}
	}
	closed := e.Closed
	if closed == nil {
		closed = []int64{}
	}
	return EventResponse{
		ID:             e.ID,
		CourseName:     e.CourseName,
		Date:           e.Date,
		TeeTime:        e.TeeTime,
		OpenSpots:      e.OpenSpots,
		NumberOfHoles:  e.NumberOfHoles,
		Private:        e.Private,
		HostName:       e.HostName,
		HostID:         e.HostID,
		Accepted:       accepted,
		Declined:       declined,
		Pending:        pending,
		Closed:         closed,
		RemainingSpots: e.RemainingSpots,
	}
}

func eventErrorStatus(err error) (int, string, string) {
	switch {
	case errors.Is(err, events.ErrEventNotFound):
		return http.StatusNotFound, "not_found", "Event not found"
	case errors.Is(err, events.ErrEventForbidden):
		return http.StatusForbidden, "forbidden", "Not allowed to access this event"
	default:
		return http.StatusInternalServerError, "internal_error", "Event operation failed"
	}
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	playerIDStr := chi.URLParam(r, "player_id")
	privateParam := r.URL.Query().Get("private")
	playerIDQuery := r.URL.Query().Get("player_id")

	var forPlayerID *int64
	publicOnly := privateParam == "false"

	if playerIDStr != "" {
		pid, err := strconv.ParseInt(playerIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
			return
		}
		forPlayerID = &pid
	} else if playerIDQuery != "" {
		pid, err := strconv.ParseInt(playerIDQuery, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
			return
		}
		forPlayerID = &pid
	}

	events, err := h.events.List(r.Context(), actorID, forPlayerID, publicOnly)
	if err != nil {
		status, code, msg := eventErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	resp := make([]EventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, mapEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid event ID")
		return
	}

	event, err := h.events.Get(r.Context(), id, actorID)
	if err != nil {
		status, code, msg := eventErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, mapEventToResponse(*event))
}

type createEventRequest struct {
	CourseID      json.Number `json:"course_id"`
	Date          string      `json:"date"`
	TeeTime       string      `json:"tee_time"`
	OpenSpots     json.Number `json:"open_spots"`
	NumberOfHoles string      `json:"number_of_holes"`
	Private       bool        `json:"private"`
	HostID        int64       `json:"host_id"` // ignored; host is the authenticated player
	Invitees      []int64     `json:"invitees"`
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

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

	e := entity.Event{
		CourseID:      int32(courseID),
		Date:          req.Date,
		TeeTime:       req.TeeTime,
		OpenSpots:     int32(openSpots),
		NumberOfHoles: req.NumberOfHoles,
		Private:       req.Private,
		HostID:        int32(actorID),
	}

	event, err := h.events.Create(r.Context(), e, req.Invitees)
	if err != nil {
		respondInternalError(w, r, err, "Failed to create event")
		return
	}

	respondJSON(w, http.StatusCreated, mapEventToResponse(*event))
}

type updateEventRequest struct {
	CourseID      json.Number `json:"course_id"`
	Date          string      `json:"date"`
	TeeTime       string      `json:"tee_time"`
	OpenSpots     json.Number `json:"open_spots"`
	NumberOfHoles string      `json:"number_of_holes"`
	Private       bool        `json:"private"`
	Invitees      []int64     `json:"invitees"`
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid event ID")
		return
	}

	var req updateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	courseID, err := req.CourseID.Int64()
	if err != nil || courseID == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Course can't be blank")
		return
	}

	openSpots, err := req.OpenSpots.Int64()
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Open spots can't be blank")
		return
	}

	e := entity.Event{
		ID:            id,
		CourseID:      int32(courseID),
		Date:          req.Date,
		TeeTime:       req.TeeTime,
		OpenSpots:     int32(openSpots),
		NumberOfHoles: req.NumberOfHoles,
		Private:       req.Private,
	}

	event, err := h.events.Update(r.Context(), actorID, e, req.Invitees)
	if err != nil {
		status, code, msg := eventErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, mapEventToResponse(*event))
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid event ID")
		return
	}

	if err := h.events.Delete(r.Context(), actorID, id); err != nil {
		status, code, msg := eventErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

func (h *Handler) ListFriendsEvents(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	playerIDStr := chi.URLParam(r, "player_id")
	pid, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
		return
	}
	if pid != actorID {
		respondError(w, http.StatusForbidden, "forbidden", "Not allowed to access another player's friends events")
		return
	}

	events, err := h.events.ListFriendsEvents(r.Context(), actorID)
	if err != nil {
		respondInternalError(w, r, err, "Failed to fetch friends events")
		return
	}

	resp := make([]EventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, mapEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, resp)
}
