package models

import (
	"time"
)

type (
	Token           string
	Role            string
	City            string
	ReceptionStatus string
	ProductType     string
)

const (
	Employee  Role = "employee"
	Moderator Role = "moderator"

	Moscow City = "Москва"
	SaintP City = "Санкт-Петербург"
	Kazan  City = "Казань"

	StatusInProgress ReceptionStatus = "in_progress"
	StatusClose      ReceptionStatus = "close"

	Electronic ProductType = "электроника"
	Clothes    ProductType = "одежда"
	Shoes      ProductType = "обувь"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             City      `json:"city"`
}

type Reception struct {
	ID       string          `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PvzID    string          `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
}

type Product struct {
	ID          string      `json:"id"`
	DateTime    time.Time   `json:"dateTime"`
	Type        ProductType `json:"type"`
	ReceptionID string      `json:"receptionId"`
}

type PVZInfo struct {
	Pvz        PVZ                     `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}
