package validation

import "pvz/internal/models"

type RoleForDummyLogin struct {
	Role models.Role `json:"role" valid:"required,role"`
}

type RegisterRequest struct {
	Email    string      `json:"email" valid:"required,email"`
	Password string      `json:"password" valid:"required"`
	Role     models.Role `json:"role" valid:"required,role"`
}

type LoginRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required"`
}

type CreatePVZRequest struct {
	City models.City `json:"city" valid:"required,city"`
}

type AddProductRequest struct {
	Type  models.ProductType `json:"type" valid:"required,productType"`
	PvzID string             `json:"pvzId" valid:"required,uuid"`
}

type CreateReceptionRequest struct {
	PvzID string `json:"pvzId" valid:"required,uuid"`
}

type GetPVZQuery struct {
	StartDate string `query:"startDate" valid:"optional,datetime"`
	EndDate   string `query:"endDate" valid:"optional,datetime"`

	Page  int `query:"page" valid:"optional,range(1|10000000)"`
	Limit int `query:"limit" valid:"optional,range(1|30)"`
}
