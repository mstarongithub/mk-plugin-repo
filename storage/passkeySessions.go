package storage

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PasskeySession struct {
	ID   string               `gorm:"primarykey"`
	Data webauthn.SessionData `gorm:"serializer:json"`
}

// ---- Section SessionStore

func (s *Storage) GenSessionID() (string, error) {
	x := uuid.NewString()
	log.Debug().Str("session-id", x).Msg("Generated new passkey session id")
	return x, nil
}

func (s *Storage) GetSession(sessionId string) (*webauthn.SessionData, bool) {
	log.Debug().Str("id", sessionId).Msg("Looking for passkey session")
	session := PasskeySession{}
	res := s.db.Where("id = ?", sessionId).First(&session)
	if res.Error != nil {
		return nil, false
	}
	log.Debug().Str("id", sessionId).Any("webauthn-data", &session).Msg("Found passkey session")
	return &session.Data, true
}

func (s *Storage) SaveSession(token string, data *webauthn.SessionData) {
	log.Debug().Str("id", token).Any("webauthn-data", data).Msg("Saving passkey session")
	session := PasskeySession{
		ID:   token,
		Data: *data,
	}
	s.db.Save(&session)
}

func (s *Storage) DeleteSession(token string) {
	log.Debug().Str("id", token).Msg("Deleting passkey session (if one exists)")
	s.db.Delete(&PasskeySession{ID: token})
}
