package services

import (
	"context"
	"fmt"
	"strings"

	"myblog/models"
)

type CreateTeamOptions struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type CreateSpaceOptions struct {
	TeamID      int64             `json:"team_id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Visibility  models.Visibility `json:"visibility"`
}

func (app *App) CreateTeam(ctx context.Context, ownerID int64, opts CreateTeamOptions) (*models.Team, error) {
	if app.Engine == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	name := strings.ToLower(strings.TrimSpace(opts.Name))
	if !validName(name) {
		return nil, fmt.Errorf("team name must be 2-80 characters and contain letters, numbers, dash, or underscore")
	}
	team := &models.Team{
		Name:        name,
		DisplayName: defaultString(strings.TrimSpace(opts.DisplayName), name),
		Description: strings.TrimSpace(opts.Description),
	}

	session := app.Engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	if _, err := session.Context(ctx).Insert(team); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	member := &models.TeamMember{
		TeamID: team.ID,
		UserID: ownerID,
		Role:   models.TeamRoleOwner,
	}
	if _, err := session.Context(ctx).Insert(member); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	return team, session.Commit()
}

func (app *App) AddTeamMember(ctx context.Context, actor *models.User, teamID int64, username string, role models.TeamRole) (*models.TeamMember, error) {
	actorRole, err := app.TeamRole(ctx, teamID, actor)
	if err != nil {
		return nil, err
	}
	if !models.CanManageTeam(actorRole) && !actor.IsAdmin {
		return nil, fmt.Errorf("permission denied")
	}
	user, err := app.FindUserByName(ctx, username)
	if err != nil {
		return nil, err
	}
	member := &models.TeamMember{
		TeamID: teamID,
		UserID: user.ID,
		Role:   models.NormalizeTeamRole(role),
	}
	if _, err := app.Engine.Context(ctx).Insert(member); err != nil {
		return nil, err
	}
	return member, nil
}

func (app *App) CreateSpace(ctx context.Context, actor *models.User, opts CreateSpaceOptions) (*models.Space, error) {
	role, err := app.TeamRole(ctx, opts.TeamID, actor)
	if err != nil {
		return nil, err
	}
	if !actor.IsAdmin && !models.CanWriteSpace(role) {
		return nil, fmt.Errorf("permission denied")
	}
	name := strings.ToLower(strings.TrimSpace(opts.Name))
	if !validName(name) {
		return nil, fmt.Errorf("space name must be 2-80 characters and contain letters, numbers, dash, or underscore")
	}
	space := &models.Space{
		TeamID:      opts.TeamID,
		Name:        name,
		DisplayName: defaultString(strings.TrimSpace(opts.DisplayName), name),
		Description: strings.TrimSpace(opts.Description),
		Visibility:  models.NormalizeVisibility(opts.Visibility),
	}
	if _, err := app.Engine.Context(ctx).Insert(space); err != nil {
		return nil, err
	}
	return space, nil
}

func (app *App) TeamRole(ctx context.Context, teamID int64, user *models.User) (models.TeamRole, error) {
	if user == nil {
		return "", nil
	}
	if user.IsAdmin {
		return models.TeamRoleOwner, nil
	}
	member := new(models.TeamMember)
	has, err := app.Engine.Context(ctx).Where("team_id = ? AND user_id = ?", teamID, user.ID).Get(member)
	if err != nil {
		return "", err
	}
	if !has {
		return "", nil
	}
	return models.NormalizeTeamRole(member.Role), nil
}

func (app *App) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	team := new(models.Team)
	has, err := app.Engine.Context(ctx).Where("name = ?", strings.ToLower(strings.TrimSpace(name))).Get(team)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("team not found")
	}
	return team, nil
}

func (app *App) ListTeamsForUser(ctx context.Context, user *models.User) ([]models.Team, error) {
	if user == nil {
		return nil, nil
	}
	var teams []models.Team
	if user.IsAdmin {
		return teams, app.Engine.Context(ctx).Find(&teams)
	}
	err := app.Engine.Context(ctx).
		Join("INNER", "team_member", "team_member.team_id = team.id").
		Where("team_member.user_id = ?", user.ID).
		Find(&teams)
	return teams, err
}

func (app *App) ListSpaces(ctx context.Context, teamID int64) ([]models.Space, error) {
	var spaces []models.Space
	err := app.Engine.Context(ctx).Where("team_id = ?", teamID).Asc("name").Find(&spaces)
	return spaces, err
}

func (app *App) GetSpaceByName(ctx context.Context, teamID int64, name string) (*models.Space, error) {
	space := new(models.Space)
	has, err := app.Engine.Context(ctx).Where("team_id = ? AND name = ?", teamID, strings.ToLower(strings.TrimSpace(name))).Get(space)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("space not found")
	}
	return space, nil
}
