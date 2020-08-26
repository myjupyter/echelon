package application

import (
	"encoding/base64"
	"html/template"
	"regexp"
	"time"

	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	"github.com/volatiletech/authboss/v3/defaults"
	_ "github.com/volatiletech/authboss/v3/logout"
	"github.com/volatiletech/authboss/v3/otp/twofactor"
	"github.com/volatiletech/authboss/v3/otp/twofactor/totp2fa"
	_ "github.com/volatiletech/authboss/v3/recover"
	_ "github.com/volatiletech/authboss/v3/register"

	"github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss-renderer"

	"github.com/aarondl/tpl"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
)

var funcs = template.FuncMap{
	"formatDate": func(date time.Time) string {
		return date.Format("2006/01/02 03:04pm")
	},
	"yield": func() string { return "" },
}

type AuthbossEntities struct {
	ab            *authboss.Authboss
	schemaDecoder *schema.Decoder
	templates     tpl.Templates
}

func NewAuthbossEntities() *AuthbossEntities {
	ab := authboss.New()
	schema := schema.NewDecoder()
	templates := tpl.Must(tpl.Load("views", "views/partials", "layout.html.tpl", funcs))

	return &AuthbossEntities{
		ab:            ab,
		schemaDecoder: schema,
		templates:     templates,
	}
}

func (ae *AuthbossEntities) Authboss() *authboss.Authboss {
	return ae.ab
}

func (ae *AuthbossEntities) configAuthboss(config *Config, storage authboss.ServerStorer, logger authboss.Logger) error {
	ab := ae.Authboss()

	ab.Config.Paths.RootURL = "http://" + config.appAddr
	ab.Config.Modules.LogoutMethod = "GET"

	ab.Config.Storage.Server = storage
	ab.Config.Core.Logger = logger

	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)
	sessionStore := abclientstate.NewSessionStorer("ab_blog", sessionStoreKey, nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false
	cstore.MaxAge(int((30 * 24 * time.Hour) / time.Second))
	ab.Config.Storage.SessionState = sessionStore

	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false
	ab.Config.Storage.CookieState = cookieStore

	ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/auth", "ab_views")
	ab.Config.Core.MailRenderer = abrenderer.NewEmail("/auth", "ab_views")

	ab.Config.Modules.RegisterPreserveFields = []string{"email", "name"}
	ab.Config.Modules.TOTP2FAIssuer = "ABBlog"
	ab.Config.Modules.ResponseOnUnauthed = authboss.RespondRedirect

	ab.Config.Modules.TwoFactorEmailAuthRequired = true

	defaults.SetCore(&ab.Config, false, false)

	emailRule := defaults.Rules{
		FieldName: "email", Required: true,
		MatchError: "Must be a valid e-mail address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]+`),
	}

	passwordRule := defaults.Rules{
		FieldName: "password", Required: true,
		MinLength: 4,
	}

	nameRule := defaults.Rules{
		FieldName: "name", Required: true,
		MinLength: 2,
		MustMatch: regexp.MustCompile(`\w+`),
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: false,
		Rulesets: map[string][]defaults.Rules{
			"register":    {emailRule, passwordRule, nameRule},
			"recover_end": {passwordRule},
		},
		Confirms: map[string][]string{
			"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": {"email", "name", "password"},
		},
	}

	twofaRecovery := &twofactor.Recovery{Authboss: ab}
	if err := twofaRecovery.Setup(); err != nil {
		return err
	}
	totp := &totp2fa.TOTP{Authboss: ab}
	if err := totp.Setup(); err != nil {
		return err
	}
	if err := ab.Init(); err != nil {
		return err
	}
	ae.schemaDecoder.IgnoreUnknownKeys(true)

	return nil
}
