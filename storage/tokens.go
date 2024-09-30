package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	Token     string
	UserID    uint
	Name      string
	ExpiresAt time.Time
}

func (storage *Storage) NewToken(ID uint, name string, ExpiresAt time.Time) (string, error) {
	tokenString := uuid.NewString()
	token := Token{
		Token:     tokenString,
		UserID:    ID,
		Name:      name,
		ExpiresAt: ExpiresAt,
	}

	res := storage.db.Create(&token)
	return tokenString, res.Error
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

func (storage *Storage) FindTokenByToken(tokenString string) (*Token, error) {
	token := Token{}
	res := storage.db.Where("token = ?", tokenString).First(&token)
	switch res.Error {
	case nil:
		return &token, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrDataNotFound
	default:
		return nil, res.Error
	}
}

func (storage *Storage) FindTokenByName(accId uint, name string) (*Token, error) {
	token := Token{}
	res := storage.db.Where("name = ?", name).Where("user_id = ?", accId).First(&token)
	switch res.Error {
	case nil:
		return &token, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrDataNotFound
	default:
		return nil, res.Error
	}
}

func (storage *Storage) ExtendToken(token *Token) error {
	res := storage.db.Model(&Token{}).
		Where("id = ?", token.ID).
		Update("expires_at", token.ExpiresAt)
	return res.Error
}

func (storage *Storage) InvalidateTokenByName(name string, ownerId uint) {
	storage.db.Model(&Token{}).Where(&Token{Name: name}).Where(&Token{UserID: ownerId}).Delete(&Token{})
}
