package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz/internal/handlers/mocks"
	"pvz/internal/models"
	"pvz/internal/validation"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeValidator struct{}

func (v *fakeValidator) Validate(i any) error { return nil }

func setup() (*echo.Echo, *mocks.PvzUserService) {
	e := echo.New()
	e.Validator = &fakeValidator{}
	svc := new(mocks.PvzUserService)
	return e, svc
}

func TestDummyLogin(t *testing.T) {
	e, _ := setup()
	h := NewHandler(nil)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString("{bad}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.DummyLogin(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		body := `{"role":"employee"}`
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.DummyLogin(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var tok string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
		require.NotEmpty(t, tok)
	})
}

func TestRegisterUser(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.RegisterUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			RegisterUser("a@b.com", "pass", models.Employee).
			Return(models.User{}, errors.New("fail")).
			Once()

		reqBody, _ := json.Marshal(validation.RegisterRequest{Email: "a@b.com", Password: "pass", Role: models.Employee})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.RegisterUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		user := models.User{ID: "u1", Email: "a@b.com", Role: models.Employee}
		svc.EXPECT().
			RegisterUser("a@b.com", "pass", models.Employee).
			Return(user, nil).
			Once()

		reqBody, _ := json.Marshal(validation.RegisterRequest{Email: "a@b.com", Password: "pass", Role: models.Employee})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.RegisterUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)

		var got models.User
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, user, got)
	})
}

func TestLoginUser(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("bad"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.LoginUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			LoginUser("a@b.com", "pass").
			Return(models.Token(""), errors.New("denied")).
			Once()

		reqBody, _ := json.Marshal(validation.LoginRequest{Email: "a@b.com", Password: "pass"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.LoginUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		token := models.Token("tok123")
		svc.EXPECT().
			LoginUser("a@b.com", "pass").
			Return(token, nil).
			Once()

		reqBody, _ := json.Marshal(validation.LoginRequest{Email: "a@b.com", Password: "pass"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.LoginUser(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var got string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, string(token), got)
	})
}

func TestCreatePVZ(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBufferString("oops"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreatePVZ(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			CreatePVZ(models.Moscow).
			Return(models.PVZ{}, errors.New("nope")).
			Once()

		reqBody, _ := json.Marshal(validation.CreatePVZRequest{City: models.Moscow})
		req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreatePVZ(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		pvz := models.PVZ{ID: "p1", City: models.Kazan}
		svc.EXPECT().
			CreatePVZ(models.Kazan).
			Return(pvz, nil).
			Once()

		reqBody, _ := json.Marshal(validation.CreatePVZRequest{City: models.Kazan})
		req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreatePVZ(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)

		var got models.PVZ
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, pvz, got)
	})
}

func TestCreateReception(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			CreateReception("pvz1").
			Return(models.Reception{}, errors.New("err")).
			Once()

		reqBody, _ := json.Marshal(validation.CreateReceptionRequest{PvzID: "pvz1"})
		req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreateReception(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		recp := models.Reception{ID: "r1", PvzID: "pvz1", Status: models.StatusInProgress}
		svc.EXPECT().
			CreateReception("pvz1").
			Return(recp, nil).
			Once()

		reqBody, _ := json.Marshal(validation.CreateReceptionRequest{PvzID: "pvz1"})
		req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreateReception(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)

		var got models.Reception
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, recp, got)
	})
}

func TestCreateProduct(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			CreateProduct("pvz1", models.Electronic).
			Return(models.Product{}, errors.New("fail")).
			Once()

		reqBody, _ := json.Marshal(validation.AddProductRequest{PvzID: "pvz1", Type: models.Electronic})
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreateProduct(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		prod := models.Product{ID: "p1", ReceptionID: "r1", Type: models.Clothes}
		svc.EXPECT().
			CreateProduct("pvz1", models.Clothes).
			Return(prod, nil).
			Once()

		reqBody, _ := json.Marshal(validation.AddProductRequest{PvzID: "pvz1", Type: models.Clothes})
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreateProduct(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)

		var got models.Product
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, prod, got)
	})
}

func TestCloseLastReception(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pvz/bad/close_last_reception", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues("bad")

		err := h.CloseLastReception(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		valid := "123e4567-e89b-12d3-a456-426655440000"
		svc.EXPECT().
			CloseLastReception(valid).
			Return(models.Reception{}, errors.New("err")).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/pvz/"+valid+"/close_last_reception", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues(valid)

		err := h.CloseLastReception(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		valid := "123e4567-e89b-12d3-a456-426655440000"
		rc := models.Reception{ID: "r2", PvzID: valid, Status: models.StatusClose}
		svc.EXPECT().
			CloseLastReception(valid).
			Return(rc, nil).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/pvz/"+valid+"/close_last_reception", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues(valid)

		err := h.CloseLastReception(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var got models.Reception
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, rc, got)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pvz/bad/delete_last_product", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues("bad")

		err := h.DeleteLastProduct(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		valid := "123e4567-e89b-12d3-a456-426655440000"
		svc.EXPECT().
			DeleteLastProduct(valid).
			Return(errors.New("err")).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/pvz/"+valid+"/delete_last_product", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues(valid)

		err := h.DeleteLastProduct(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		valid := "123e4567-e89b-12d3-a456-426655440000"
		svc.EXPECT().
			DeleteLastProduct(valid).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/pvz/"+valid+"/delete_last_product", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pvzId")
		c.SetParamValues(valid)

		err := h.DeleteLastProduct(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Empty(t, rec.Body.String())
	})
}

func TestGetPVZ(t *testing.T) {
	e, svc := setup()
	h := NewHandler(svc)

	t.Run("service error", func(t *testing.T) {
		svc.EXPECT().
			GetPVZInfo("2025-04-01", "2025-04-20", 1, 10).
			Return(nil, errors.New("oops")).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/pvz?startDate=2025-04-01&endDate=2025-04-20&page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetPVZ(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success empty", func(t *testing.T) {
		svc.EXPECT().
			GetPVZInfo("2025-04-01", "2025-04-20", 1, 10).
			Return([]models.PVZInfo{}, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/pvz?startDate=2025-04-01&endDate=2025-04-20&page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetPVZ(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, "[]", rec.Body.String())
	})
}
