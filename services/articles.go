package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"myblog/models"
	"myblog/modules/indexer"
	"myblog/modules/render"
)

type CreateArticleOptions struct {
	TeamID          int64             `json:"team_id"`
	SpaceID         int64             `json:"space_id"`
	Title           string            `json:"title"`
	Slug            string            `json:"slug"`
	Summary         string            `json:"summary"`
	ContentMarkdown string            `json:"content_markdown"`
	Visibility      models.Visibility `json:"visibility"`
	IsBlog          bool              `json:"is_blog"`
}

func (app *App) CreateArticle(ctx context.Context, author *models.User, opts CreateArticleOptions) (*models.Article, error) {
	if author == nil {
		return nil, fmt.Errorf("authentication required")
	}
	role, err := app.TeamRole(ctx, opts.TeamID, author)
	if err != nil {
		return nil, err
	}
	if !author.IsAdmin && !models.CanWriteSpace(role) {
		return nil, fmt.Errorf("permission denied")
	}
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	slug := strings.TrimSpace(opts.Slug)
	if slug == "" {
		slug = slugify(title)
	}
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().UnixNano())
	}
	if !validName(slug) {
		return nil, fmt.Errorf("slug must be 2-80 characters and contain letters, numbers, dash, or underscore")
	}
	html, err := render.MarkdownToSafeHTML(opts.ContentMarkdown)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	article := &models.Article{
		TeamID:          opts.TeamID,
		SpaceID:         opts.SpaceID,
		AuthorID:        author.ID,
		Slug:            strings.ToLower(slug),
		Title:           title,
		Summary:         strings.TrimSpace(opts.Summary),
		ContentMarkdown: opts.ContentMarkdown,
		ContentHTML:     html,
		Visibility:      models.NormalizeVisibility(opts.Visibility),
		IsBlog:          opts.IsBlog,
		Published:       now,
	}
	revision := &models.ArticleRevision{
		AuthorID:        author.ID,
		Title:           article.Title,
		ContentMarkdown: article.ContentMarkdown,
	}

	session := app.Engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	if _, err := session.Context(ctx).Insert(article); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	revision.ArticleID = article.ID
	if _, err := session.Context(ctx).Insert(revision); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if err := session.Commit(); err != nil {
		return nil, err
	}
	if app.Indexer != nil {
		_ = app.Indexer.IndexArticle(article)
	}
	return article, nil
}

func (app *App) GetArticle(ctx context.Context, user *models.User, idOrSlug string) (*models.Article, error) {
	article := new(models.Article)
	var has bool
	var err error
	if id, parseErr := strconv.ParseInt(idOrSlug, 10, 64); parseErr == nil {
		has, err = app.Engine.Context(ctx).ID(id).Get(article)
	} else {
		has, err = app.Engine.Context(ctx).Where("slug = ?", strings.ToLower(strings.TrimSpace(idOrSlug))).Get(article)
	}
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("article not found")
	}
	role, err := app.TeamRole(ctx, article.TeamID, user)
	if err != nil {
		return nil, err
	}
	if !models.CanViewArticle(user, role, article.Visibility, article.AuthorID) {
		return nil, fmt.Errorf("article not found")
	}
	return article, nil
}

func (app *App) ListPublicArticles(ctx context.Context, limit int) ([]models.Article, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var articles []models.Article
	err := app.Engine.Context(ctx).
		Where("visibility = ?", models.VisibilityPublic).
		Desc("published").
		Limit(limit).
		Find(&articles)
	return articles, err
}

func (app *App) ListUserBlogArticles(ctx context.Context, viewer *models.User, author *models.User, limit int) ([]models.Article, error) {
	if author == nil {
		return nil, fmt.Errorf("author is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var articles []models.Article
	if err := app.Engine.Context(ctx).
		Where("author_id = ? AND is_blog = ?", author.ID, true).
		Desc("published").
		Limit(limit).
		Find(&articles); err != nil {
		return nil, err
	}
	visible := make([]models.Article, 0, len(articles))
	for _, article := range articles {
		role, err := app.TeamRole(ctx, article.TeamID, viewer)
		if err != nil {
			return nil, err
		}
		if models.CanViewArticle(viewer, role, article.Visibility, article.AuthorID) {
			visible = append(visible, article)
		}
	}
	return visible, nil
}

func (app *App) ListArticlesInSpace(ctx context.Context, user *models.User, teamID, spaceID int64) ([]models.Article, error) {
	var articles []models.Article
	if err := app.Engine.Context(ctx).Where("team_id = ? AND space_id = ?", teamID, spaceID).Desc("updated").Find(&articles); err != nil {
		return nil, err
	}
	role, err := app.TeamRole(ctx, teamID, user)
	if err != nil {
		return nil, err
	}
	visible := make([]models.Article, 0, len(articles))
	for _, article := range articles {
		if models.CanViewArticle(user, role, article.Visibility, article.AuthorID) {
			visible = append(visible, article)
		}
	}
	return visible, nil
}

func (app *App) SearchArticles(ctx context.Context, user *models.User, query string) ([]models.Article, error) {
	hits, err := app.Indexer.Search(ctx, query, 20)
	if err != nil {
		return nil, err
	}
	articles := make([]models.Article, 0, len(hits))
	for _, hit := range hits {
		article := new(models.Article)
		has, err := app.Engine.Context(ctx).ID(hit.ID).Get(article)
		if err != nil || !has {
			continue
		}
		role, err := app.TeamRole(ctx, article.TeamID, user)
		if err != nil {
			continue
		}
		if models.CanViewArticle(user, role, article.Visibility, article.AuthorID) {
			articles = append(articles, *article)
		}
	}
	return articles, nil
}

func (app *App) ReindexAllArticles(ctx context.Context) error {
	var articles []models.Article
	if err := app.Engine.Context(ctx).Find(&articles); err != nil {
		return err
	}
	for i := range articles {
		if err := app.Indexer.IndexArticle(&articles[i]); err != nil {
			return err
		}
	}
	return nil
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var b strings.Builder
	lastDash := false
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func ArticleSearchNgrams(text string) string {
	return indexer.Ngrams(text)
}
