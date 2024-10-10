package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hw/internal/model"
	"hw/internal/service/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestGetUser_Success tests the successful retrieval of user data.
func TestGetUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	userID := "user123"
	user := &model.User{
		ID:          1,
		Address:     userID,
		TotalPoints: 150.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	swapSummary := map[string]float64{
		"tokenABC": 1000.50,
		"tokenXYZ": 500.25,
	}

	pointsHistoryABC := []model.PointsHistory{
		{
			Description: "Task 1",
			Points:      10.5,
			CreatedAt:   time.Now(),
		},
	}

	pointsHistoryXYZ := []model.PointsHistory{
		{
			Description: "Task 2",
			Points:      5.25,
			CreatedAt:   time.Now(),
		},
	}

	// Set expected service calls and return values
	mockService.EXPECT().
		GetOrCreateAccount(gomock.Any(), userID).
		Return(user, nil)

	mockService.EXPECT().
		GetUserSwapSummary(gomock.Any(), userID).
		Return(swapSummary, nil)

	mockService.EXPECT().
		GetPointsHistory(gomock.Any(), userID, "tokenABC").
		Return(pointsHistoryABC, nil)

	mockService.EXPECT().
		GetPointsHistory(gomock.Any(), userID, "tokenXYZ").
		Return(pointsHistoryXYZ, nil)

	server := Server{
		Service: mockService,
	}

	router := chi.NewRouter()
	router.Get("/user/{id}", server.GetUser)

	req, err := http.NewRequest("GET", "/user/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp response
	err = render.DecodeJSON(rr.Body, &resp)
	assert.NoError(t, err)

	assert.Equal(t, user.TotalPoints, resp.TotalPoints)
	assert.Equal(t, 1500.75, resp.TotalUsdValue)
	assert.Len(t, resp.Pool, 2)

	poolABC, exists := resp.Pool["tokenABC"]
	assert.True(t, exists)
	assert.Equal(t, 1000.50, poolABC.TotalUsdValue)
	assert.Equal(t, 10.5, poolABC.Points)
	assert.Len(t, poolABC.Tasks, 1)
	assert.Equal(t, "Task 1", poolABC.Tasks[0].Description)
	assert.Equal(t, 10.5, poolABC.Tasks[0].Points)

	poolXYZ, exists := resp.Pool["tokenXYZ"]
	assert.True(t, exists)
	assert.Equal(t, 500.25, poolXYZ.TotalUsdValue)
	assert.Equal(t, 5.25, poolXYZ.Points)
	assert.Len(t, poolXYZ.Tasks, 1)
	assert.Equal(t, "Task 2", poolXYZ.Tasks[0].Description)
	assert.Equal(t, 5.25, poolXYZ.Tasks[0].Points)
}

// TestGetUser_GetOrCreateAccountError tests the scenario when an error occurs while getting or creating a user account.
func TestGetUser_GetOrCreateAccountError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	userID := "user123"
	expectedError := model.ErrUserNotFound

	// Set expected service call and return error
	mockService.EXPECT().
		GetOrCreateAccount(gomock.Any(), userID).
		Return(nil, expectedError)

	server := Server{
		Service: mockService,
	}

	router := chi.NewRouter()
	router.Get("/user/{id}", server.GetUser)

	req, err := http.NewRequest("GET", "/user/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp errorResponse
	err = render.DecodeJSON(rr.Body, &errResp)
	assert.NoError(t, err)
	assert.Equal(t, expectedError.Error(), errResp.Error)
}

// TestGetUser_GetUserSwapSummaryError tests the scenario when an error occurs while retrieving the user swap summary.
func TestGetUser_GetUserSwapSummaryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	userID := "user123"
	user := &model.User{
		ID:          1,
		Address:     userID,
		TotalPoints: 150.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	expectedError := errors.New("failed to retrieve swap summary")

	// Set expected service calls and return values
	mockService.EXPECT().
		GetOrCreateAccount(gomock.Any(), userID).
		Return(user, nil)

	mockService.EXPECT().
		GetUserSwapSummary(gomock.Any(), userID).
		Return(nil, expectedError)

	server := Server{
		Service: mockService,
	}

	router := chi.NewRouter()
	router.Get("/user/{id}", server.GetUser)

	req, err := http.NewRequest("GET", "/user/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp errorResponse
	err = render.DecodeJSON(rr.Body, &errResp)
	assert.NoError(t, err)
	assert.Equal(t, expectedError.Error(), errResp.Error)
}

// TestGetUser_EmptySwapSummary tests the scenario when the swap summary is empty.
func TestGetUser_EmptySwapSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	userID := "user123"
	user := &model.User{
		ID:          1,
		Address:     userID,
		TotalPoints: 0.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	swapSummary := map[string]float64{}

	// Set expected service calls and return values
	mockService.EXPECT().
		GetOrCreateAccount(gomock.Any(), userID).
		Return(user, nil)

	mockService.EXPECT().
		GetUserSwapSummary(gomock.Any(), userID).
		Return(swapSummary, nil)

	server := Server{
		Service: mockService,
	}

	router := chi.NewRouter()
	router.Get("/user/{id}", server.GetUser)

	req, err := http.NewRequest("GET", "/user/"+userID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp response
	err = render.DecodeJSON(rr.Body, &resp)
	assert.NoError(t, err)

	assert.Equal(t, user.TotalPoints, resp.TotalPoints)
	assert.Equal(t, 0.0, resp.TotalUsdValue)
	assert.Empty(t, resp.Pool)
}
