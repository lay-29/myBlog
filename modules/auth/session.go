package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const CookieName = "myblog_session"

type Session struct {
	Token     string
	UserID    int64
	CSRFToken string
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	ttl      time.Duration
	sessions map[string]*Session
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		ttl:      ttl,
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) Create(w http.ResponseWriter, userID int64) (*Session, error) {
	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	session := &Session{
		Token:     token,
		UserID:    userID,
		CSRFToken: csrf,
		ExpiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})
	return session, nil
}

func (s *SessionStore) Current(r *http.Request) (*Session, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return nil, false
	}
	s.mu.RLock()
	session, ok := s.sessions[cookie.Value]
	s.mu.RUnlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		if ok {
			s.mu.Lock()
			delete(s.sessions, cookie.Value)
			s.mu.Unlock()
		}
		return nil, false
	}
	return session, true
}

func (s *SessionStore) Destroy(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(CookieName); err == nil {
		s.mu.Lock()
		delete(s.sessions, cookie.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func ValidateCSRF(r *http.Request, session *Session) bool {
	if session == nil {
		return false
	}
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		token = r.FormValue("_csrf")
	}
	return token != "" && token == session.CSRFToken
}

type LoginLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
}

type loginAttempt struct {
	Count     int
	BlockedAt time.Time
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{attempts: make(map[string]loginAttempt)}
}

func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	if attempt.Count < 5 {
		return true
	}
	return time.Since(attempt.BlockedAt) > 10*time.Minute
}

func (l *LoginLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	attempt.Count++
	if attempt.Count >= 5 {
		attempt.BlockedAt = time.Now()
	}
	l.attempts[key] = attempt
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
