package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestAuth_Login_redirectsWhenAuthenticated(t *testing.T) {
	s := setupTestStore(t)
	sessions := s.Sessions()
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), sessions, cais.Config{}, testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestAuth_LoginPost_invalidCredentials(t *testing.T) {
	s := setupTestStore(t)
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, testCatalog())

	form := url.Values{"email": {"nobody@example.com"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("body missing error: %s", rr.Body.String())
	}
}
