package inertia

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	petaki "github.com/petaki/inertia-go"
)

// writeRootTemplate writes a minimal Inertia root template and returns its path.
// It mirrors the real template: SSR body when present, empty app div otherwise.
func writeRootTemplate(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.gohtml")
	body := `<!DOCTYPE html><html><body>` +
		`{{ if .ssr }}{{ raw .ssr.Body }}{{ else }}<div id="app" data-page="{{ marshal .page }}"></div>{{ end }}` +
		`</body></html>`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// deadSSRURL returns a URL that is guaranteed to refuse connections, by starting
// a server and immediately closing it.
func deadSSRURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL + "/render"
	srv.Close()
	return url
}

func TestSSRFallbackRendersCSR(t *testing.T) {
	rootView := writeRootTemplate(t)

	primary := petaki.New("", rootView, "v1")
	primary.EnableSsr(deadSSRURL(t), &http.Client{Timeout: time.Second})
	adapter := NewAdapter(primary)

	csr := petaki.New("", rootView, "v1")
	adapter.SetCSR(csr)

	m := NewInertiaManager(adapter, "", "v1")

	// Full HTML load (no X-Inertia header) so the SSR path is exercised.
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	ctx := newFakeContext(req, rec)

	resp := m.Render(ctx, "Dashboard", map[string]any{"a": 1})
	if err := resp.Render(); err != nil {
		t.Fatalf("Render() with SSR down should fall back to CSR, got error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `id="app"`) || !strings.Contains(body, "data-page") {
		t.Errorf("expected CSR HTML with app div, got: %q", body)
	}
	if !strings.Contains(body, "Dashboard") {
		t.Errorf("expected the page component in the data-page payload, got: %q", body)
	}
}

func TestSSRFailureWithoutFallbackSurfacesError(t *testing.T) {
	rootView := writeRootTemplate(t)

	primary := petaki.New("", rootView, "v1")
	primary.EnableSsr(deadSSRURL(t), &http.Client{Timeout: time.Second})

	m := NewInertiaManager(NewAdapter(primary), "", "v1") // no CSR engine

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	ctx := newFakeContext(req, httptest.NewRecorder())

	resp := m.Render(ctx, "Dashboard", nil)
	if err := resp.Render(); err == nil {
		t.Error("without a CSR fallback engine, SSR failure should surface as an error")
	}
}

func TestNoFallbackOnSuccessfulRender(t *testing.T) {
	rootView := writeRootTemplate(t)

	// No SSR: a normal render must succeed and produce the CSR app div directly,
	// without involving the fallback path.
	primary := petaki.New("", rootView, "v1")
	m := NewInertiaManager(NewAdapter(primary), "", "v1")

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	ctx := newFakeContext(req, rec)

	resp := m.Render(ctx, "Dashboard", nil)
	if err := resp.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(rec.Body.String(), `id="app"`) {
		t.Errorf("expected app div, got: %q", rec.Body.String())
	}
}
