package inertia

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
)

func TestPropInProps(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Prop(ctx, "name", "Eddy")

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	if props["name"] != "Eddy" {
		t.Errorf("props[name] = %#v, want Eddy", props["name"])
	}
}

func TestDeepMergePropListed(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.DeepMerge(ctx, "tree", func() any { return map[string]any{"a": 1} }, "id")

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	deep, _ := page["deepMergeProps"].([]any)
	if len(deep) != 1 || deep[0] != "tree" {
		t.Fatalf("deepMergeProps = %#v, want [tree]", page["deepMergeProps"])
	}
	matchOn, _ := page["matchPropsOn"].([]any)
	if len(matchOn) != 1 || matchOn[0] != "tree.id" {
		t.Errorf("matchPropsOn = %#v, want [tree.id]", page["matchPropsOn"])
	}
}

func TestPrependPropListed(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Prepend(ctx, "items", func() any { return []any{"a"} })

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	prepend, _ := page["prependProps"].([]any)
	if len(prepend) != 1 || prepend[0] != "items" {
		t.Fatalf("prependProps = %#v, want [items]", page["prependProps"])
	}
	props, _ := page["props"].(map[string]any)
	if _, ok := props["items"]; !ok {
		t.Errorf("prepend value missing from props: %#v", props)
	}
}

func TestOncePropListed(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Once(ctx, "token", func() any { return "abc" })

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	once, _ := page["onceProps"].(map[string]any)
	if _, ok := once["token"]; !ok {
		t.Fatalf("onceProps = %#v, want token key", page["onceProps"])
	}
	props, _ := page["props"].(map[string]any)
	if props["token"] != "abc" {
		t.Errorf("props[token] = %#v, want abc", props["token"])
	}
}

func TestEncryptHistoryFlag(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.EncryptHistory(ctx)

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	if page["encryptHistory"] != true {
		t.Errorf("encryptHistory = %#v, want true", page["encryptHistory"])
	}
}

func TestPreserveFragmentFlag(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.PreserveFragment(ctx)

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	if page["preserveFragment"] != true {
		t.Errorf("preserveFragment = %#v, want true", page["preserveFragment"])
	}
}

func TestShareIntoProps(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.Share("appName", "Goravel")

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	if props["appName"] != "Goravel" {
		t.Errorf("shared prop appName = %#v, want Goravel", props["appName"])
	}
}

func TestShareFuncIntoProps(t *testing.T) {
	m := newTestManager("", "")
	ctx := newInertiaCtx()

	m.ShareFunc("env", func(_ contractshttp.Context) any { return "test" })

	page := inertiaJSON(t, m, ctx, "Dashboard", nil)
	props, _ := page["props"].(map[string]any)
	if props["env"] != "test" {
		t.Errorf("shared func prop env = %#v, want test", props["env"])
	}
}

func TestRedirectUsesStatusForMethod(t *testing.T) {
	m := newTestManager("", "")

	cases := map[string]int{
		http.MethodGet:    http.StatusFound,
		http.MethodPut:    http.StatusSeeOther,
		http.MethodDelete: http.StatusSeeOther,
	}
	for method, want := range cases {
		req := httptest.NewRequest(method, "/x", nil)
		ctx := newFakeContext(req, httptest.NewRecorder())

		if resp := m.Redirect(ctx, "/target"); resp == nil {
			t.Fatalf("Redirect(%s) returned nil", method)
		}
		if ctx.resp.redirectCode != want {
			t.Errorf("Redirect(%s) code = %d, want %d", method, ctx.resp.redirectCode, want)
		}
		if ctx.resp.redirectURL != "/target" {
			t.Errorf("Redirect URL = %q, want /target", ctx.resp.redirectURL)
		}
	}
}

func TestLocationInertiaHeader(t *testing.T) {
	m := newTestManager("", "")

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Inertia", "true")
	rec := httptest.NewRecorder()
	ctx := newFakeContext(req, rec)

	resp := m.Location(ctx, "https://example.com")
	if err := resp.Render(); err != nil {
		t.Fatalf("Location render error = %v", err)
	}

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
	if got := rec.Header().Get("X-Inertia-Location"); got != "https://example.com" {
		t.Errorf("X-Inertia-Location = %q, want https://example.com", got)
	}
}

func TestTemplateFuncRendersDevTags(t *testing.T) {
	v := NewVite("public", "build", filepath.Join(t.TempDir(), "nohot"), "http://localhost:5173")
	fn := v.TemplateFunc()

	html := string(fn("resources/js/app.ts"))
	if !strings.Contains(html, "http://localhost:5173/@vite/client") {
		t.Errorf("expected @vite/client tag, got: %q", html)
	}
	if !strings.Contains(html, "resources/js/app.ts") {
		t.Errorf("expected entry tag, got: %q", html)
	}
}

func TestCountingResponseWriter(t *testing.T) {
	w := &countingResponseWriter{ResponseWriter: httptest.NewRecorder()}
	if w.wrote {
		t.Fatal("wrote should be false before any write")
	}
	w.WriteHeader(http.StatusCreated)
	if !w.wrote {
		t.Error("wrote should be true after WriteHeader")
	}

	w2 := &countingResponseWriter{ResponseWriter: httptest.NewRecorder()}
	if _, err := w2.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if !w2.wrote {
		t.Error("wrote should be true after Write")
	}
}
