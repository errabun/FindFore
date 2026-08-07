package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func mapCourseToResponse(c entity.Course) CourseResponse {
	return CourseResponse{
		ID:      c.ID,
		Name:    c.Name,
		Street:  c.Street,
		City:    c.City,
		State:   c.State,
		ZipCode: c.ZipCode,
		Phone:   c.Phone,
		Cost:    c.Cost,
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
		respondJSON(w, http.StatusOK, []CourseResponse{})
		return
	}

	courses, err := h.courses.Search(r.Context(), query)
	if err != nil {
		respondLoggedError(w, r, http.StatusBadGateway, "upstream_error", "Failed to search courses", err)
		return
	}

	resp := make([]CourseResponse, 0, len(courses))
	for _, c := range courses {
		resp = append(resp, mapCourseToResponse(c))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) FindOrCreateCourse(w http.ResponseWriter, r *http.Request) {
	var req CourseResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	c := entity.Course{
		Name:    req.Name,
		Street:  req.Street,
		City:    req.City,
		State:   req.State,
		ZipCode: req.ZipCode,
		Phone:   req.Phone,
		Cost:    req.Cost,
	}

	result, err := h.courses.FindOrCreate(r.Context(), c)
	if err != nil {
		respondInternalError(w, r, err, "Failed to create course")
		return
	}

	resp := mapCourseToResponse(*result)

	// If the course already had an ID, it was found (200); otherwise created (201)
	status := http.StatusCreated
	if req.Name == result.Name && result.ID != 0 && req.ID == 0 {
		// FindOrCreate found an existing one
		status = http.StatusOK
	}

	respondJSON(w, status, resp)
}
