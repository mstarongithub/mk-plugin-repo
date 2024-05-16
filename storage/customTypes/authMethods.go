package customtypes

type AuthMethods uint

const (
	// Account has no authentication method set (Login requires only username, should only have read perm)
	AUTH_METHOD_NONE = AuthMethods(0)
	// Account requires a password to login
	AUTH_METHOD_PASSWORD = AuthMethods(1 << iota)
	// Account requires a passkey to login. If both this and password are set, both are valid options
	AUTH_METHOD_PASSKEY
	// Account requires a fido key in addition to a password
	AUTH_METHOD_FIDO
	// Account requires a totp token in addition to a password
	AUTH_METHOD_TOTP
	// Account requires a mail token in addition to a password (should not be used unless no other mfa method is set)
	AUTH_METHOD_MAIL
)

func AuthIsFlagSet(val, flag AuthMethods) bool {
	return val&flag != 0
}
