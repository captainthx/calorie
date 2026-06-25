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
	switch v := value.(type) {
	case string:
		*r = UserRole(v)
	case []byte:
		*r = UserRole(string(v))
	default:
		return errors.New("failed to scan UserRole")
	}
	return nil
}

type Users struct {
	Name               string   `gorm:"type:varchar(100);not null" json:"name"`
	Role               UserRole `gorm:"type:user_role_enum;default:'USER';not null" json:"role"`
	Token              string   `gorm:"type:varchar(100);unique" json:"token"`
	DailyCalorieLimit  int      `gorm:"type:int;not null;default:2100" json:"daily_calorie_limit"`
	MonthlyPriceLimit  int      `gorm:"type:int;not null;default:1000" json:"monthly_price_limit"`
	gorm.Model
}
