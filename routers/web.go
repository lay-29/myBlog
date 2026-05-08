package routers

import (
	"context"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"myblog/models"
	"myblog/services"
)

func (rt *Router) installPage(w http.ResponseWriter, r *http.Request) {
	if rt.app.IsInstalled() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	rt.render(w, r, "install.tmpl", viewData{Title: "安装"})
}

func (rt *Router) installPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "install.tmpl", viewData{Title: "安装", Error: err.Error()})
		return
	}
	opts := services.InstallOptions{
		AppName:       r.FormValue("app_name"),
		HTTPAddr:      r.FormValue("http_addr"),
		HTTPPort:      r.FormValue("http_port"),
		RootURL:       r.FormValue("root_url"),
		DBType:        r.FormValue("db_type"),
		DBHost:        r.FormValue("db_host"),
		DBPort:        r.FormValue("db_port"),
		DBName:        r.FormValue("db_name"),
		DBUser:        r.FormValue("db_user"),
		DBPassword:    r.FormValue("db_password"),
		DBPath:        r.FormValue("db_path"),
		DBSSLMode:     r.FormValue("db_ssl_mode"),
		AdminName:     r.FormValue("admin_name"),
		AdminEmail:    r.FormValue("admin_email"),
		AdminPassword: r.FormValue("admin_password"),
	}
	if err := rt.app.Install(r.Context(), opts); err != nil {
		rt.render(w, r, "install.tmpl", viewData{Title: "安装", Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/user/login?installed=1", http.StatusFound)
}

func (rt *Router) home(w http.ResponseWriter, r *http.Request) {
	if err := rt.app.RequireRuntime(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	articles, err := rt.app.ListPublicArticles(r.Context(), 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rt.render(w, r, "home.tmpl", viewData{Title: "首页", Articles: articles})
}

func (rt *Router) loginPage(w http.ResponseWriter, r *http.Request) {
	message := ""
	if r.URL.Query().Get("installed") == "1" {
		message = "安装完成，请使用管理员账号登录。"
	}
	rt.render(w, r, "login.tmpl", viewData{Title: "登录", Message: message})
}

func (rt *Router) loginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "login.tmpl", viewData{Title: "登录", Error: err.Error()})
		return
	}
	user, err := rt.app.Authenticate(r.Context(), r.FormValue("login"), r.FormValue("password"), r.RemoteAddr)
	if err != nil {
		rt.render(w, r, "login.tmpl", viewData{Title: "登录", Error: err.Error()})
		return
	}
	if err := rt.app.SignIn(w, user.ID); err != nil {
		rt.render(w, r, "login.tmpl", viewData{Title: "登录", Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/teams", http.StatusFound)
}

func (rt *Router) logoutPost(w http.ResponseWriter, r *http.Request) {
	_, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	rt.app.Sessions.Destroy(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (rt *Router) signUpPage(w http.ResponseWriter, r *http.Request) {
	rt.render(w, r, "signup.tmpl", viewData{Title: "注册"})
}

func (rt *Router) signUpPost(w http.ResponseWriter, r *http.Request) {
	if rt.app.Config.Service.DisableRegistration {
		rt.render(w, r, "signup.tmpl", viewData{Title: "注册", Error: "站点已关闭公开注册"})
		return
	}
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "signup.tmpl", viewData{Title: "注册", Error: err.Error()})
		return
	}
	user, err := rt.app.CreateUser(r.Context(), services.CreateUserOptions{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		rt.render(w, r, "signup.tmpl", viewData{Title: "注册", Error: err.Error()})
		return
	}
	if err := rt.app.SignIn(w, user.ID); err != nil {
		rt.render(w, r, "signup.tmpl", viewData{Title: "注册", Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/teams", http.StatusFound)
}

func (rt *Router) settingsPage(w http.ResponseWriter, r *http.Request) {
	user, _, ok := rt.requireUser(w, r)
	if !ok {
		return
	}
	message := ""
	switch r.URL.Query().Get("saved") {
	case "profile":
		message = "个人资料已保存。"
	case "password":
		message = "密码已更新。"
	case "theme":
		message = "外观设置已保存。"
	}
	rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Message: message})
}

func (rt *Router) settingsProfilePost(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	err := rt.app.UpdateUserProfile(r.Context(), user, services.UpdateUserProfileOptions{
		FullName:    r.FormValue("full_name"),
		Email:       r.FormValue("email"),
		Description: r.FormValue("description"),
		Website:     r.FormValue("website"),
		Location:    r.FormValue("location"),
		AvatarURL:   r.FormValue("avatar_url"),
	})
	if err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/user/settings?saved=profile", http.StatusFound)
}

func (rt *Router) settingsPasswordPost(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	if r.FormValue("new_password") != r.FormValue("confirm_password") {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: "两次输入的新密码不一致。"})
		return
	}
	if err := rt.app.ChangeUserPassword(r.Context(), user, r.FormValue("current_password"), r.FormValue("new_password")); err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/user/settings?saved=password", http.StatusFound)
}

func (rt *Router) settingsThemePost(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	if err := rt.app.UpdateUserTheme(r.Context(), user, r.FormValue("theme")); err != nil {
		rt.render(w, r, "settings.tmpl", viewData{Title: "个人设置", ProfileUser: user, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/user/settings?saved=theme#appearance", http.StatusFound)
}

func (rt *Router) personalArticleNewPage(w http.ResponseWriter, r *http.Request) {
	user, _, ok := rt.requireUser(w, r)
	if !ok {
		return
	}
	rt.render(w, r, "personal_article_new.tmpl", viewData{Title: "写博文", ProfileUser: user})
}

func (rt *Router) personalArticleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	if err := r.ParseForm(); err != nil {
		rt.render(w, r, "personal_article_new.tmpl", viewData{Title: "写博文", ProfileUser: user, Error: err.Error()})
		return
	}
	article, err := rt.app.CreatePersonalBlogArticle(r.Context(), user, services.CreateArticleOptions{
		Title:           r.FormValue("title"),
		Slug:            r.FormValue("slug"),
		Summary:         r.FormValue("summary"),
		ContentMarkdown: r.FormValue("content_markdown"),
		Visibility:      models.Visibility(r.FormValue("visibility")),
	})
	if err != nil {
		rt.render(w, r, "personal_article_new.tmpl", viewData{Title: "写博文", ProfileUser: user, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/articles/"+strconv.FormatInt(article.ID, 10), http.StatusFound)
}

func (rt *Router) userProfilePage(w http.ResponseWriter, r *http.Request) {
	profileUser, err := rt.app.FindUserByName(r.Context(), chi.URLParam(r, "name"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	viewer, _, _ := rt.app.CurrentUser(r)
	articles, err := rt.app.ListUserBlogArticles(r.Context(), viewer, profileUser, 30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	title := profileUser.Name
	if profileUser.FullName != "" {
		title = profileUser.FullName
	}
	rt.render(w, r, "user_profile.tmpl", viewData{Title: title, ProfileUser: profileUser, Articles: articles})
}

func (rt *Router) teamsPage(w http.ResponseWriter, r *http.Request) {
	user, _, ok := rt.requireUser(w, r)
	if !ok {
		return
	}
	teams, err := rt.app.ListTeamsForUser(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rt.render(w, r, "teams.tmpl", viewData{Title: "团队", Teams: teams})
}

func (rt *Router) teamPage(w http.ResponseWriter, r *http.Request) {
	user, _, ok := rt.requireUser(w, r)
	if !ok {
		return
	}
	team, err := rt.app.GetTeamByName(r.Context(), chi.URLParam(r, "team"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	role, err := rt.app.TeamRole(context.Background(), team.ID, user)
	if err != nil || (!user.IsAdmin && role == "") {
		http.NotFound(w, r)
		return
	}
	spaces, err := rt.app.ListSpaces(r.Context(), team.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rt.render(w, r, "team.tmpl", viewData{Title: team.DisplayName, Team: team, Spaces: spaces})
}

func (rt *Router) spacePage(w http.ResponseWriter, r *http.Request) {
	user, _, ok := rt.requireUser(w, r)
	if !ok {
		return
	}
	team, err := rt.app.GetTeamByName(r.Context(), chi.URLParam(r, "team"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	space, err := rt.app.GetSpaceByName(r.Context(), team.ID, chi.URLParam(r, "space"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	articles, err := rt.app.ListArticlesInSpace(r.Context(), user, team.ID, space.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rt.render(w, r, "space.tmpl", viewData{Title: space.DisplayName, Team: team, Space: space, Articles: articles})
}

func (rt *Router) articlePage(w http.ResponseWriter, r *http.Request) {
	user, _, _ := rt.app.CurrentUser(r)
	article, err := rt.app.GetArticle(r.Context(), user, chi.URLParam(r, "idOrSlug"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rt.render(w, r, "article.tmpl", viewData{
		Title:       article.Title,
		Article:     article,
		ArticleHTML: template.HTML(article.ContentHTML),
	})
}

func (rt *Router) searchPage(w http.ResponseWriter, r *http.Request) {
	user, _, _ := rt.app.CurrentUser(r)
	query := r.URL.Query().Get("q")
	var articles []models.Article
	var err error
	if query != "" {
		articles, err = rt.app.SearchArticles(r.Context(), user, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	rt.render(w, r, "search.tmpl", viewData{Title: "搜索", Query: query, Articles: articles})
}
