package api

import (
	"net/http"
	"sort"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// UserPoints represents a user's address and their points.
type UserPoints struct {
	Address string  `json:"address"`
	Points  float64 `json:"points"`
}

// LeaderboardResponse represents the response structure for the leaderboard.
type LeaderboardResponse struct {
	Users []UserPoints `json:"users"`
}

// GetLeaderboard retrieves the leaderboard data and returns it as JSON.
func (s *Server) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Fetch users from the domain
	users, err := s.Service.GetLeaderboard(r.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty response if no rows are found
			res := LeaderboardResponse{
				Users: []UserPoints{},
			}
			render.JSON(w, r, res)
			return
		}
		// Return internal server error for other errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize response with preallocated capacity
	res := LeaderboardResponse{
		Users: make([]UserPoints, 0, len(users)),
	}

	// Populate Users slice
	for _, user := range users {
		res.Users = append(res.Users, UserPoints{
			Address: user.Address,
			Points:  user.TotalPoints,
		})
	}

	// Sort users by points in descending order
	sort.Slice(res.Users, func(i, j int) bool {
		return res.Users[i].Points > res.Users[j].Points
	})

	// Respond with the sorted leaderboard
	render.JSON(w, r, res)
}
