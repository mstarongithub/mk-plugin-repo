package storage

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PasskeySession struct {
	ID   string               `gorm:"primarykey"`
	Data webauthn.SessionData `gorm:"serializer:json"`
}

// ---- Section SessionStore

func (s *Storage) GenSessionID() (string, error) {
	x := uuid.NewString()
	logrus.WithField("id", x).Debugln("Generating passkey session id")
	return x, nil
}

func (s *Storage) GetSession(token string) (*webauthn.SessionData, bool) {
	logrus.WithField("token", token).Debug("Grabbing passkey session")
	session := PasskeySession{}
	res := s.db.Where("id = ?", token).First(&session)
	if res.Error != nil {
		return nil, false
	}
	logrus.WithField("session", &session).Debug("Found session")
	return &session.Data, true
}

func (s *Storage) SaveSession(token string, data *webauthn.SessionData) {
	logrus.WithField("token", token).WithField("data", data).Debug("Saving passkey session")
	session := PasskeySession{
		ID:   token,
		Data: *data,
	}
	s.db.Save(&session)
}

func (s *Storage) DeleteSession(token string) {
	s.db.Delete(&PasskeySession{ID: token})
}
