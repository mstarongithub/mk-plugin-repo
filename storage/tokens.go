package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	Token     string
	UserID    uint
	ExpiresAt *time.Time
}

func (storage *Storage) InsertNewToken(ID uint, token string, ExpiresAt *time.Time) error {

	res := storage.db.Create(&token)
	return res.Error
}

func (storage *Storage) GetTokensForUsername(username string) ([]Token, error) {
	acc, err := storage.FindAccountByName(username)
	if err != nil {
		return nil, err
	}
	tokens := []Token{}
	res := storage.db.Where("user_id = ?", acc.ID).Find(&tokens)
	if res.Error != nil {
		return nil, res.Error
	}
	return tokens, nil

}

func (storage *Storage) GetTokensForAccountID(id uint) ([]Token, error) {
	acc, err := storage.FindAccountByID(id)
	if err != nil {
		return nil, err
	}
	tokens := []Token{}
	res := storage.db.Where("user_id = ?", acc.ID).Find(&tokens)
	if res.Error != nil {
		return nil, res.Error
	}
	return tokens, nil
}

func (storage *Storage) FindToken(tokenString string) (*Token, error) {
	token := Token{}
	res := storage.db.Where("token = ?", tokenString).First(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("couldn'T find token: %w", res.Error)
	}
	return &token, nil
}
