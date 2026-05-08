package routers

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"myblog/models"
	"myblog/modules/auth"
	"myblog/services"
	"myblog/web"
)

type Router struct {
	app       *services.App
	templates *template.Template
}

type viewData struct {
	AppName     string
	Title       string
	User        *models.User
	Session     *auth.Session
	CSRF        string
	Theme       string
	Error       string
	Message     string
	Query       string
	Articles    []models.Article
	Article     *models.Article
	ArticleHTML template.HTML
	ProfileUser *models.User
	Teams       []models.Team
	Team        *models.Team
	Spaces      []models.Space
	Space       *models.Space
}

func New(app *services.App) (http.Handler, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04")
		},
	}).ParseFS(web.FS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	rt := &Router{app: app, templates: tmpl}
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(rt.installGate)

	staticFS, err := fs.Sub(web.FS, "public")
	if err != nil {
		return nil, err
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.FS(staticFS))))

	r.Get("/install", rt.installPage)
	r.Post("/install", rt.installPost)

	r.Get("/", rt.home)
	r.Get("/user/login", rt.loginPage)
	r.Post("/user/login", rt.loginPost)
	r.Post("/user/logout", rt.logoutPost)
	r.Get("/user/sign_up", rt.signUpPage)
	r.Post("/user/sign_up", rt.signUpPost)
	r.Get("/user/settings", rt.settingsPage)
	r.Post("/user/settings/profile", rt.settingsProfilePost)
	r.Post("/user/settings/password", rt.settingsPasswordPost)
	r.Post("/user/settings/theme", rt.settingsThemePost)
	r.Get("/user/articles/new", rt.personalArticleNewPage)
	r.Post("/user/articles/new", rt.personalArticleCreatePost)
	r.Get("/users/{name}", rt.userProfilePage)
	r.Get("/teams", rt.teamsPage)
	r.Get("/teams/{team}", rt.teamPage)
	r.Get("/teams/{team}/{space}", rt.spacePage)
	r.Get("/articles/{idOrSlug}", rt.articlePage)
	r.Get("/search", rt.searchPage)

	r.Route("/api/v1", rt.apiRoutes)
	return r, nil
}

func (rt *Router) installGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rt.app.IsInstalled() || strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/install" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "application is not installed"})
			return
		}
		http.Redirect(w, r, "/install", http.StatusFound)
	})
}

func (rt *Router) render(w http.ResponseWriter, r *http.Request, name string, data viewData) {
	user, session, ok := rt.app.CurrentUser(r)
	if ok {
		data.User = user
		data.Session = session
		data.CSRF = session.CSRFToken
		data.Theme = services.NormalizeTheme(user.Theme)
	}
	if data.Theme == "" {
		data.Theme = "light"
	}
	data.AppName = rt.app.Config.AppName
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := rt.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (rt *Router) requireUser(w http.ResponseWriter, r *http.Request) (*models.User, *auth.Session, bool) {
	user, session, ok := rt.app.CurrentUser(r)
	if ok {
		return user, session, true
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "authentication required"})
	} else {
		http.Redirect(w, r, "/user/login", http.StatusFound)
	}
	return nil, nil, false
}

func (rt *Router) requireCSRF(w http.ResponseWriter, r *http.Request, session *auth.Session) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	if auth.ValidateCSRF(r, session) {
		return true
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "invalid csrf token"})
	} else {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseInt64(value string) int64 {
	n, _ := strconv.ParseInt(value, 10, 64)
	return n
}
