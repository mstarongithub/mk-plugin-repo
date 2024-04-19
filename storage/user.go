package storage

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// A user account. Profile images are matched by ID
type Account struct {
	gorm.Model
	CanApprovePlugins bool   // Can this account approve new plugin requests?
	CanApproveUsers   bool   // Can this account approve account creation requests?
	Name              string // Name of the account, NOT THE ID
	// Custom links the user has added to the account
	// This can be things like Fedi profile, personal webpage, etc
	Links        customtypes.GenericSlice[string]
	Description  string                         // A description of the account, added by the user. Not necessary
	PluginsOwned customtypes.GenericSlice[uint] // IDs of plugins this account owns (has created)
	Approved     bool                           // Is this account approved for performing any actions

	// ---- authboss things ----
	// Auth
	Mail     string // Mail address of the account
	Password string // Hash of the password

	// Confirm
	ConfirmSelector string
	ConfirmVerifier string
	Confirmed       bool

	// Lock
	AttemptCount int
	LastAttempt  time.Time
	Locked       time.Time

	// Recover
	RecoverSelector    string
	RecoverVerifier    string
	RecoverTokenExpiry time.Time

	// OAuth2
	OAuth2UID          string
	OAuth2Provider     string
	OAuth2AccessToken  string
	OAuth2RefreshToken string
	OAuth2Expiry       time.Time

	// 2fa
	TOTPSecretKey      string
	SMSPhoneNumber     string
	SMSSeedPhoneNumber string
	RecoveryCodes      string
}

var ErrAccountNotFound = errors.New("account not found")
var ErrAccountNotApproved = errors.New("account not approved for this action")

func (s *Storage) FindAccountByName(name string) (*Account, error) {
	// TODO: Add logging
	acc := Account{}

	res := s.db.First(&acc, "name = ?", name)
	if res.RowsAffected == 0 {
		// TODO: Add logging
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf("error while searching for account %s: %w", name, res.Error)
	}
	// TODO: Add logging
	return &acc, nil
}

func (s *Storage) FindAccountByID(id uint) (*Account, error) {
	acc := Account{}
	// TODO: Add logging
	res := s.db.First(&acc, id)
	if res.RowsAffected == 0 {
		// TODO: Add logging
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf("problem while finding account id %d: %w", id, res.Error)
	}
	// TODO: Add logging
	return &acc, nil
}

// Section authboss

func (a *Account) PutPID(pid string) {
	nrBig, err := strconv.ParseUint(pid, 10, 0)
	if err != nil {
		nrBig = math.MaxUint
	}
	a.ID = uint(nrBig)
}
func (a *Account) PutPassword(password string)    { a.Password = password }
func (a *Account) PutEmail(email string)          { a.Mail = email }
func (a *Account) PutConfirmed(c bool)            { a.Confirmed = c }
func (a *Account) PutConfirmedSelector(c string)  { a.ConfirmSelector = c }
func (a *Account) PutConfirmVerifier(c string)    { a.ConfirmVerifier = c }
func (a *Account) PutLocked(l time.Time)          { a.Locked = l }
func (a *Account) PutAttemptCount(c int)          { a.AttemptCount = c }
func (a *Account) PutLastAttempt(l time.Time)     { a.LastAttempt = l }
func (a *Account) PutRecoverSelector(r string)    { a.RecoverSelector = r }
func (a *Account) PutRecoverVerifier(t string)    { a.RecoverVerifier = t }
func (a *Account) PutRecoverExpiry(e time.Time)   { a.RecoverTokenExpiry = e }
func (a *Account) PutTOTPSecretKey(k string)      { a.TOTPSecretKey = k }
func (a *Account) PutSMSPhoneNumber(k string)     { a.SMSPhoneNumber = k }
func (a *Account) PutRecoveryCodes(k string)      { a.RecoveryCodes = k }
func (a *Account) PutOAuth2UID(i string)          { a.OAuth2UID = i }
func (a *Account) PutOAuth2Provider(p string)     { a.OAuth2Provider = p }
func (a *Account) PutOAuth2RefreshToken(t string) { a.OAuth2RefreshToken = t }
func (a *Account) PutOAuth2Expiry(e time.Time)    { a.OAuth2Expiry = e }
func (a *Account) PutArbitrary(values map[string]string) {
	if n, ok := values["name"]; ok {
		a.Name = n
	}
}

func (a *Account) GetPID() string                { return fmt.Sprintf("%d", a.ID) }
func (a *Account) GetPassword() string           { return a.Password }
func (a *Account) GetConfirmed() bool            { return a.Confirmed }
func (a *Account) GetConfirmSelector() string    { return a.ConfirmSelector }
func (a *Account) GetConfirmVerifier() string    { return a.ConfirmVerifier }
func (a *Account) GetLocked() time.Time          { return a.Locked }
func (a *Account) GetAttemptCount() int          { return a.AttemptCount }
func (a *Account) GetLastAttempt() time.Time     { return a.LastAttempt }
func (a *Account) GetRecoverSelector() string    { return a.RecoverSelector }
func (a *Account) GetRecoverVerifier() string    { return a.RecoverVerifier }
func (a *Account) GetRecoverExpiry() time.Time   { return a.RecoverTokenExpiry }
func (a *Account) GetTOTPSecretKey() string      { return a.TOTPSecretKey }
func (a *Account) GetSMSPhoneNumber() string     { return a.SMSPhoneNumber }
func (a *Account) GetSMSPhoneNumberSeed() string { return a.SMSSeedPhoneNumber }
func (a *Account) GetRecoveryCodes() string      { return a.RecoveryCodes }
func (a *Account) IsOAuth2User() bool            { return len(a.OAuth2UID) != 2 }
func (a *Account) GetOAuth2UID() string          { return a.OAuth2UID }
func (a *Account) GetOAuth2Provider() string     { return a.OAuth2Provider }
func (a *Account) GetOAuth2AccessToken() string  { return a.OAuth2AccessToken }
func (a *Account) GetOAuth2RefreshToken() string { return a.OAuth2RefreshToken }
func (a *Account) GetOAuth2Expiry() time.Time    { return a.OAuth2Expiry }
func (a *Account) GetArbitrary() map[string]string {
	return map[string]string{
		"name": a.Name,
	}
}
