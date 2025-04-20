package handlers

import (
	"net/http"
	"pvz/internal/models"
	"pvz/internal/validation"
	"pvz/pkg/utils"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Service PvzUserService
}

//go:generate mockery --name=PvzUserService --dir=. --output=./mocks --outpkg=mocks --with-expecter
type PvzUserService interface {
	PvzService
	UserService
}

func NewHandler(service PvzUserService) *Handler {
	return &Handler{Service: service}
}

type PvzService interface {
	CreatePVZ(city models.City) (models.PVZ, error)
	GetPVZInfo(start, end string, page, limit int) ([]models.PVZInfo, error)

	CreateReception(pvzID string) (models.Reception, error)
	CloseLastReception(pvzID string) (models.Reception, error)

	CreateProduct(pvzID string, prType models.ProductType) (models.Product, error)
	DeleteLastProduct(pvzID string) error
}
type UserService interface {
	RegisterUser(email, password string, role models.Role) (models.User, error)
	LoginUser(email, password string) (models.Token, error)
}

func (h *Handler) DummyLogin(c echo.Context) error {
	var req validation.RoleForDummyLogin

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	token, err := utils.GetJWToken(req.Role, "")
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("Invalid request"))
	}
	return c.JSON(http.StatusOK, token)
}

func (h *Handler) RegisterUser(c echo.Context) error {

	var req validation.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	user, err := h.Service.RegisterUser(req.Email, req.Password, req.Role)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) LoginUser(c echo.Context) error {

	var req validation.LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnauthorized, models.Err("invalid JSON: "+err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusUnauthorized, models.Err(err.Error()))
	}

	token, err := h.Service.LoginUser(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.Err(err.Error()))
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handler) CreatePVZ(c echo.Context) error {

	var pvz validation.CreatePVZRequest

	if err := c.Bind(&pvz); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}
	if err := c.Validate(pvz); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	reqPVZ, err := h.Service.CreatePVZ(pvz.City)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	return c.JSON(http.StatusCreated, reqPVZ)
}

func (h *Handler) CreateReception(c echo.Context) error {

	var req validation.CreateReceptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}
	res, err := h.Service.CreateReception(req.PvzID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) CreateProduct(c echo.Context) error {
	var req validation.AddProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	res, err := h.Service.CreateProduct(req.PvzID, req.Type)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) CloseLastReception(c echo.Context) error {
	pvzID := c.Param("pvzId")
	if !govalidator.IsUUID(pvzID) {
		return c.JSON(http.StatusBadRequest, models.Err("invalid pvzId, uuid expected"))
	}
	res, err := h.Service.CloseLastReception(pvzID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) DeleteLastProduct(c echo.Context) error {
	pvzID := c.Param("pvzId")
	if !govalidator.IsUUID(pvzID) {
		return c.JSON(http.StatusBadRequest, models.Err("invalid pvzId, uuid expected"))
	}

	err := h.Service.DeleteLastProduct(pvzID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetPVZ(c echo.Context) error {

	var req validation.GetPVZQuery
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err("invalid JSON: "+err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.Err(err.Error()))
	}

	res, err := h.Service.GetPVZInfo(req.StartDate, req.EndDate, req.Page, req.Limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Err(err.Error()))
	}

	return c.JSON(http.StatusOK, res)
}
