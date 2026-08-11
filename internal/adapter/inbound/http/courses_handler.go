package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func mapCourseToResponse(c entity.Course) CourseResponse {
	return CourseResponse{
		ID:        c.ID,
		Name:      c.Name,
		Street:    c.Street,
		City:      c.City,
		State:     c.State,
		ZipCode:   c.ZipCode,
		Phone:     c.Phone,
		Cost:      c.Cost,
		Country:   c.Country,
		Latitude:  c.Latitude,
		Longitude: c.Longitude,
		Timezone:  c.Timezone,
	}
}

func mapSearchHitToResponse(hit entity.CourseSearchResult) CourseSearchResponse {
	return CourseSearchResponse{
		CourseResponse: mapCourseToResponse(hit.Course),
		Provider:       hit.Provider,
		ExternalID:     hit.ExternalID,
	}
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courses.List(r.Context())
	if err != nil {
		respondInternalError(w, r, err, "Failed to fetch courses")
		return
	}

	resp := make([]CourseResponse, len(courses))
	for i, c := range courses {
		resp[i] = mapCourseToResponse(c)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) SearchCourses(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		respondJSON(w, http.StatusOK, []CourseSearchResponse{})
		return
	}

	hits, err := h.courses.Search(r.Context(), query)
	if err != nil {
		respondLoggedError(w, r, http.StatusBadGateway, "upstream_error", "Failed to search courses", err)
		return
	}

	resp := make([]CourseSearchResponse, 0, len(hits))
	for _, hit := range hits {
		resp = append(resp, mapSearchHitToResponse(hit))
	}

	respondJSON(w, http.StatusOK, resp)
}

type findOrCreateCourseRequest struct {
	CourseResponse
	Provider   string `json:"provider,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
}

func (h *Handler) FindOrCreateCourse(w http.ResponseWriter, r *http.Request) {
	var req findOrCreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	c := entity.Course{
		Name:      req.Name,
		Street:    req.Street,
		City:      req.City,
		State:     req.State,
		ZipCode:   req.ZipCode,
		Phone:     req.Phone,
		Cost:      req.Cost,
		Country:   req.Country,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timezone:  req.Timezone,
	}

	var link *entity.CourseProvider
	if req.Provider != "" && req.ExternalID != "" {
		link = &entity.CourseProvider{
			Provider:   req.Provider,
			ExternalID: req.ExternalID,
		}
	}

	result, created, err := h.courses.FindOrCreate(r.Context(), c, link)
	if err != nil {
		if errors.Is(err, entity.ErrProviderCourseConflict) {
			respondError(w, http.StatusConflict, "conflict", "Provider course id already linked to another course")
			return
		}
		respondInternalError(w, r, err, "Failed to create course")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	respondJSON(w, status, mapCourseToResponse(*result))
}
