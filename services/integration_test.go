package services

import (
	"context"
	"path/filepath"
	"testing"

	"myblog/models"
	"myblog/modules/setting"
)

func TestSQLiteTeamArticleSearchFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	ctx := context.Background()
	admin, err := app.CreateUser(ctx, CreateUserOptions{
		Name:     "admin",
		Email:    "admin@example.com",
		Password: "password123",
		IsAdmin:  true,
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if _, err := app.Authenticate(ctx, "admin", "password123", "127.0.0.1:1234"); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	team, err := app.CreateTeam(ctx, admin.ID, CreateTeamOptions{Name: "core", DisplayName: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	space, err := app.CreateSpace(ctx, admin, CreateSpaceOptions{
		TeamID:     team.ID,
		Name:       "docs",
		Visibility: models.VisibilityTeam,
	})
	if err != nil {
		t.Fatalf("create space: %v", err)
	}
	article, err := app.CreateArticle(ctx, admin, CreateArticleOptions{
		TeamID:          team.ID,
		SpaceID:         space.ID,
		Title:           "知识库设计",
		Slug:            "kb-design",
		Summary:         "团队知识库骨架",
		ContentMarkdown: "这是一个关于 **Go + Vue** 自托管知识库的设计。",
		Visibility:      models.VisibilityTeam,
		IsBlog:          true,
	})
	if err != nil {
		t.Fatalf("create article: %v", err)
	}
	if article.ID == 0 {
		t.Fatal("expected article id")
	}
	results, err := app.SearchArticles(ctx, admin, "知识库")
	if err != nil {
		t.Fatalf("search articles: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search result")
	}
}

func TestPersonalProfileAndBlogFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	ctx := context.Background()
	user, err := app.CreateUser(ctx, CreateUserOptions{
		Name:     "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := app.UpdateUserProfile(ctx, user, UpdateUserProfileOptions{
		FullName:    "Alice Chen",
		Email:       "alice@new.example.com",
		Description: "记录工程和产品思考。",
		Website:     "https://example.com",
		Location:    "Shanghai",
	}); err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if err := app.UpdateUserTheme(ctx, user, "dark"); err != nil {
		t.Fatalf("update theme: %v", err)
	}
	if user.Theme != "dark" {
		t.Fatalf("theme = %q", user.Theme)
	}

	article, err := app.CreatePersonalBlogArticle(ctx, user, CreateArticleOptions{
		Title:           "第一篇个人博文",
		Summary:         "个人主页发布流程",
		ContentMarkdown: "从个人设置进入写博文页面，然后发布到个人主页。",
		Visibility:      models.VisibilityPublic,
	})
	if err != nil {
		t.Fatalf("create personal blog: %v", err)
	}
	if !article.IsBlog {
		t.Fatal("expected personal article to be a blog post")
	}
	if article.TeamID == 0 || article.SpaceID == 0 {
		t.Fatal("expected personal team and space to be assigned")
	}

	found, err := app.ListUserBlogArticles(ctx, nil, user, 10)
	if err != nil {
		t.Fatalf("list user blog articles: %v", err)
	}
	if len(found) != 1 || found[0].ID != article.ID {
		t.Fatalf("unexpected personal articles: %+v", found)
	}
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	cfg := setting.Default(dir)
	cfg.ConfigPath = filepath.Join(dir, "custom", "conf", "app.ini")
	cfg.Security.InstallLock = true
	cfg.Security.SecretKey = "test-secret"
	cfg.Database.Type = "sqlite3"
	cfg.Database.Path = "data/test.db"
	cfg.Indexer.ArticlePath = ""

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	return app
}
