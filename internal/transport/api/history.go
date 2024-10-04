package api

import (
	"github.com/gofiber/fiber/v2"
)

type historyTask struct {
	Description string  `json:"description"`
	Points      float64 `json:"points"`
	CreatedAt   string  `json:"created_at"`
}

type historyResponse struct {
	Tasks map[string][]historyTask `json:"tasks"`
}

// GetHistory handles fetching the user's history
func (s Server) GetHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	res := &historyResponse{
		Tasks: make(map[string][]historyTask),
	}

	// Get user swap summary
	swapSummary, err := s.Service.GetUserSwapSummary(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Iterate over each token in swap summary
	for token := range swapSummary {
		pointsHistory, err := s.Service.GetPointsHistory(ctx, id, token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		for _, points := range pointsHistory {
			res.Tasks[token] = append(res.Tasks[token], historyTask{
				Description: points.Description,
				Points:      points.Points,
				CreatedAt:   points.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return c.JSON(res)
}
