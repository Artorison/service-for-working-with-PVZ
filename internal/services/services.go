package services

import (
	"pvz/internal/models"
	"pvz/pkg/utils"
	"time"
)

//go:generate mockery --name=PvzUserStore --dir=. --output=./mocks --outpkg=mocks --with-expecter
type PvzUserStore interface {
	PvzStore
	UserStore
}

type PvzStore interface {
	CreatePVZ(city models.City) (models.PVZ, error)
	CreateReception(pvzID string) (models.Reception, error)
	CreateProduct(pvzID string, prType models.ProductType) (models.Product, error)
	CloseLastReception(pvzID string) (models.Reception, error)
	DeleteLastProduct(pvzID string) error
	GetPVZInfo(start, end time.Time, page, limit int) ([]models.PVZInfo, error)
}
type UserStore interface {
	RegisterUser(email, password string, role models.Role) (models.User, error)
	LoginUser(email, password string) (models.Token, error)
}

type Service struct {
	Repo PvzUserStore
}

func NewService(repo PvzUserStore) *Service {
	return &Service{Repo: repo}
}

func (s *Service) RegisterUser(email, password string, role models.Role) (models.User, error) {

	hash, err := utils.GenerateHashPassword(password)
	if err != nil {
		return models.User{}, models.Wrap("generate hash", err)
	}

	user, err := s.Repo.RegisterUser(email, hash, role)
	if err != nil {
		return models.User{}, models.Wrap("failed to register user", err)

	}
	return user, nil
}

func (s *Service) LoginUser(email, password string) (models.Token, error) {
	return s.Repo.LoginUser(email, password)
}

func (s *Service) CreatePVZ(city models.City) (models.PVZ, error) {
	return s.Repo.CreatePVZ(city)
}

func (s *Service) CreateReception(pvzID string) (models.Reception, error) {
	return s.Repo.CreateReception(pvzID)
}
func (s *Service) CreateProduct(pvzID string, prType models.ProductType) (models.Product, error) {
	return s.Repo.CreateProduct(pvzID, prType)
}

func (s *Service) CloseLastReception(pvzID string) (models.Reception, error) {
	return s.Repo.CloseLastReception(pvzID)
}

func (s *Service) DeleteLastProduct(pvzID string) error {
	return s.Repo.DeleteLastProduct(pvzID)
}

func (s *Service) GetPVZInfo(start, end string, page, limit int) ([]models.PVZInfo, error) {

	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}
	startDate := time.Time{}
	endDate := time.Now().UTC()
	var err error

	if start != "" {
		startDate, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, models.Wrap("invalid startDate", err)
		}
	}
	if end != "" {
		endDate, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return nil, models.Wrap("invalid endDate", err)
		}
	}

	return s.Repo.GetPVZInfo(startDate, endDate, page, limit)
}
