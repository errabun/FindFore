package handler

import (
	"net/http"

	"github.com/ericrabun/findfore-go/internal/model"
)

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.queries.ListCourses(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch courses")
		return
	}

	resp := make([]model.CourseResponse, len(courses))
	for i, c := range courses {
		resp[i] = model.CourseResponse{
			ID:      c.ID,
			Name:    c.Name.String,
			Street:  c.Street.String,
			City:    c.City.String,
			State:   c.State.String,
			ZipCode: c.ZipCode.String,
			Phone:   c.Phone.String,
			Cost:    c.Cost.String,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}
