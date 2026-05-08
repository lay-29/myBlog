package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"myblog/models"
	"myblog/modules/auth"
)

type CreateUserOptions struct {
	Name     string
	FullName string
	Email    string
	Password string
	IsAdmin  bool
}

type UpdateUserProfileOptions struct {
	FullName    string
	Email       string
	Description string
	Website     string
	Location    string
	AvatarURL   string
}

func (app *App) CreateUser(ctx context.Context, opts CreateUserOptions) (*models.User, error) {
	if app.Engine == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	name := strings.TrimSpace(opts.Name)
	email := strings.ToLower(strings.TrimSpace(opts.Email))
	if !validName(name) {
		return nil, fmt.Errorf("username must be 2-80 characters and contain letters, numbers, dash, or underscore")
	}
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("valid email is required")
	}
	if len(opts.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	has, err := app.Engine.Context(ctx).Where("name = ? OR email = ?", name, email).Exist(new(models.User))
	if err != nil {
		return nil, err
	}
	if has {
		return nil, fmt.Errorf("username or email already exists")
	}

	hash, err := auth.HashPassword(opts.Password)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Name:         name,
		FullName:     strings.TrimSpace(opts.FullName),
		Email:        email,
		Theme:        "light",
		PasswordHash: hash,
		IsAdmin:      opts.IsAdmin,
	}
	if _, err := app.Engine.Context(ctx).Insert(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (app *App) Authenticate(ctx context.Context, login, password, remoteAddr string) (*models.User, error) {
	if app.Engine == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	key := loginLimiterKey(remoteAddr, login)
	if !app.LoginLimits.Allow(key) {
		return nil, fmt.Errorf("too many login attempts; try again later")
	}

	login = strings.ToLower(strings.TrimSpace(login))
	user := new(models.User)
	has, err := app.Engine.Context(ctx).Where("lower(name) = ? OR lower(email) = ?", login, login).Get(user)
	if err != nil {
		return nil, err
	}
	if !has || !auth.VerifyPassword(password, user.PasswordHash) {
		app.LoginLimits.Fail(key)
		return nil, fmt.Errorf("invalid username or password")
	}
	app.LoginLimits.Reset(key)
	return user, nil
}

func (app *App) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)
	has, err := app.Engine.Context(ctx).ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (app *App) FindUserByName(ctx context.Context, name string) (*models.User, error) {
	user := new(models.User)
	has, err := app.Engine.Context(ctx).Where("lower(name) = ?", strings.ToLower(strings.TrimSpace(name))).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (app *App) UpdateUserProfile(ctx context.Context, user *models.User, opts UpdateUserProfileOptions) error {
	if user == nil {
		return fmt.Errorf("authentication required")
	}
	email := strings.ToLower(strings.TrimSpace(opts.Email))
	if !strings.Contains(email, "@") {
		return fmt.Errorf("valid email is required")
	}
	existing := new(models.User)
	has, err := app.Engine.Context(ctx).Where("email = ? AND id <> ?", email, user.ID).Get(existing)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("email already exists")
	}

	user.FullName = strings.TrimSpace(opts.FullName)
	user.Email = email
	user.Description = strings.TrimSpace(opts.Description)
	user.Website = strings.TrimSpace(opts.Website)
	user.Location = strings.TrimSpace(opts.Location)
	user.AvatarURL = strings.TrimSpace(opts.AvatarURL)
	_, err = app.Engine.Context(ctx).
		ID(user.ID).
		Cols("full_name", "email", "description", "website", "location", "avatar_url").
		Update(user)
	return err
}

func (app *App) UpdateUserTheme(ctx context.Context, user *models.User, theme string) error {
	if user == nil {
		return fmt.Errorf("authentication required")
	}
	theme = NormalizeTheme(theme)
	user.Theme = theme
	_, err := app.Engine.Context(ctx).ID(user.ID).Cols("theme").Update(user)
	return err
}

func NormalizeTheme(theme string) string {
	switch strings.ToLower(strings.TrimSpace(theme)) {
	case "dark":
		return "dark"
	default:
		return "light"
	}
}

func (app *App) ChangeUserPassword(ctx context.Context, user *models.User, currentPassword, nextPassword string) error {
	if user == nil {
		return fmt.Errorf("authentication required")
	}
	if !auth.VerifyPassword(currentPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}
	if len(nextPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	hash, err := auth.HashPassword(nextPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	_, err = app.Engine.Context(ctx).ID(user.ID).Cols("password_hash").Update(user)
	return err
}

func (app *App) SignIn(w http.ResponseWriter, userID int64) error {
	_, err := app.Sessions.Create(w, userID)
	return err
}

func loginLimiterKey(remoteAddr, login string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return host + ":" + strings.ToLower(strings.TrimSpace(login))
}

func validName(name string) bool {
	if len(name) < 2 || len(name) > 80 {
		return false
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}
