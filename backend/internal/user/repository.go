package user

import "gorm.io/gorm"

type Repository interface {
	GetUserByToken(token string) (*Users, error)
}

type reporsitory struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) Repository {
	return &reporsitory{db: db}
}

// GetUserByToken implements [Reporsitory].
func (u *reporsitory) GetUserByToken(token string) (*Users, error) {
	user := Users{}
	if err := u.db.Where("token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
