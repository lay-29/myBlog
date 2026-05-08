package services

import (
	"context"
	"fmt"

	"myblog/models"
)

const personalBlogSpaceName = "blog"

func PersonalTeamName(userID int64) string {
	return fmt.Sprintf("user-%d", userID)
}

func (app *App) EnsurePersonalBlogSpace(ctx context.Context, user *models.User) (*models.Team, *models.Space, error) {
	if user == nil {
		return nil, nil, fmt.Errorf("authentication required")
	}
	teamName := PersonalTeamName(user.ID)
	team := new(models.Team)
	has, err := app.Engine.Context(ctx).Where("name = ?", teamName).Get(team)
	if err != nil {
		return nil, nil, err
	}
	if !has {
		team = &models.Team{
			Name:        teamName,
			DisplayName: displayName(user) + " 的个人空间",
			Description: "系统自动创建的个人博客空间。",
		}
		session := app.Engine.NewSession()
		defer session.Close()
		if err := session.Begin(); err != nil {
			return nil, nil, err
		}
		if _, err := session.Context(ctx).Insert(team); err != nil {
			_ = session.Rollback()
			return nil, nil, err
		}
		member := &models.TeamMember{
			TeamID: team.ID,
			UserID: user.ID,
			Role:   models.TeamRoleOwner,
		}
		if _, err := session.Context(ctx).Insert(member); err != nil {
			_ = session.Rollback()
			return nil, nil, err
		}
		if err := session.Commit(); err != nil {
			return nil, nil, err
		}
	}

	space := new(models.Space)
	has, err = app.Engine.Context(ctx).Where("team_id = ? AND name = ?", team.ID, personalBlogSpaceName).Get(space)
	if err != nil {
		return nil, nil, err
	}
	if !has {
		space = &models.Space{
			TeamID:      team.ID,
			Name:        personalBlogSpaceName,
			DisplayName: "个人博客",
			Description: "个人公开博文和随笔。",
			Visibility:  models.VisibilityPublic,
		}
		if _, err := app.Engine.Context(ctx).Insert(space); err != nil {
			return nil, nil, err
		}
	}
	return team, space, nil
}

func (app *App) CreatePersonalBlogArticle(ctx context.Context, author *models.User, opts CreateArticleOptions) (*models.Article, error) {
	team, space, err := app.EnsurePersonalBlogSpace(ctx, author)
	if err != nil {
		return nil, err
	}
	opts.TeamID = team.ID
	opts.SpaceID = space.ID
	opts.IsBlog = true
	if opts.Visibility == "" {
		opts.Visibility = models.VisibilityPublic
	}
	return app.CreateArticle(ctx, author, opts)
}

func displayName(user *models.User) string {
	if user.FullName != "" {
		return user.FullName
	}
	return user.Name
}
