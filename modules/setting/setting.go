package setting

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	ConfigPath string
	WorkPath   string

	AppName  string
	RunMode  string
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Service  ServiceConfig
	Storage  StorageConfig
	Indexer  IndexerConfig
}

type ServerConfig struct {
	HTTPAddr string
	HTTPPort string
	RootURL  string
}

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	Path     string
	SSLMode  string
	LogSQL   bool
}

type SecurityConfig struct {
	InstallLock bool
	SecretKey   string
}

type ServiceConfig struct {
	DisableRegistration bool
}

type StorageConfig struct {
	AttachmentsPath string
}

type IndexerConfig struct {
	ArticlePath string
}

func Default(workPath string) *Config {
	return &Config{
		WorkPath: workPath,
		AppName:  "MyBlog",
		RunMode:  "prod",
		Server: ServerConfig{
			HTTPAddr: "0.0.0.0",
			HTTPPort: "3000",
			RootURL:  "http://localhost:3000/",
		},
		Database: DatabaseConfig{
			Type:    "sqlite3",
			Host:    "127.0.0.1",
			Port:    "5432",
			Name:    "myblog",
			User:    "myblog",
			Path:    "data/myblog.db",
			SSLMode: "disable",
		},
		Service: ServiceConfig{
			DisableRegistration: false,
		},
		Storage: StorageConfig{
			AttachmentsPath: "data/attachments",
		},
		Indexer: IndexerConfig{
			ArticlePath: "data/indexers/articles",
		},
	}
}

func Load(configPath, workPath string) (*Config, error) {
	if workPath == "" {
		workPath = "."
	}
	absWorkPath, err := filepath.Abs(workPath)
	if err != nil {
		return nil, err
	}

	if configPath == "" {
		configPath = "custom/conf/app.ini"
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(absWorkPath, configPath)
	}

	cfg := Default(absWorkPath)
	cfg.ConfigPath = configPath

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			cfg.Normalize()
			return cfg, nil
		}
		return nil, err
	}

	iniFile, err := ini.Load(configPath)
	if err != nil {
		return nil, err
	}

	cfg.AppName = iniFile.Section("").Key("APP_NAME").MustString(cfg.AppName)
	cfg.RunMode = iniFile.Section("").Key("RUN_MODE").MustString(cfg.RunMode)

	server := iniFile.Section("server")
	cfg.Server.HTTPAddr = server.Key("HTTP_ADDR").MustString(cfg.Server.HTTPAddr)
	cfg.Server.HTTPPort = server.Key("HTTP_PORT").MustString(cfg.Server.HTTPPort)
	cfg.Server.RootURL = server.Key("ROOT_URL").MustString(cfg.Server.RootURL)

	db := iniFile.Section("database")
	cfg.Database.Type = db.Key("DB_TYPE").MustString(cfg.Database.Type)
	cfg.Database.Host = db.Key("HOST").MustString(cfg.Database.Host)
	cfg.Database.Port = db.Key("PORT").MustString(cfg.Database.Port)
	cfg.Database.Name = db.Key("NAME").MustString(cfg.Database.Name)
	cfg.Database.User = db.Key("USER").MustString(cfg.Database.User)
	cfg.Database.Password = db.Key("PASSWD").MustString(cfg.Database.Password)
	cfg.Database.Path = db.Key("PATH").MustString(cfg.Database.Path)
	cfg.Database.SSLMode = db.Key("SSL_MODE").MustString(cfg.Database.SSLMode)
	cfg.Database.LogSQL = db.Key("LOG_SQL").MustBool(cfg.Database.LogSQL)

	security := iniFile.Section("security")
	cfg.Security.InstallLock = security.Key("INSTALL_LOCK").MustBool(cfg.Security.InstallLock)
	cfg.Security.SecretKey = security.Key("SECRET_KEY").MustString(cfg.Security.SecretKey)

	service := iniFile.Section("service")
	cfg.Service.DisableRegistration = service.Key("DISABLE_REGISTRATION").MustBool(cfg.Service.DisableRegistration)

	storage := iniFile.Section("storage")
	cfg.Storage.AttachmentsPath = storage.Key("ATTACHMENTS_PATH").MustString(cfg.Storage.AttachmentsPath)

	indexer := iniFile.Section("indexer")
	cfg.Indexer.ArticlePath = indexer.Key("ARTICLE_PATH").MustString(cfg.Indexer.ArticlePath)

	cfg.Normalize()
	return cfg, nil
}

func (cfg *Config) Normalize() {
	cfg.Database.Type = strings.ToLower(strings.TrimSpace(cfg.Database.Type))
	if cfg.Database.Type == "sqlite" {
		cfg.Database.Type = "sqlite3"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Security.SecretKey == "" && cfg.Security.InstallLock {
		cfg.Security.SecretKey = MustRandomSecret()
	}
}

func (cfg *Config) Save() error {
	if cfg.ConfigPath == "" {
		return fmt.Errorf("config path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ConfigPath), 0o755); err != nil {
		return err
	}

	iniFile := ini.Empty()
	iniFile.Section("").Key("APP_NAME").SetValue(cfg.AppName)
	iniFile.Section("").Key("RUN_MODE").SetValue(cfg.RunMode)

	server := iniFile.Section("server")
	server.Key("HTTP_ADDR").SetValue(cfg.Server.HTTPAddr)
	server.Key("HTTP_PORT").SetValue(cfg.Server.HTTPPort)
	server.Key("ROOT_URL").SetValue(cfg.Server.RootURL)

	db := iniFile.Section("database")
	db.Key("DB_TYPE").SetValue(cfg.Database.Type)
	db.Key("HOST").SetValue(cfg.Database.Host)
	db.Key("PORT").SetValue(cfg.Database.Port)
	db.Key("NAME").SetValue(cfg.Database.Name)
	db.Key("USER").SetValue(cfg.Database.User)
	db.Key("PASSWD").SetValue(cfg.Database.Password)
	db.Key("PATH").SetValue(cfg.Database.Path)
	db.Key("SSL_MODE").SetValue(cfg.Database.SSLMode)
	db.Key("LOG_SQL").SetValue(boolString(cfg.Database.LogSQL))

	security := iniFile.Section("security")
	security.Key("INSTALL_LOCK").SetValue(boolString(cfg.Security.InstallLock))
	security.Key("SECRET_KEY").SetValue(cfg.Security.SecretKey)

	service := iniFile.Section("service")
	service.Key("DISABLE_REGISTRATION").SetValue(boolString(cfg.Service.DisableRegistration))

	storage := iniFile.Section("storage")
	storage.Key("ATTACHMENTS_PATH").SetValue(cfg.Storage.AttachmentsPath)

	indexer := iniFile.Section("indexer")
	indexer.Key("ARTICLE_PATH").SetValue(cfg.Indexer.ArticlePath)

	return iniFile.SaveTo(cfg.ConfigPath)
}

func (cfg *Config) ResolvePath(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cfg.WorkPath, path)
}

func (cfg ServerConfig) ListenAddr() string {
	host := cfg.HTTPAddr
	if host == "" {
		host = "0.0.0.0"
	}
	port := cfg.HTTPPort
	if port == "" {
		port = "3000"
	}
	return net.JoinHostPort(host, port)
}

func RandomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func MustRandomSecret() string {
	secret, err := RandomSecret()
	if err != nil {
		panic(err)
	}
	return secret
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
