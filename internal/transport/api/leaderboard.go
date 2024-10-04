package api

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// UserPoints represents a user's address and their points.
type UserPoints struct {
	Address string  `json:"address"`
	Points  float64 `json:"points"`
}

// leaderboardResponse represents the response structure for the leaderboard.
type leaderboardResponse struct {
	Users []UserPoints `json:"users"`
}

// GetLeaderboard retrieves the leaderboard data and returns it as JSON.
func (s Server) GetLeaderboard(c *fiber.Ctx) error {
	// Fetch users from the domain
	users, err := s.Service.GetLeaderboard(c.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty response if no rows are found
			return c.JSON(&leaderboardResponse{
				Users: []UserPoints{},
			})
		}
		// Return internal server error for other errors
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Initialize response with preallocated capacity
	res := &leaderboardResponse{
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

	return c.JSON(res)
}
