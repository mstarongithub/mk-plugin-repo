package authold

import (
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/authboss-clientstate"
	_ "github.com/volatiletech/authboss-renderer"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	_ "github.com/volatiletech/authboss/v3/confirm"
	"github.com/volatiletech/authboss/v3/defaults"
	_ "github.com/volatiletech/authboss/v3/lock"
	_ "github.com/volatiletech/authboss/v3/logout"
	aboauth "github.com/volatiletech/authboss/v3/oauth2"
	_ "github.com/volatiletech/authboss/v3/otp/twofactor"
	_ "github.com/volatiletech/authboss/v3/otp/twofactor/sms2fa"
	_ "github.com/volatiletech/authboss/v3/otp/twofactor/totp2fa"
	_ "github.com/volatiletech/authboss/v3/recover"
	_ "github.com/volatiletech/authboss/v3/register"
	_ "github.com/volatiletech/authboss/v3/remember"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
)

const AB_SESSION_COOKIE_NAME = "mk_plugin_repo"

func SetupAuthboss(
	store *storage.Storage,
	cookieStoreKey []byte,
	sessionStoreKey []byte,
	mailRenderer authboss.Renderer,
) (*authboss.Authboss, error) {
	if config.GlobalConfig == nil {
		logrus.Fatalln("No config loaded! Did you call config.ReadConfig sucessfully first?")
		return nil, config.ErrNoConfig
	}
	ab := authboss.New()
	ab.Config.Storage.Server = store

	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	sessionStore := abclientstate.NewSessionStorer(AB_SESSION_COOKIE_NAME, sessionStoreKey, nil)
	ctstore := sessionStore.Store.(*sessions.CookieStore)
	ctstore.MaxAge(int(30*24*time.Hour) / int(time.Second))
	ab.Config.Storage.CookieState = cookieStore
	ab.Config.Storage.SessionState = sessionStore

	// Leave those up to the caller.
	// Yes, I will have to implement them elsewhere myself, but this is not the place to implement the renderers
	// Needs stuff like the builds from the frontend
	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.MailRenderer = mailRenderer

	// What fields to preserve during user registration
	// I don't quite know if that is necessary since Svelte *probably* handles that itself
	ab.Config.Modules.RegisterPreserveFields = []string{"email", "name"}

	// What we appear as when you scan the oauth2 qr code with for example Google Authenticator
	ab.Config.Modules.TOTP2FAIssuer = "MkPluginRepo"
	ab.Config.Modules.ResponseOnUnauthed = authboss.RespondRedirect

	ab.Config.Modules.TwoFactorEmailAuthRequired = true

	defaults.SetCore(&ab.Config, true, false)

	// Custom reader. Responsible for decoding requests
	ab.Config.Core.BodyReader = &AuthBodyReader{}

	// If credentials for Google oauth are available in the config, set it up
	if config.GlobalConfig.OAuthConfig != nil &&
		len(config.GlobalConfig.OAuthConfig.ClientID) > 1 &&
		len(config.GlobalConfig.OAuthConfig.ClientSecret) > 1 {
		logrus.Infoln("Oauth credentials exist, configuring it")
		ab.Config.Modules.OAuth2Providers = map[string]authboss.OAuth2Provider{
			"google": {
				OAuth2Config: &oauth2.Config{
					ClientID:     config.GlobalConfig.OAuthConfig.ClientID,
					ClientSecret: config.GlobalConfig.OAuthConfig.ClientSecret,
					Scopes:       []string{`profile`, `email`},
					Endpoint:     google.Endpoint,
				},
				FindUserDetails: aboauth.GoogleUserDetails,
			},
		}
	}
	if err := ab.Init(); err != nil {
		return nil, fmt.Errorf("failed to init authboss: %w", err)
	}
	return ab, nil
}
