package inertia

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/session"

	"github.com/goravel/inertia/contracts"
)

// fakeContext implements just enough of contractshttp.Context for the manager.
// context.Context is not embedded anonymously because the interface also exposes
// a Context() method, which would collide with the embedded field name.
type fakeContext struct {
	base context.Context
	vals map[any]any
	req  *http.Request
	w    http.ResponseWriter
	sess *fakeSession
	resp *fakeResponse
}

func newFakeContext(req *http.Request, w http.ResponseWriter) *fakeContext {
	return &fakeContext{base: context.Background(), vals: map[any]any{}, req: req, w: w}
}

func (f *fakeContext) Deadline() (time.Time, bool)   { return f.base.Deadline() }
func (f *fakeContext) Done() <-chan struct{}         { return f.base.Done() }
func (f *fakeContext) Err() error                    { return f.base.Err() }
func (f *fakeContext) Value(key any) any             { return f.vals[key] }
func (f *fakeContext) Context() context.Context      { return f.base }
func (f *fakeContext) WithContext(c context.Context) { f.base = c }
func (f *fakeContext) WithValue(k, v any)            { f.vals[k] = v }
func (f *fakeContext) Request() contractshttp.ContextRequest {
	return &fakeRequest{r: f.req, sess: f.sess}
}
func (f *fakeContext) Response() contractshttp.ContextResponse {
	if f.resp == nil {
		f.resp = &fakeResponse{w: f.w}
	}
	return f.resp
}

type fakeRequest struct {
	contractshttp.ContextRequest
	r    *http.Request
	sess *fakeSession
}

func (f *fakeRequest) Origin() *http.Request { return f.r }
func (f *fakeRequest) Method() string        { return f.r.Method }
func (f *fakeRequest) HasSession() bool      { return f.sess != nil }
func (f *fakeRequest) Session() session.Session {
	return f.sess
}

type fakeResponse struct {
	contractshttp.ContextResponse
	w            http.ResponseWriter
	redirectCode int
	redirectURL  string
}

func (f *fakeResponse) Writer() http.ResponseWriter { return f.w }

func (f *fakeResponse) Redirect(code int, url string) contractshttp.AbortableResponse {
	f.redirectCode = code
	f.redirectURL = url
	return &fakeAbortable{}
}

// fakeAbortable is a no-op AbortableResponse returned by fakeResponse.Redirect.
type fakeAbortable struct{}

func (f *fakeAbortable) Render() error { return nil }
func (f *fakeAbortable) Abort() error  { return nil }

func inertiaJSON(t *testing.T, m *InertiaManager, ctx contractshttp.Context, component string, props map[string]any) map[string]any {
	t.Helper()

	resp := m.Render(ctx, component, props)
	if err := resp.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	rec, ok := ctx.Response().Writer().(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("writer is not a ResponseRecorder")
	}

	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("unmarshal page: %v (body=%q)", err, rec.Body.String())
	}

	return page
}

func newInertiaCtx() *fakeContext {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("X-Inertia", "true") // force the JSON branch, no template needed
	return newFakeContext(req, httptest.NewRecorder())
}

func TestDeferThreadsIntoRenderedJSON(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Defer(ctx, "stats", func() any { return map[string]any{"x": 1} }, "metrics")

	page := inertiaJSON(t, m, ctx, "Dashboard", map[string]any{"a": 1})

	deferred, ok := page["deferredProps"].(map[string]any)
	if !ok {
		t.Fatalf("deferredProps missing or wrong type: %#v", page["deferredProps"])
	}

	group, ok := deferred["metrics"].([]any)
	if !ok || len(group) != 1 || group[0] != "stats" {
		t.Fatalf("deferredProps[metrics] = %#v, want [stats]", deferred["metrics"])
	}

	// deferred prop must NOT be evaluated into props on the initial load.
	props, _ := page["props"].(map[string]any)
	if _, present := props["stats"]; present {
		t.Errorf("deferred prop 'stats' should be absent from initial props, got %#v", props)
	}
}

func TestOptionalAbsentOnFullLoad(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Optional(ctx, "expensive", func() any { return 42 })

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	if _, present := props["expensive"]; present {
		t.Errorf("optional prop should be absent on full load, got %#v", props["expensive"])
	}
}

func TestAlwaysPresentInProps(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Always(ctx, "now", func() any { return "tick" })

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	if props["now"] != "tick" {
		t.Errorf("always prop = %#v, want tick", props["now"])
	}
}

func TestMergePropListedAndValued(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Merge(ctx, "items", func() any { return []any{"a"} }, "id")

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)

	mergeProps, _ := page["mergeProps"].([]any)
	if len(mergeProps) != 1 || mergeProps[0] != "items" {
		t.Fatalf("mergeProps = %#v, want [items]", page["mergeProps"])
	}
	props, _ := page["props"].(map[string]any)
	if _, present := props["items"]; !present {
		t.Errorf("merge prop value missing from props: %#v", props)
	}
	matchOn, _ := page["matchPropsOn"].([]any)
	if len(matchOn) != 1 || matchOn[0] != "items.id" {
		t.Errorf("matchPropsOn = %#v, want [items.id]", page["matchPropsOn"])
	}
}

func TestErrorIntoProps(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Error(ctx, "email", "is required")

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	errs, _ := props["errors"].(map[string]any)
	if errs["email"] != "is required" {
		t.Errorf("errors[email] = %#v, want 'is required'", errs["email"])
	}
}

func TestFlashIntoPage(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Flash(ctx, map[string]any{"success": "saved"})

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	flash, _ := props["flash"].(map[string]any)
	if flash["success"] != "saved" {
		t.Errorf("props.flash[success] = %#v, want saved", flash["success"])
	}
}

func TestClearHistoryFlag(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.ClearHistory(ctx)

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	if page["clearHistory"] != true {
		t.Errorf("clearHistory = %#v, want true", page["clearHistory"])
	}
}

func TestScrollProp(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Scroll(ctx, "feed", contracts.ScrollProp{PageName: "page", CurrentPage: 2})

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	scroll, _ := page["scrollProps"].(map[string]any)
	feed, ok := scroll["feed"].(map[string]any)
	if !ok || feed["pageName"] != "page" {
		t.Errorf("scrollProps[feed] = %#v, want pageName=page", scroll["feed"])
	}
}

func TestRenderWithoutDeferIsClean(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	page := inertiaJSON(t, m, ctx, "Dashboard", map[string]any{"a": float64(1)})

	if _, present := page["deferredProps"]; present {
		t.Errorf("deferredProps should be absent when nothing deferred, got %#v", page["deferredProps"])
	}
	props, _ := page["props"].(map[string]any)
	if props["a"] != float64(1) {
		t.Errorf("props[a] = %#v, want 1", props["a"])
	}
}
