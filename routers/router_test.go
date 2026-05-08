package routers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"myblog/modules/setting"
	"myblog/services"
)

func TestUninstalledAppRedirectsToInstall(t *testing.T) {
	dir := t.TempDir()
	cfg := setting.Default(dir)
	cfg.ConfigPath = filepath.Join(dir, "custom", "conf", "app.ini")
	app, err := services.NewApp(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	handler, err := New(app)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/install" {
		t.Fatalf("location = %q", rec.Header().Get("Location"))
	}
}
