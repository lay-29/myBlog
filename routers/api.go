package routers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"myblog/models"
	"myblog/services"
)

func (rt *Router) apiRoutes(r chi.Router) {
	r.Get("/search", rt.apiSearch)
	r.Post("/teams", rt.apiCreateTeam)
	r.Post("/teams/{teamID}/members", rt.apiAddTeamMember)
	r.Post("/spaces", rt.apiCreateSpace)
	r.Post("/articles", rt.apiCreateArticle)
	r.Post("/attachments", rt.apiUploadAttachment)
}

func (rt *Router) apiSearch(w http.ResponseWriter, r *http.Request) {
	user, _, _ := rt.app.CurrentUser(r)
	articles, err := rt.app.SearchArticles(r.Context(), user, r.URL.Query().Get("q"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": articles})
}

func (rt *Router) apiCreateTeam(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	team, err := rt.app.CreateTeam(r.Context(), user.ID, services.CreateTeamOptions{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": team})
}

func (rt *Router) apiAddTeamMember(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	teamID, _ := strconv.ParseInt(chi.URLParam(r, "teamID"), 10, 64)
	var req struct {
		Username string          `json:"username"`
		Role     models.TeamRole `json:"role"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	member, err := rt.app.AddTeamMember(r.Context(), user, teamID, req.Username, req.Role)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": member})
}

func (rt *Router) apiCreateSpace(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	var req struct {
		TeamID      int64             `json:"team_id"`
		Name        string            `json:"name"`
		DisplayName string            `json:"display_name"`
		Description string            `json:"description"`
		Visibility  models.Visibility `json:"visibility"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	space, err := rt.app.CreateSpace(r.Context(), user, services.CreateSpaceOptions{
		TeamID:      req.TeamID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Visibility:  req.Visibility,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": space})
}

func (rt *Router) apiCreateArticle(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	var req services.CreateArticleOptions
	if !decodeBody(w, r, &req) {
		return
	}
	article, err := rt.app.CreateArticle(r.Context(), user, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": article})
}

func (rt *Router) apiUploadAttachment(w http.ResponseWriter, r *http.Request) {
	user, session, ok := rt.requireUser(w, r)
	if !ok || !rt.requireCSRF(w, r, session) {
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	teamID := parseInt64(r.FormValue("team_id"))
	role, err := rt.app.TeamRole(r.Context(), teamID, user)
	if err != nil || (!user.IsAdmin && !models.CanWriteSpace(role)) {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "permission denied"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	defer file.Close()

	stored, err := rt.app.Attachments.Save(header.Filename, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	attachment := &models.Attachment{
		UUID:       stored.UUID,
		TeamID:     teamID,
		SpaceID:    parseInt64(r.FormValue("space_id")),
		ArticleID:  parseInt64(r.FormValue("article_id")),
		UploaderID: user.ID,
		Name:       stored.Name,
		Path:       stored.Path,
		Size:       stored.Size,
		SHA256:     stored.SHA256,
	}
	if _, err := rt.app.Engine.Context(r.Context()).Insert(attachment); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": attachment})
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return false
	}
	return true
}
