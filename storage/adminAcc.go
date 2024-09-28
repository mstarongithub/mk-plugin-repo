package storage

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"gitlab.com/mstarongitlab/goutils/other"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

func (storage *Storage) InsertDevAccount() {
	acc := &Account{}
	storage.db.FirstOrCreate(acc, 1)
	acc.Name = "developer"
	// acc.PasswordHash = []byte(
	// 	"$2id$CKXfPfWzlGPT1xUOZ5k.4u$1$65536$16$32$uih.e8WZNJ8PWj6Z.axzh0SARgRjXjnP.p5JWs36c6K",
	// )
	acc.Mail = other.IntoPointer("developer@example.com")
	acc.Links = customtypes.GenericSlice[string]{"example.com"}
	acc.Description = "Developer account. Only exists in the build with the flag authDev"
	acc.CanApprovePlugins = true
	acc.CanApproveUsers = true
	acc.PluginsOwned = make(customtypes.GenericSlice[uint], 0)
	acc.Approved = true
	// acc.AuthMethods = customtypes.AUTH_METHOD_PASSWORD
	// acc.FidoToken = nil
	// acc.TotpToken = nil
	acc.Credentials = make([]webauthn.Credential, 0)
	// Note: This is actually fucked and cursed and everything but good
	// Not a good idea. Good thing this function has no path with which it can be called from main
	acc.PasskeyId = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	storage.db.Save(acc)
}

func (storage *Storage) InsertSudoAccount(username string, passwordHash []byte) {

	acc := &Account{}
	storage.db.FirstOrCreate(acc, 1)
	acc.Name = username
	// acc.PasswordHash = passwordHash
	acc.Mail = nil
	acc.Links = nil
	acc.Description = "Superuser account"
	acc.CanApprovePlugins = true
	acc.CanApproveUsers = true
	acc.PluginsOwned = make(customtypes.GenericSlice[uint], 0)
	acc.Approved = true
	// acc.AuthMethods = customtypes.AUTH_METHOD_PASSWORD
	// acc.FidoToken = nil
	// acc.TotpToken = nil
	acc.Credentials = make([]webauthn.Credential, 0)
	// See note for same field as for dev account .-.
	acc.PasskeyId = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	storage.db.Save(acc)
}
