package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/session"
	petaki "github.com/petaki/inertia-go"

	goravelinertia "github.com/goravel/inertia"
	"github.com/goravel/inertia/facades"
)

// newManager builds a real InertiaManager and registers it on the facade so the
// middleware (which calls facades.Inertia()) resolves it.
func newManager(t *testing.T, version string) {
	t.Helper()
	adapter := goravelinertia.NewAdapter(petaki.New("", "app", version))
	facades.RegisterInertia(goravelinertia.NewInertiaManager(adapter, "", version))
}

// fakeContext implements just enough of contractshttp.Context for the middleware
// and for a follow-up Render. The request/response are persistent so Next/Abort/
// Header calls are observable. context.Context is forwarded, not embedded, because
// the interface also exposes a Context() method.
type fakeContext struct {
	base context.Context
	vals map[any]any
	req  *fakeRequest
	resp *fakeResponse
}

func newFakeContext(r *http.Request, w http.ResponseWriter) *fakeContext {
	return &fakeContext{
		base: context.Background(),
		vals: map[any]any{},
		req:  &fakeRequest{r: r},
		resp: &fakeResponse{w: w, headers: map[string]string{}},
	}
}

func (f *fakeContext) Deadline() (time.Time, bool)             { return f.base.Deadline() }
func (f *fakeContext) Done() <-chan struct{}                   { return f.base.Done() }
func (f *fakeContext) Err() error                              { return f.base.Err() }
func (f *fakeContext) Value(key any) any                       { return f.vals[key] }
func (f *fakeContext) Context() context.Context                { return f.base }
func (f *fakeContext) WithContext(c context.Context)           { f.base = c }
func (f *fakeContext) WithValue(k, v any)                      { f.vals[k] = v }
func (f *fakeContext) Request() contractshttp.ContextRequest   { return f.req }
func (f *fakeContext) Response() contractshttp.ContextResponse { return f.resp }

type fakeRequest struct {
	contractshttp.ContextRequest
	r          *http.Request
	nextCalled bool
	abortCode  int
	aborted    bool
}

func (f *fakeRequest) Origin() *http.Request    { return f.r }
func (f *fakeRequest) Method() string           { return f.r.Method }
func (f *fakeRequest) FullUrl() string          { return f.r.URL.String() }
func (f *fakeRequest) HasSession() bool         { return false }
func (f *fakeRequest) Session() session.Session { return nil }
func (f *fakeRequest) Next()                    { f.nextCalled = true }
func (f *fakeRequest) Abort(code ...int) {
	f.aborted = true
	if len(code) > 0 {
		f.abortCode = code[0]
	}
}
func (f *fakeRequest) Header(key string, defaultValue ...string) string {
	if v := f.r.Header.Get(key); v != "" {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

type fakeResponse struct {
	contractshttp.ContextResponse
	w       http.ResponseWriter
	headers map[string]string
}

func (f *fakeResponse) Writer() http.ResponseWriter { return f.w }
func (f *fakeResponse) Header(key, value string) contractshttp.ContextResponse {
	f.headers[key] = value
	return f
}

func TestHandlePassesThroughNonInertia(t *testing.T) {
	newManager(t, "v1")
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil) // no X-Inertia
	ctx := newFakeContext(req, httptest.NewRecorder())

	Handle(Options{}).Handle(ctx)

	if !ctx.req.nextCalled {
		t.Error("Next() not called on a non-Inertia request")
	}
	if ctx.req.aborted {
		t.Error("non-Inertia request should not abort")
	}
}

func TestHandleVersionMismatchAborts(t *testing.T) {
	newManager(t, "v1")
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "stale")
	ctx := newFakeContext(req, httptest.NewRecorder())

	Handle(Options{}).Handle(ctx)

	if !ctx.req.aborted || ctx.req.abortCode != http.StatusConflict {
		t.Errorf("expected Abort(409), got aborted=%v code=%d", ctx.req.aborted, ctx.req.abortCode)
	}
	if got := ctx.resp.headers["X-Inertia-Location"]; got != "/dashboard" {
		t.Errorf("X-Inertia-Location = %q, want /dashboard", got)
	}
	if ctx.req.nextCalled {
		t.Error("Next() must not be called after a version-mismatch abort")
	}
}

func TestHandleAppliesShareProps(t *testing.T) {
	newManager(t, "v1")
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "v1") // matching → no 409
	ctx := newFakeContext(req, httptest.NewRecorder())

	called := false
	Handle(Options{Share: func(c contractshttp.Context) map[string]any {
		called = true
		return map[string]any{"auth": "user-1"}
	}}).Handle(ctx)

	if !called {
		t.Fatal("Share callback was not invoked")
	}
	if !ctx.req.nextCalled {
		t.Error("Next() not called on a valid Inertia request")
	}

	// The shared prop must land in the rendered page props.
	page := renderPage(t, ctx)
	props, _ := page["props"].(map[string]any)
	if props["auth"] != "user-1" {
		t.Errorf("props[auth] = %#v, want user-1", props["auth"])
	}
}

func TestInertiaIsHandleWithoutShare(t *testing.T) {
	newManager(t, "v1")
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "v1")
	ctx := newFakeContext(req, httptest.NewRecorder())

	Inertia().Handle(ctx)

	if !ctx.req.nextCalled {
		t.Error("Inertia() should pass a valid request through to Next()")
	}
	if ctx.req.aborted {
		t.Error("Inertia() should not abort a version-matching request")
	}
}

func TestMiddlewareSignature(t *testing.T) {
	if got := Handle(Options{}).Signature(); got != "goravel-inertia:handle_inertia_requests" {
		t.Errorf("Signature() = %q, want goravel-inertia:handle_inertia_requests", got)
	}
}

// renderPage renders through the registered manager and returns the decoded page.
func renderPage(t *testing.T, ctx *fakeContext) map[string]any {
	t.Helper()
	resp := facades.Inertia().Render(ctx, "Dashboard", nil)
	if err := resp.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	rec, ok := ctx.resp.w.(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("writer is not a ResponseRecorder")
	}
	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("unmarshal page: %v (body=%q)", err, rec.Body.String())
	}
	return page
}
