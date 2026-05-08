package setting

import (
	"path/filepath"
	"testing"
)

func TestConfigSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := Default(dir)
	cfg.ConfigPath = filepath.Join(dir, "custom", "conf", "app.ini")
	cfg.AppName = "Knowledge"
	cfg.Database.Type = "postgres"
	cfg.Database.Host = "db"
	cfg.Database.Port = "5432"
	cfg.Security.InstallLock = true
	cfg.Security.SecretKey = "secret"

	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	loaded, err := Load(cfg.ConfigPath, dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AppName != "Knowledge" {
		t.Fatalf("AppName = %q", loaded.AppName)
	}
	if loaded.Database.Type != "postgres" {
		t.Fatalf("DB type = %q", loaded.Database.Type)
	}
	if !loaded.Security.InstallLock {
		t.Fatal("expected install lock")
	}
}
