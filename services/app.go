package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"myblog/models"
	"myblog/modules/auth"
	"myblog/modules/database"
	"myblog/modules/indexer"
	"myblog/modules/setting"
	"myblog/modules/storage"
	"xorm.io/xorm"
)

type App struct {
	mu sync.RWMutex

	Config      *setting.Config
	Engine      *xorm.Engine
	Sessions    *auth.SessionStore
	LoginLimits *auth.LoginLimiter
	Indexer     *indexer.ArticleIndexer
	Attachments *storage.AttachmentStore
}

func NewApp(cfg *setting.Config) (*App, error) {
	app := &App{
		Config:      cfg,
		Sessions:    auth.NewSessionStore(7 * 24 * time.Hour),
		LoginLimits: auth.NewLoginLimiter(),
	}
	if cfg.Security.InstallLock {
		if err := app.InitRuntime(); err != nil {
			return nil, err
		}
	}
	return app, nil
}

func (app *App) InitRuntime() error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.Engine != nil {
		return nil
	}
	engine, err := database.Open(app.Config.Database, app.Config.WorkPath)
	if err != nil {
		return err
	}
	if err := database.Migrate(engine); err != nil {
		_ = engine.Close()
		return err
	}

	idx, err := indexer.Open(app.Config.ResolvePath(app.Config.Indexer.ArticlePath))
	if err != nil {
		_ = engine.Close()
		return err
	}

	app.Engine = engine
	app.Indexer = idx
	app.Attachments = storage.NewAttachmentStore(app.Config.ResolvePath(app.Config.Storage.AttachmentsPath))
	return nil
}

func (app *App) Close() {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.Indexer != nil {
		_ = app.Indexer.Close()
		app.Indexer = nil
	}
	if app.Engine != nil {
		_ = app.Engine.Close()
		app.Engine = nil
	}
}

func (app *App) IsInstalled() bool {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.Config.Security.InstallLock
}

func (app *App) RequireRuntime() error {
	if !app.IsInstalled() {
		return fmt.Errorf("application is not installed")
	}
	if app.Engine == nil {
		return app.InitRuntime()
	}
	return nil
}

func (app *App) CurrentSession(r *http.Request) (*auth.Session, bool) {
	return app.Sessions.Current(r)
}

func (app *App) CurrentUser(r *http.Request) (*models.User, *auth.Session, bool) {
	session, ok := app.Sessions.Current(r)
	if !ok || app.Engine == nil {
		return nil, nil, false
	}
	user, err := app.GetUserByID(context.Background(), session.UserID)
	if err != nil {
		return nil, session, false
	}
	return user, session, true
}
