package user

import "gorm.io/gorm"

type Repository interface {
	GetUserByToken(token string) (*Users, error)
	GetUserByID(id uint) (*Users, error)
}

type reporsitory struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) Repository {
	return &reporsitory{db: db}
}

func (u *reporsitory) GetUserByToken(token string) (*Users, error) {
	var user Users
	if err := u.db.Model(&Users{}).Where("token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *reporsitory) GetUserByID(id uint) (*Users, error) {
	var usr Users
	if err := u.db.First(&usr, id).Error; err != nil {
		return nil, err
	}
	return &usr, nil
}
