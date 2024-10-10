package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hw/internal/model"
	"hw/internal/service/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestGetHistory_Success tests the successful retrieval of history records.
// TestGetHistory_NoTokens tests the scenario when the user has no swap summaries (i.e., no tokens).
func TestGetHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	userID := "user123"
	token := "tokenABC"
	swapSummary := map[string]float64{
		token: 100.0,
	}
	pointsHistory := []model.PointsHistory{
		{
			Description: "Task 1",
			Points:      10.5,
			CreatedAt:   time.Now(),
		},
	}

	mockService.
		EXPECT().
		GetUserSwapSummary(gomock.Any(), userID).
		Return(swapSummary, nil)

	mockService.
		EXPECT().
		GetPointsHistory(gomock.Any(), userID, token).
		Return(pointsHistory, nil)

	r := chi.NewRouter()
	r.Get("/history/{id}", server.GetHistory)

	req, err := http.NewRequest("GET", "/history/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response historyResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.Tasks, token)
	assert.Equal(t, 1, len(response.Tasks[token]))
	assert.Equal(t, "Task 1", response.Tasks[token][0].Description)
	assert.Equal(t, 10.5, response.Tasks[token][0].Points)
}

// TestGetHistory_NoTokens tests the scenario when the user has no swap summaries (i.e., no tokens).
func TestGetHistory_NoTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	userID := "user123"
	swapSummary := map[string]float64{}

	mockService.
		EXPECT().
		GetUserSwapSummary(gomock.Any(), userID).
		Return(swapSummary, nil)

	r := chi.NewRouter()
	r.Get("/history/{id}", server.GetHistory)

	req, err := http.NewRequest("GET", "/history/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response historyResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Empty(t, response.Tasks)
}
