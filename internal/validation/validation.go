package validation

import (
	"pvz/internal/models"
	"slices"

	"github.com/asaskevich/govalidator"
)

type Govalidator struct{}

func NewValidator() *Govalidator {
	initValidation()
	return &Govalidator{}
}

func check[T comparable](v T, set ...T) bool {
	return slices.Contains(set, v)
}

func (v *Govalidator) Validate(i any) error {
	if ok, err := govalidator.ValidateStruct(i); !ok {
		return err
	}
	return nil
}

func initValidation() {
	govalidator.TagMap["role"] = func(str string) bool {
		return check(models.Role(str), models.Employee, models.Moderator)
	}
	govalidator.TagMap["city"] = func(str string) bool {
		return check(models.City(str), models.Moscow, models.SaintP, models.Kazan)
	}
	govalidator.TagMap["productType"] = func(str string) bool {
		return check(models.ProductType(str), models.Electronic, models.Clothes, models.Shoes)
	}
	govalidator.TagMap["datetime"] = govalidator.IsRFC3339
}
