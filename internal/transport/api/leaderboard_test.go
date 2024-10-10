package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hw/internal/model"
	"hw/internal/service/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestGetLeaderboard_Success tests the successful retrieval of the leaderboard.
func TestGetLeaderboard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	// Define mock user data
	users := []model.User{
		{
			Address:     "0xUser1",
			TotalPoints: 150.0,
		},
		{
			Address:     "0xUser2",
			TotalPoints: 200.0,
		},
		{
			Address:     "0xUser3",
			TotalPoints: 100.0,
		},
	}

	// Set expected behavior for the service layer
	mockService.EXPECT().GetLeaderboard(gomock.Any()).Return(users, nil)

	// Set up the router
	r := chi.NewRouter()
	r.Get("/leaderboard", server.GetLeaderboard)

	// Send HTTP request
	req, err := http.NewRequest("GET", "/leaderboard", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify HTTP status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response LeaderboardResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify that user data is sorted in descending order of points
	expected := LeaderboardResponse{
		Users: []UserPoints{
			{Address: "0xUser2", Points: 200.0},
			{Address: "0xUser1", Points: 150.0},
			{Address: "0xUser3", Points: 100.0},
		},
	}
	assert.Equal(t, expected, response)
}

func TestGetLeaderboard_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	expectedError := errors.New("database connection failed")

	mockService.EXPECT().GetLeaderboard(gomock.Any()).Return(nil, expectedError)

	r := chi.NewRouter()
	r.Get("/leaderboard", server.GetLeaderboard)

	req, err := http.NewRequest("GET", "/leaderboard", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify HTTP status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Compare the trimmed response content
	assert.Equal(t, expectedError.Error(), strings.TrimSpace(rr.Body.String()))
}

// TestGetLeaderboard_NoRows tests the scenario when no user data is found.
func TestGetLeaderboard_NoRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	// Set the service layer to return pgx.ErrNoRows
	mockService.EXPECT().GetLeaderboard(gomock.Any()).Return(nil, pgx.ErrNoRows)

	// Set up the router
	r := chi.NewRouter()
	r.Get("/leaderboard", server.GetLeaderboard)

	// Send HTTP request
	req, err := http.NewRequest("GET", "/leaderboard", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify HTTP status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response LeaderboardResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify that the user data is empty
	expected := LeaderboardResponse{
		Users: []UserPoints{},
	}
	assert.Equal(t, expected, response)
}

// TestGetLeaderboard_Sorting tests whether the leaderboard data is correctly sorted.
func TestGetLeaderboard_Sorting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := Server{
		Service: mockService,
	}

	// Define unsorted user data
	users := []model.User{
		{
			Address:     "0xUserA",
			TotalPoints: 120.0,
		},
		{
			Address:     "0xUserB",
			TotalPoints: 300.0,
		},
		{
			Address:     "0xUserC",
			TotalPoints: 50.0,
		},
		{
			Address:     "0xUserD",
			TotalPoints: 200.0,
		},
	}

	// Set expected behavior for the service layer
	mockService.EXPECT().GetLeaderboard(gomock.Any()).Return(users, nil)

	// Set up the router
	r := chi.NewRouter()
	r.Get("/leaderboard", server.GetLeaderboard)

	// Send HTTP request
	req, err := http.NewRequest("GET", "/leaderboard", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify HTTP status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response LeaderboardResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Define the expected sorted result
	expected := LeaderboardResponse{
		Users: []UserPoints{
			{Address: "0xUserB", Points: 300.0},
			{Address: "0xUserD", Points: 200.0},
			{Address: "0xUserA", Points: 120.0},
			{Address: "0xUserC", Points: 50.0},
		},
	}
	assert.Equal(t, expected, response)
}
