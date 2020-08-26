package application

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/justinas/nosurf"
	"github.com/volatiletech/authboss/v3/lock"
	"github.com/volatiletech/authboss/v3/remember"

	rbac "github.com/euroteltr/rbac"
	chi "github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	ttool "github.com/tarantool/go-tarantool"
	authboss "github.com/volatiletech/authboss/v3"
	confirm "github.com/volatiletech/authboss/v3/confirm"

	store "github.com/myjupyter/echelon/store"
)

var (
	permissionID = "users"
	description  = "Users Resource"
	actions      = []rbac.Action{
		rbac.Action("foo"),
		rbac.Action("bar"),
		rbac.Action("sigma"),
	}
)

type Application struct {
	config *Config
	logger *log.Logger
	mux    *chi.Mux

	roleController *rbac.RBAC
	permission     *rbac.Permission
	database       *store.TarantoolStorer
	ae             *AuthbossEntities
}

func NewApplication(configs string, roleConfig string) *Application {
	logger := log.New()

	config := NewConfig(configs)
	if err := config.loadConfig(); err != nil {
		logger.Fatal(err)
	}

	level, err := log.ParseLevel(config.logLevel)
	if err != nil {
		logger.Fatal(err)
	}

	logger.SetLevel(level)

	roleController := rbac.New(logger)
	permission, err := roleController.RegisterPermission(permissionID, description, actions...)
	if err != nil {
		logger.Fatal(err)
	}

	file, err := os.Open(roleConfig)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()
	if err = roleController.LoadJSON(file); err != nil {
		logger.Fatal(err)
	}

	return &Application{
		config:         config,
		logger:         logger,
		mux:            chi.NewRouter(),
		roleController: roleController,
		permission:     permission,
	}
}

func (app *Application) Start() error {
	app.database = store.NewTarantoolStorer(app.config.dataBaseAddr, ttool.Opts{
		User: app.config.dataBaseUser,
		Pass: app.config.dataBasePass,
	})

	app.ae = NewAuthbossEntities()
	ablogger := NewLogger(app.logger)
	if err := app.ae.configAuthboss(app.config, app.database, ablogger); err != nil {
		return err
	}

	mux := app.mux
	ab := app.ae.Authboss()

	mux.Use(logger, nosurfing, ab.LoadClientStateMiddleware, remember.Middleware(ab), app.dataInjector)
	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.Middleware2(ab, authboss.RequireFullAuth, authboss.RespondUnauthorized), lock.Middleware(ab), confirm.Middleware(ab))
		mux.MethodFunc("GET", "/foo", app.foo)
		mux.MethodFunc("GET", "/bar", app.bar)
		mux.MethodFunc("GET", "/sigma", app.sigma)
	})
	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(ab))
		mux.Mount("/auth", http.StripPrefix("/auth", ab.Config.Core.Router))
	})
	mux.Get("/", app.index)

	app.logger.Info("Server has been started at " + app.config.appAddr)
	return http.ListenAndServe(app.config.appAddr, app.mux)
}

// Rendering
func (app *Application) layoutData(w http.ResponseWriter, r **http.Request) authboss.HTMLData {
	ab := app.ae.Authboss()

	var currentUserName string
	var permission bool

	userInfer, err := ab.LoadCurrentUser(r)
	if userInfer != nil && err == nil {
		currentUserName = userInfer.(*store.User).Name
		permission = userInfer.(*store.User).RoleID == "admin"
	}

	return authboss.HTMLData{
		"loggedin":          userInfer != nil,
		"current_user_name": currentUserName,
		"permission":        permission,
		"csrf_token":        nosurf.Token(*r),
		"flash_success":     authboss.FlashSuccess(w, *r),
		"flash_error":       authboss.FlashError(w, *r),
	}
}

func (app *Application) dataInjector(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := app.layoutData(w, &r)
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

func (app *Application) mustRender(w http.ResponseWriter, r *http.Request, name string, data authboss.HTMLData) {
	var current authboss.HTMLData
	dataIntf := r.Context().Value(authboss.CTXKeyData)
	if dataIntf == nil {
		current = authboss.HTMLData{}
	} else {
		current = dataIntf.(authboss.HTMLData)
	}

	current.MergeKV("csrf_token", nosurf.Token(r))
	current.Merge(data)

	err := app.ae.templates.Render(w, name, current)
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintln(w, "Error occurred rendering template:", err)
}

func (app *Application) mustRenderWithPermission(w http.ResponseWriter, r *http.Request, name string, data authboss.HTMLData) {
	ab := app.ae.Authboss()

	var current authboss.HTMLData
	dataIntf := r.Context().Value(authboss.CTXKeyData)
	if dataIntf == nil {
		current = authboss.HTMLData{}
	} else {
		current = dataIntf.(authboss.HTMLData)
	}

	current.MergeKV("csrf_token", nosurf.Token(r))
	current.Merge(data)

	var role string
	userInfer, err := ab.LoadCurrentUser(&r)
	if userInfer != nil && err == nil {
		role = userInfer.(*store.User).RoleID
	}

	if app.roleController.IsGranted(role, app.permission, rbac.Action(name)) {
		err := app.ae.templates.Render(w, name, current)
		if err == nil {
			return
		}
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintln(w, "Error occurred rendering template:", err)
}

func (app *Application) index(w http.ResponseWriter, r *http.Request) {
	app.mustRender(w, r, "index", authboss.HTMLData{})
}

func (app *Application) foo(w http.ResponseWriter, r *http.Request) {
	app.mustRenderWithPermission(w, r, "foo", authboss.HTMLData{})
}

func (app *Application) bar(w http.ResponseWriter, r *http.Request) {
	app.mustRenderWithPermission(w, r, "bar", authboss.HTMLData{})
}

func (app *Application) sigma(w http.ResponseWriter, r *http.Request) {
	app.mustRenderWithPermission(w, r, "sigma", authboss.HTMLData{})
}
