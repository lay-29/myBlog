package services

import (
	"context"
	"fmt"
	"strings"

	"myblog/modules/database"
	"myblog/modules/indexer"
	"myblog/modules/setting"
	"myblog/modules/storage"
)

type InstallOptions struct {
	AppName string

	HTTPAddr string
	HTTPPort string
	RootURL  string

	DBType     string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBPath     string
	DBSSLMode  string

	AdminName     string
	AdminEmail    string
	AdminPassword string
}

func (app *App) Install(ctx context.Context, opts InstallOptions) error {
	if app.IsInstalled() {
		return fmt.Errorf("application is already installed")
	}
	if strings.TrimSpace(opts.AdminName) == "" || strings.TrimSpace(opts.AdminEmail) == "" || opts.AdminPassword == "" {
		return fmt.Errorf("administrator name, email, and password are required")
	}

	next := *app.Config
	next.AppName = defaultString(strings.TrimSpace(opts.AppName), app.Config.AppName)
	next.Server.HTTPAddr = defaultString(strings.TrimSpace(opts.HTTPAddr), app.Config.Server.HTTPAddr)
	next.Server.HTTPPort = defaultString(strings.TrimSpace(opts.HTTPPort), app.Config.Server.HTTPPort)
	next.Server.RootURL = defaultString(strings.TrimSpace(opts.RootURL), app.Config.Server.RootURL)
	next.Database.Type = defaultString(strings.ToLower(strings.TrimSpace(opts.DBType)), app.Config.Database.Type)
	next.Database.Host = defaultString(strings.TrimSpace(opts.DBHost), app.Config.Database.Host)
	next.Database.Port = defaultString(strings.TrimSpace(opts.DBPort), app.Config.Database.Port)
	next.Database.Name = defaultString(strings.TrimSpace(opts.DBName), app.Config.Database.Name)
	next.Database.User = strings.TrimSpace(opts.DBUser)
	next.Database.Password = opts.DBPassword
	next.Database.Path = defaultString(strings.TrimSpace(opts.DBPath), app.Config.Database.Path)
	next.Database.SSLMode = defaultString(strings.TrimSpace(opts.DBSSLMode), app.Config.Database.SSLMode)
	next.Security.SecretKey = setting.MustRandomSecret()
	next.Security.InstallLock = true
	next.Normalize()

	engine, err := database.Open(next.Database, next.WorkPath)
	if err != nil {
		return err
	}
	if err := database.Migrate(engine); err != nil {
		_ = engine.Close()
		return err
	}

	app.mu.Lock()
	oldEngine := app.Engine
	oldIndexer := app.Indexer
	app.Config = &next
	app.Engine = engine
	app.Indexer = nil
	app.Attachments = storage.NewAttachmentStore(app.Config.ResolvePath(app.Config.Storage.AttachmentsPath))
	app.mu.Unlock()

	if oldIndexer != nil {
		_ = oldIndexer.Close()
	}
	if oldEngine != nil {
		_ = oldEngine.Close()
	}

	if _, err := app.CreateUser(ctx, CreateUserOptions{
		Name:     opts.AdminName,
		Email:    opts.AdminEmail,
		Password: opts.AdminPassword,
		IsAdmin:  true,
	}); err != nil {
		return err
	}

	idx, err := indexer.Open(app.Config.ResolvePath(app.Config.Indexer.ArticlePath))
	if err != nil {
		return err
	}
	app.mu.Lock()
	app.Indexer = idx
	app.mu.Unlock()

	return app.Config.Save()
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
