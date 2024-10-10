package api

import (
	"net/http"

	"hw/pkg/bigrat"
	"hw/pkg/micro-tree/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// task represents a single task with a description and points.
type task struct {
	Description string  `json:"description"`
	Points      float64 `json:"points"`
}

// pool contains the total USD value, points, and associated tasks.
type pool struct {
	TotalUsdValue float64 `json:"total_usd_value"`
	Points        float64 `json:"points"`
	Tasks         []task  `json:"tasks"`
}

// response structures the JSON response with total values and pools.
type response struct {
	TotalUsdValue float64          `json:"total_usd_value"`
	TotalPoints   float64          `json:"total_points"`
	Pool          map[string]*pool `json:"pool"`
}

// GetUser handles retrieving a user's data.
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res := &response{
		Pool: make(map[string]*pool),
	}
	totalUsdValue := bigrat.NewBigN(0)

	user, err := s.Service.GetOrCreateAccount(r.Context(), id)
	if err != nil {
		render.Render(w, r, &errorResponse{Error: err.Error()})
		return
	}

	swapSummary, err := s.Service.GetUserSwapSummary(r.Context(), id)
	if err != nil {
		render.Render(w, r, &errorResponse{Error: err.Error()})
		return
	}

	for token, usdValue := range swapSummary {
		p, exists := res.Pool[token]
		if !exists {
			p = &pool{
				TotalUsdValue: 0,
				Points:        0,
				Tasks:         make([]task, 0),
			}
			res.Pool[token] = p
		}
		totalUsdValue = totalUsdValue.Add(usdValue)
		p.TotalUsdValue += usdValue

		pointsHistory, err := s.Service.GetPointsHistory(r.Context(), id, token)
		if err != nil {
			middleware.HTTPErrorLogging(w, r, err)
			render.Render(w, r, &errorResponse{Error: err.Error()})
			return
		}

		for _, points := range pointsHistory {
			p.Points += points.Points
			p.Tasks = append(p.Tasks, task{
				Description: points.Description,
				Points:      points.Points,
			})
		}
	}

	res.TotalPoints = user.TotalPoints
	res.TotalUsdValue = totalUsdValue.ToTruncateFloat64(6)

	render.JSON(w, r, res)
}
