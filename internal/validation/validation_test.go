package validation

import (
	"testing"
	"time"

	"pvz/internal/models"

	"github.com/stretchr/testify/require"
)

func TestValidateRole(t *testing.T) {
	v := NewValidator()
	cases := []struct {
		name    string
		role    models.Role
		wantErr bool
	}{
		{"employee", models.Employee, false},
		{"moderator", models.Moderator, false},
		{"invalid", models.Role("moder"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := struct {
				Role models.Role `valid:"role"`
			}{tc.role}
			err := v.Validate(&obj)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestValidateCity(t *testing.T) {
	v := NewValidator()
	cases := []struct {
		name    string
		city    models.City
		wantErr bool
	}{
		{"Moscow", models.Moscow, false},
		{"SaintP", models.SaintP, false},
		{"Kazan", models.Kazan, false},
		{"invalid", models.City("Новосибирск"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := struct {
				City models.City `valid:"city"`
			}{tc.city}
			err := v.Validate(&obj)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestValidateProductType(t *testing.T) {
	v := NewValidator()
	cases := []struct {
		name    string
		pt      models.ProductType
		wantErr bool
	}{
		{"elec", models.Electronic, false},
		{"cloth", models.Clothes, false},
		{"shoes", models.Shoes, false},
		{"invalid", models.ProductType("food"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := struct {
				Type models.ProductType `valid:"productType"`
			}{tc.pt}
			err := v.Validate(&obj)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestValidateDateTime(t *testing.T) {
	v := NewValidator()
	valid := time.Now().UTC().Format(time.RFC3339)
	cases := []struct {
		name    string
		dt      string
		wantErr bool
	}{
		{"good", valid, false},
		{"bad", "20-04-2025 15:00", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := struct {
				DateTime string `valid:"datetime"`
			}{tc.dt}
			err := v.Validate(&obj)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}
