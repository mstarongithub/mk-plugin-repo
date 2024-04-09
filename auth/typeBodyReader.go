package auth

import (
	"net/http"

	"github.com/volatiletech/authboss/v3"
)

type AuthBodyReader struct{}

// Interface authboss.BodyReader
func (abr *AuthBodyReader) Read(page string, r *http.Request) (authboss.Validator, error) {
	return nil, nil
}

// Interface authboss.Validator
func (abr *AuthBodyReader) Validate() []error {
	return nil
}

// Interface authboss.UserValuer
func (abr *AuthBodyReader) GetPID() string {
	return "Placeholder PID"
}

// Interface authboss.UserValuer
// Interface authboss.RecoverEndValuer
func (abr *AuthBodyReader) GetPassword() string {
	return "Placeholder password"
}

// Interface authboss.twofactor.EmailVerifyTokenValuer
// Interface authboss.ConfirmValuer
// Interface authboss.RecoverMiddleValuer
// Interface authboss.RecoverEndValuer
func (abr *AuthBodyReader) GetToken() string {
	return "Placeholder token"
}

// Interface authboss.sms2fa.SMSValuer
func (abr *AuthBodyReader) GetCode() string {
	return "Placeholder Code"
}

// Interface authboss.sms2fa.SMSValuer
func (abr *AuthBodyReader) GetRecoveryCode() string {
	return "Placeholder RecoveryCode"
}

// Interface authboss.RememberValuer
func (abr *AuthBodyReader) GetShouldRemember() bool {
	return false
}

// Interface authboss.ArbitraryValuer
func (abr *AuthBodyReader) GetValues() map[string]string {
	return nil
}
