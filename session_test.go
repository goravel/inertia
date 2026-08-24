package inertia

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goravel/framework/contracts/session"
)

// fakeSession implements just the read/flash surface of session.Session used by
// the manager. Unimplemented methods promote to the nil embedded interface.
type fakeSession struct {
	session.Session
	data    map[string]any
	flashed map[string]any
}

func newFakeSession(data map[string]any) *fakeSession {
	return &fakeSession{data: data, flashed: map[string]any{}}
}

func (s *fakeSession) Has(key string) bool { _, ok := s.data[key]; return ok }
func (s *fakeSession) Get(key string, defaultValue ...any) any {
	if v, ok := s.data[key]; ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}
func (s *fakeSession) Flash(key string, value any) session.Session {
	s.flashed[key] = value
	return s
}

// fakeErrors implements contracts/validation.Errors for FlashErrors.
type fakeErrors struct {
	all map[string]map[string]string
}

func (e *fakeErrors) All() map[string]map[string]string { return e.all }
func (e *fakeErrors) Get(key string) map[string]string  { return e.all[key] }
func (e *fakeErrors) Has(key string) bool               { _, ok := e.all[key]; return ok }
func (e *fakeErrors) One(key ...string) string {
	if len(key) == 0 {
		return ""
	}
	for _, msg := range e.all[key[0]] {
		return msg
	}
	return ""
}

func newSessionCtx(sess *fakeSession) *fakeContext {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("X-Inertia", "true")
	ctx := newFakeContext(req, httptest.NewRecorder())
	ctx.sess = sess
	return ctx
}

func TestShareSessionInjectsFlashAndErrors(t *testing.T) {
	m := newTestManager("", "")
	sess := newFakeSession(map[string]any{
		"success": "Saved",
		"errors":  map[string]any{"email": "is required"},
		"ignored": "not a flash key",
	})
	ctx := newSessionCtx(sess)

	m.ShareSession(ctx)
	page := inertiaJSON(t, m, ctx, "Dashboard", nil)

	props, _ := page["props"].(map[string]any)
	flash, _ := props["flash"].(map[string]any)
	if flash["success"] != "Saved" {
		t.Errorf("props.flash[success] = %#v, want Saved", flash["success"])
	}
	if _, present := flash["ignored"]; present {
		t.Errorf("non-flash-key 'ignored' leaked into flash: %#v", flash)
	}

	errs, _ := props["errors"].(map[string]any)
	if errs["email"] != "is required" {
		t.Errorf("errors[email] = %#v, want 'is required'", errs["email"])
	}
}

func TestShareSessionNoSessionIsNoop(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx() // no session attached

	m.ShareSession(ctx) // must not panic
	page := inertiaJSON(t, m, ctx, "Dashboard", nil)

	props, _ := page["props"].(map[string]any)
	if _, present := props["flash"]; present {
		t.Errorf("props.flash should be absent without a session, got %#v", props["flash"])
	}
}

func TestFlashErrorsWritesToSession(t *testing.T) {
	m := newTestManager("", "")
	sess := newFakeSession(map[string]any{})
	ctx := newSessionCtx(sess)

	m.FlashErrors(ctx, &fakeErrors{all: map[string]map[string]string{
		"email": {"required": "email is required"},
		"name":  {"min": "name too short"},
	}})

	flashed, ok := sess.flashed["errors"].(map[string]any)
	if !ok {
		t.Fatalf("errors not flashed to session: %#v", sess.flashed)
	}
	if flashed["email"] != "email is required" || flashed["name"] != "name too short" {
		t.Errorf("flashed errors = %#v", flashed)
	}
}
