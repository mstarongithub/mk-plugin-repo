package auth

import (
	"errors"
	"image"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
	"github.com/mstarongithub/mk-plugin-repo/util"
)

type RegisterProcess struct {
	ProcessId            string
	StartingTimestamp    time.Time
	LastUpdatedTimestamp time.Time

	Username      string
	Mail          *string
	Description   *string
	AuthMethods   customtypes.AuthMethods
	PasswordHash  []byte
	FidoToken     *string
	FidoConfirmed bool
	TotpUrl       *string
	TotpConfirmed bool
	Passkeys      map[string]string
}

type RegisterRequestHolder struct {
	Requests map[string]RegisterProcess
	sync.RWMutex
}

var (
	ErrRegisterProcessCancelled = errors.New("registration process cancelled")
	ErrRegisterNoSuchProcess    = errors.New("no registration process with this id")
)

// Start a new registration process
// Returns the process ID
func (a *Auth) RegisterStart(username string) string {
	// Only allow new usernames
	if _, err := a.store.FindAccountByName(username); !errors.Is(err, storage.ErrAccountNotFound) {
		return ""
	}
	processId := uuid.NewString()
	startingTime := time.Now()
	tmpProcess := RegisterProcess{
		ProcessId:            processId,
		Username:             username,
		StartingTimestamp:    startingTime,
		LastUpdatedTimestamp: startingTime,
	}
	a.registerRequests.Lock()
	a.registerRequests.Requests[processId] = tmpProcess
	a.registerRequests.Unlock()
	return processId
}

// Set the password on a registration process
// Takes in the process Id and the raw, non-hashed password
// This function takes care of hashing itself
// If any step fails, returns an error
// If nil, the process updated properly
// NOTE: Overwrites any previously set password
func (a *Auth) RegisterContinuePassword(processId string, passwordRaw string) error {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return ErrRegisterNoSuchProcess
	}

	passwordHash, err := a.hasher.Hash([]byte(passwordRaw))
	if err != nil {
		a.log.Warn().Err(err).Msg("Failed to hash password during register progress")
		return ErrRegisterProcessCancelled
	}
	process.PasswordHash = passwordHash
	process.LastUpdatedTimestamp = time.Now()

	a.registerRequests.Lock()
	a.registerRequests.Requests[processId] = process
	a.registerRequests.Unlock()

	return nil
}

// Set the description for a registration process
// This is the description later displayed in the account
func (a *Auth) RegisterContinueDescription(processId, description string) error {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return ErrRegisterNoSuchProcess
	}
	process.Description = &description
	a.registerRequests.Lock()
	a.registerRequests.Requests[processId] = process
	a.registerRequests.Unlock()
	return nil
}

// Set the email for an account in the registration process
// TODO: Add confirmation check
func (a *Auth) RegisterContinueMail(processId, mail string) error {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return ErrRegisterNoSuchProcess
	}
	process.Mail = &mail
	a.registerRequests.Lock()
	a.registerRequests.Requests[processId] = process
	a.registerRequests.Unlock()
	return nil
}

// Generate a totp key for the given account during the registration process
// Token will only be used after getting confirmed via RegisterContinueMfaTotpConfirm
func (a *Auth) RegisterContinueMfaTotpGenerate(processId string) (error, string, *image.Image) {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return ErrRegisterNoSuchProcess, "", nil
	}

	totpConfig := totp.GenerateOpts{
		Issuer: "Mk Plugin Repo on " + util.TakeApartRootUrlString(
			"http://localhost:8080",
		).Domain,
		AccountName: process.Username,
	}
	key, err := totp.Generate(totpConfig)
	if err != nil {
		return err, "", nil
	}
	tmpUrl := key.URL()
	process.TotpUrl = &tmpUrl
	img, err := key.Image(200, 200)
	if err != nil {
		return err, "", nil
	}
	a.registerRequests.Lock()
	a.registerRequests.Requests[processId] = process
	a.registerRequests.Unlock()

	return nil, key.URL(), &img
}

func (a *Auth) RegisterContinueMfaTotpConfirm(
	processId string,
	passcode string,
) (bool, error) {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return false, ErrRegisterNoSuchProcess
	}
	key, err := otp.NewKeyFromURL(*process.TotpUrl)
	if err != nil {
		return false, err
	}
	if totp.Validate(passcode, key.Secret()) {
		process.TotpConfirmed = true
		a.registerRequests.Lock()
		a.registerRequests.Requests[processId] = process
		a.registerRequests.Unlock()
		return true, nil
	} else {
		return false, nil
	}
}

func (a *Auth) RegisterContinuePasskeyStart()    {}
func (a *Auth) RegisterContinuePassKeyComplete() {}

func (a *Auth) RegisterFinalise(processId string) error {
	a.registerRequests.RLock()
	process, ok := a.registerRequests.Requests[processId]
	a.registerRequests.RUnlock()
	if !ok {
		return ErrRegisterNoSuchProcess
	}
	newAcc := storage.Account{
		Name:        process.Username,
		Mail:        process.Mail,
		Links:       make(customtypes.GenericSlice[string], 0),
		Description: "",

		CanApprovePlugins: false,
		CanApproveUsers:   false,
		PluginsOwned:      make(customtypes.GenericSlice[uint], 0),
		Approved:          false,

		PasswordHash: process.PasswordHash,
	}
	if process.Description != nil {
		newAcc.Description = *process.Description
	}
	if process.TotpConfirmed {
		newAcc.TotpToken = process.TotpUrl
	}
	if process.FidoConfirmed {
		newAcc.FidoToken = process.FidoToken
	}
	if newAcc.PasswordHash != nil {
		newAcc.AuthMethods |= customtypes.AUTH_METHOD_PASSWORD
		switch {
		case newAcc.TotpToken != nil:
			newAcc.AuthMethods |= customtypes.AUTH_METHOD_TOTP
		case newAcc.FidoToken != nil:
			newAcc.AuthMethods |= customtypes.AUTH_METHOD_FIDO
		}
	}
	_, err := a.store.AddNewAccount(newAcc)
	if err != nil {
		return err
	}
	a.registerRequests.Lock()
	delete(a.registerRequests.Requests, processId)
	a.registerRequests.Unlock()
	return nil
}

func (a *Auth) RegisterCancel(processId string) {
	a.registerRequests.Lock()
	delete(a.registerRequests.Requests, processId)
	a.registerRequests.Unlock()
}
