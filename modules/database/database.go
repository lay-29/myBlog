package database

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"modernc.org/sqlite"
	"myblog/models"
	"myblog/modules/setting"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

func Open(cfg setting.DatabaseConfig, workPath string) (*xorm.Engine, error) {
	driverName, dsn, err := buildDSN(cfg, workPath)
	if err != nil {
		return nil, err
	}
	engine, err := xorm.NewEngine(driverName, dsn)
	if err != nil {
		return nil, err
	}
	engine.SetMapper(names.SnakeMapper{})
	engine.ShowSQL(cfg.LogSQL)
	if err := engine.Ping(); err != nil {
		_ = engine.Close()
		return nil, err
	}
	return engine, nil
}

func Migrate(engine *xorm.Engine) error {
	return engine.Sync2(
		new(models.User),
		new(models.Team),
		new(models.TeamMember),
		new(models.Space),
		new(models.Article),
		new(models.ArticleRevision),
		new(models.Attachment),
	)
}

func buildDSN(cfg setting.DatabaseConfig, workPath string) (string, string, error) {
	switch strings.ToLower(cfg.Type) {
	case "sqlite", "sqlite3", "":
		dbPath := cfg.Path
		if dbPath == "" {
			dbPath = "data/myblog.db"
		}
		if !filepath.IsAbs(dbPath) {
			dbPath = filepath.Join(workPath, dbPath)
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return "", "", err
		}
		return "sqlite3", dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", nil
	case "mysql":
		port := cfg.Port
		if port == "" {
			port = "3306"
		}
		if cfg.Name == "" {
			return "", "", fmt.Errorf("database name is required for mysql")
		}
		addr := net.JoinHostPort(defaultString(cfg.Host, "127.0.0.1"), port)
		return "mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.User, cfg.Password, addr, cfg.Name), nil
	case "postgres", "postgresql":
		port := cfg.Port
		if port == "" {
			port = "5432"
		}
		if cfg.Name == "" {
			return "", "", fmt.Errorf("database name is required for postgres")
		}
		u := url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort(defaultString(cfg.Host, "127.0.0.1"), port),
			Path:   cfg.Name,
		}
		if cfg.User != "" {
			u.User = url.UserPassword(cfg.User, cfg.Password)
		}
		q := u.Query()
		q.Set("sslmode", defaultString(cfg.SSLMode, "disable"))
		u.RawQuery = q.Encode()
		return "postgres", u.String(), nil
	default:
		return "", "", fmt.Errorf("unsupported database type %q", cfg.Type)
	}
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
