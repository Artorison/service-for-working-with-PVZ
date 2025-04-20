package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pvz/internal/models"
	"pvz/internal/services/mocks"
)

func newSvc() (*mocks.PvzUserStore, *Service) {
	repo := new(mocks.PvzUserStore)
	return repo, NewService(repo)
}

func TestServiceRegisterUserErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		RegisterUser("user@example.com", mock.AnythingOfType("string"), models.Employee).
		Return(models.User{}, errors.New("db fail")).Once()

	_, err := svc.RegisterUser("user@example.com", "password123", models.Employee)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to register user")
	repo.AssertExpectations(t)
}

func TestServiceRegisterUserSuccess(t *testing.T) {
	repo, svc := newSvc()

	want := models.User{ID: "u1", Email: "user@example.com", Role: models.Employee}
	repo.EXPECT().
		RegisterUser("user@example.com", mock.AnythingOfType("string"), models.Employee).
		Return(want, nil).Once()

	got, err := svc.RegisterUser("user@example.com", "password123", models.Employee)
	require.NoError(t, err)
	require.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestServiceLoginUserErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		LoginUser("user@example.com", "password123").
		Return(models.Token(""), errors.New("db fail")).Once()

	_, err := svc.LoginUser("user@example.com", "password123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceLoginUserSuccess(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		LoginUser("user@example.com", "password123").
		Return(models.Token("tok-123"), nil).Once()

	got, err := svc.LoginUser("user@example.com", "password123")
	require.NoError(t, err)
	require.Equal(t, models.Token("tok-123"), got)
	repo.AssertExpectations(t)
}

func TestServiceCreatePVZErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		CreatePVZ(models.Moscow).
		Return(models.PVZ{}, errors.New("db fail")).Once()

	_, err := svc.CreatePVZ(models.Moscow)
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceCreatePVZSuccess(t *testing.T) {
	repo, svc := newSvc()

	want := models.PVZ{ID: "uuid-123", RegistrationDate: time.Now().UTC(), City: models.Moscow}
	repo.EXPECT().
		CreatePVZ(models.Moscow).
		Return(want, nil).Once()

	got, err := svc.CreatePVZ(models.Moscow)
	require.NoError(t, err)
	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.City, got.City)
	repo.AssertExpectations(t)
}

func TestServiceCreateReceptionErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		CreateReception("uuid-123").
		Return(models.Reception{}, errors.New("db fail")).Once()

	_, err := svc.CreateReception("uuid-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceCreateReceptionSuccess(t *testing.T) {
	repo, svc := newSvc()

	want := models.Reception{ID: "rec-1", PvzID: "uuid-123"}
	repo.EXPECT().
		CreateReception("uuid-123").
		Return(want, nil).Once()

	got, err := svc.CreateReception("uuid-123")
	require.NoError(t, err)
	require.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestServiceCreateProductErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		CreateProduct("uuid-123", models.Electronic).
		Return(models.Product{}, errors.New("db fail")).Once()

	_, err := svc.CreateProduct("uuid-123", models.Electronic)
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceCreateProductSuccess(t *testing.T) {
	repo, svc := newSvc()

	want := models.Product{ID: "prod-1", ReceptionID: "r1", Type: models.Electronic}
	repo.EXPECT().
		CreateProduct("uuid-123", models.Electronic).
		Return(want, nil).Once()

	got, err := svc.CreateProduct("uuid-123", models.Electronic)
	require.NoError(t, err)
	require.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestServiceCloseLastReceptionErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		CloseLastReception("uuid-123").
		Return(models.Reception{}, errors.New("db fail")).Once()

	_, err := svc.CloseLastReception("uuid-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceCloseLastReceptionSuccess(t *testing.T) {
	repo, svc := newSvc()

	want := models.Reception{ID: "r2", Status: models.StatusClose}
	repo.EXPECT().
		CloseLastReception("uuid-123").
		Return(want, nil).Once()

	got, err := svc.CloseLastReception("uuid-123")
	require.NoError(t, err)
	require.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestServiceDeleteLastProductErrors(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		DeleteLastProduct("uuid-123").
		Return(errors.New("db fail")).Once()

	err := svc.DeleteLastProduct("uuid-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "db fail")
	repo.AssertExpectations(t)
}

func TestServiceDeleteLastProductSuccess(t *testing.T) {
	repo, svc := newSvc()

	repo.EXPECT().
		DeleteLastProduct("uuid-123").
		Return(nil).Once()

	err := svc.DeleteLastProduct("uuid-123")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceGetPVZInfoErrors(t *testing.T) {
	_, svc := newSvc()

	_, err := svc.GetPVZInfo("bad-date", "", 0, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid startDate")

	valid := time.Now().UTC().Format(time.RFC3339)
	_, err = svc.GetPVZInfo(valid, "bad-date", 0, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid endDate")
}

func TestServiceGetPVZInfoSuccess(t *testing.T) {
	repo, svc := newSvc()

	startStr := "2025-01-01T00:00:00Z"
	endStr := "2025-01-02T00:00:00Z"
	page := 2
	limit := 5

	startTime, _ := time.Parse(time.RFC3339, startStr)
	endTime, _ := time.Parse(time.RFC3339, endStr)

	want := []models.PVZInfo{{Pvz: models.PVZ{ID: "1"}}}
	repo.EXPECT().
		GetPVZInfo(startTime, endTime, page, limit).
		Return(want, nil).Once()

	got, err := svc.GetPVZInfo(startStr, endStr, page, limit)
	require.NoError(t, err)
	require.Equal(t, want, got)
	repo.AssertExpectations(t)
}
