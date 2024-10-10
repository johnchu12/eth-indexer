package api

import (
	"net/http"

	"hw/pkg/micro-tree/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// historyTask represents a single task with a description and points.
type historyTask struct {
	Description string  `json:"description"`
	Points      float64 `json:"points"`
	CreatedAt   string  `json:"created_at"`
}

// historyResponse structures the JSON response with tasks categorized by tokens.
type historyResponse struct {
	Tasks map[string][]historyTask `json:"tasks"`
}

// GetHistory handles fetching the user's history.
func (s Server) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res := &historyResponse{
		Tasks: make(map[string][]historyTask),
	}

	// Get user swap summary
	swapSummary, err := s.Service.GetUserSwapSummary(r.Context(), id)
	if err != nil {
		middleware.HTTPErrorLogging(w, r, err)
		render.Render(w, r, &errorResponse{Error: err.Error()})
		return
	}

	// Iterate over each token in swap summary
	for token := range swapSummary {
		pointsHistory, err := s.Service.GetPointsHistory(r.Context(), id, token)
		if err != nil {
			middleware.HTTPErrorLogging(w, r, err)
			// handleError(w, http.StatusInternalServerError, err)
			render.Render(w, r, &errorResponse{Error: err.Error()})
			return
		}

		for _, points := range pointsHistory {
			res.Tasks[token] = append(res.Tasks[token], historyTask{
				Description: points.Description,
				Points:      points.Points,
				CreatedAt:   points.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	render.JSON(w, r, res)
}
