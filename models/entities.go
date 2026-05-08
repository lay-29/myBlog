package models

import "time"

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityTeam    Visibility = "team"
	VisibilityPrivate Visibility = "private"
)

type TeamRole string

const (
	TeamRoleOwner  TeamRole = "owner"
	TeamRoleAdmin  TeamRole = "admin"
	TeamRoleMember TeamRole = "member"
)

type User struct {
	ID           int64     `json:"id" xorm:"pk autoincr 'id'"`
	Name         string    `json:"name" xorm:"varchar(80) unique notnull 'name'"`
	FullName     string    `json:"full_name" xorm:"varchar(120) 'full_name'"`
	Email        string    `json:"email" xorm:"varchar(255) unique notnull 'email'"`
	Description  string    `json:"description" xorm:"TEXT 'description'"`
	Website      string    `json:"website" xorm:"varchar(255) 'website'"`
	Location     string    `json:"location" xorm:"varchar(120) 'location'"`
	AvatarURL    string    `json:"avatar_url" xorm:"varchar(500) 'avatar_url'"`
	Theme        string    `json:"theme" xorm:"varchar(20) default 'light' 'theme'"`
	PasswordHash string    `json:"-" xorm:"varchar(255) notnull 'password_hash'"`
	IsAdmin      bool      `json:"is_admin" xorm:"notnull default 0 'is_admin'"`
	Created      time.Time `json:"created" xorm:"created 'created'"`
	Updated      time.Time `json:"updated" xorm:"updated 'updated'"`
}

func (User) TableName() string { return "user" }

type Team struct {
	ID          int64     `json:"id" xorm:"pk autoincr 'id'"`
	Name        string    `json:"name" xorm:"varchar(80) unique notnull 'name'"`
	DisplayName string    `json:"display_name" xorm:"varchar(120) 'display_name'"`
	Description string    `json:"description" xorm:"TEXT 'description'"`
	Created     time.Time `json:"created" xorm:"created 'created'"`
	Updated     time.Time `json:"updated" xorm:"updated 'updated'"`
}

func (Team) TableName() string { return "team" }

type TeamMember struct {
	ID      int64     `json:"id" xorm:"pk autoincr 'id'"`
	TeamID  int64     `json:"team_id" xorm:"unique(team_user) index notnull 'team_id'"`
	UserID  int64     `json:"user_id" xorm:"unique(team_user) index notnull 'user_id'"`
	Role    TeamRole  `json:"role" xorm:"varchar(20) notnull 'role'"`
	Created time.Time `json:"created" xorm:"created 'created'"`
	Updated time.Time `json:"updated" xorm:"updated 'updated'"`
}

func (TeamMember) TableName() string { return "team_member" }

type Space struct {
	ID          int64      `json:"id" xorm:"pk autoincr 'id'"`
	TeamID      int64      `json:"team_id" xorm:"unique(team_space) index notnull 'team_id'"`
	Name        string     `json:"name" xorm:"unique(team_space) varchar(80) notnull 'name'"`
	DisplayName string     `json:"display_name" xorm:"varchar(120) 'display_name'"`
	Description string     `json:"description" xorm:"TEXT 'description'"`
	Visibility  Visibility `json:"visibility" xorm:"varchar(20) notnull 'visibility'"`
	Created     time.Time  `json:"created" xorm:"created 'created'"`
	Updated     time.Time  `json:"updated" xorm:"updated 'updated'"`
}

func (Space) TableName() string { return "space" }

type Article struct {
	ID              int64      `json:"id" xorm:"pk autoincr 'id'"`
	TeamID          int64      `json:"team_id" xorm:"index notnull 'team_id'"`
	SpaceID         int64      `json:"space_id" xorm:"unique(space_slug) index notnull 'space_id'"`
	AuthorID        int64      `json:"author_id" xorm:"index notnull 'author_id'"`
	Slug            string     `json:"slug" xorm:"unique(space_slug) varchar(160) notnull 'slug'"`
	Title           string     `json:"title" xorm:"varchar(220) notnull 'title'"`
	Summary         string     `json:"summary" xorm:"TEXT 'summary'"`
	ContentMarkdown string     `json:"content_markdown" xorm:"TEXT 'content_markdown'"`
	ContentHTML     string     `json:"content_html" xorm:"TEXT 'content_html'"`
	Visibility      Visibility `json:"visibility" xorm:"varchar(20) notnull 'visibility'"`
	IsBlog          bool       `json:"is_blog" xorm:"notnull default 0 'is_blog'"`
	Published       time.Time  `json:"published" xorm:"'published'"`
	Created         time.Time  `json:"created" xorm:"created 'created'"`
	Updated         time.Time  `json:"updated" xorm:"updated 'updated'"`
}

func (Article) TableName() string { return "article" }

type ArticleRevision struct {
	ID              int64     `json:"id" xorm:"pk autoincr 'id'"`
	ArticleID       int64     `json:"article_id" xorm:"index notnull 'article_id'"`
	AuthorID        int64     `json:"author_id" xorm:"index notnull 'author_id'"`
	Title           string    `json:"title" xorm:"varchar(220) notnull 'title'"`
	ContentMarkdown string    `json:"content_markdown" xorm:"TEXT 'content_markdown'"`
	Created         time.Time `json:"created" xorm:"created 'created'"`
}

func (ArticleRevision) TableName() string { return "article_revision" }

type Attachment struct {
	ID         int64     `json:"id" xorm:"pk autoincr 'id'"`
	UUID       string    `json:"uuid" xorm:"varchar(64) unique notnull 'uuid'"`
	TeamID     int64     `json:"team_id" xorm:"index notnull 'team_id'"`
	SpaceID    int64     `json:"space_id" xorm:"index 'space_id'"`
	ArticleID  int64     `json:"article_id" xorm:"index 'article_id'"`
	UploaderID int64     `json:"uploader_id" xorm:"index notnull 'uploader_id'"`
	Name       string    `json:"name" xorm:"varchar(255) notnull 'name'"`
	Path       string    `json:"path" xorm:"varchar(500) notnull 'path'"`
	Size       int64     `json:"size" xorm:"notnull 'size'"`
	SHA256     string    `json:"sha256" xorm:"varchar(64) index notnull 'sha256'"`
	Created    time.Time `json:"created" xorm:"created 'created'"`
}

func (Attachment) TableName() string { return "attachment" }
