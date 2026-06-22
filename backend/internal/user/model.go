package user

import (
	"database/sql/driver"
	"errors"

	"gorm.io/gorm"
)

type UserRole string

const (
	Admin UserRole = "ADMIN"
	User  UserRole = "USER"
)

func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan UserRole")
	}
	*r = UserRole(string(bytes))
	return nil
}

type Users struct {
	Name               string   `gorm:"type:varchar(100);not null" json:"name"`
	Role               UserRole `gorm:"type:user_role_enum;default:'USER';not null" json:"role"`
	Token              string   `gorm:"type:varchar(100);unique" json:"token"`
	DaylyCalorieLimit  int      `gorm:"type:int;not null" json:"dayly_calorie_limit"`
	MounthlyPriceLimit int      `gorm:"type:int;not null" json:"mounthly_price_limit"`
	gorm.Model
}
