package inertia

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	petaki "github.com/petaki/inertia-go"
)

// TestV3RootTemplateRendersScriptElement renders the scaffolded vue app.gohtml
// stub through petaki and asserts the Inertia v3 CSR shape: an empty #app div and
// a <script data-page="app" type="application/json"> holding valid JSON.
func TestV3RootTemplateRendersScriptElement(t *testing.T) {
	// Strip the vite helper (not registered here) so the template parses and executes.
	tmpl := rootTemplateFromStub(t, "console/stubs/vue/app.gohtml.stub", func(body string) string {
		return strings.ReplaceAll(body, `{{ vite "resources/js/app.ts" }}`, "")
	})

	i := petaki.New("", tmpl, "v1")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil) // no X-Inertia → full HTML
	if err := i.Render(rr, req, "Home", map[string]any{"msg": "hi </script> & 'x'"}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := rr.Body.String()

	if !strings.Contains(out, `<div id="app"></div>`) {
		t.Errorf("missing empty app div; got:\n%s", out)
	}
	if !strings.Contains(out, `<script data-page="app" type="application/json">`) {
		t.Errorf("missing v3 data-page script element; got:\n%s", out)
	}
	if strings.Contains(out, `<div id="app" data-page=`) {
		t.Error("legacy v2 data-page div attribute still present")
	}
	if !strings.Contains(out, `<title data-inertia>`) {
		t.Error("head title missing data-inertia attribute")
	}
	// The script JSON must be valid and must NOT contain a raw </script> breakout.
	start := strings.Index(out, `type="application/json">`) + len(`type="application/json">`)
	end := strings.Index(out[start:], `</script>`)
	jsonBlob := out[start : start+end]
	var page map[string]any
	if err := json.Unmarshal([]byte(jsonBlob), &page); err != nil {
		t.Fatalf("script JSON invalid: %v\nblob: %s", err, jsonBlob)
	}
	if page["component"] != "Home" {
		t.Errorf("component = %v, want Home", page["component"])
	}
	if strings.Contains(jsonBlob, "</script>") {
		t.Error("raw </script> in JSON would break out of the script element")
	}
	t.Logf("rendered script JSON: %s", jsonBlob)
}

// TestRootTemplateRendersViteTags renders the real vue stub through petaki with the
// framework Vite service registered, proving {{ vite }} resolves against the manifest.
func TestRootTemplateRendersViteTags(t *testing.T) {
	service := newTestService(t, `{"resources/js/app.ts": {"file": "assets/app-abc123.js"}}`)

	i := petaki.New("", rootTemplateFromStub(t, "console/stubs/vue/app.gohtml.stub", nil), "v1")
	shareViteFuncs(i, service)

	rr := httptest.NewRecorder()
	if err := i.Render(rr, httptest.NewRequest("GET", "/", nil), "Home", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(rr.Body.String(), `src="/build/assets/app-abc123.js"`) {
		t.Errorf("root template did not render the built entry; got:\n%s", rr.Body.String())
	}
}

// TestRootTemplateRendersReactRefresh renders the react stub in dev mode, where the
// React entry needs the @vitejs/plugin-react preamble before @vite/client.
func TestRootTemplateRendersReactRefresh(t *testing.T) {
	service := useDevServer(t, newTestService(t, ""), "http://localhost:5173")

	i := petaki.New("", rootTemplateFromStub(t, "console/stubs/react/app.gohtml.stub", nil), "v1")
	shareViteFuncs(i, service)

	rr := httptest.NewRecorder()
	if err := i.Render(rr, httptest.NewRequest("GET", "/", nil), "Home", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := rr.Body.String()
	if !strings.Contains(out, "/@react-refresh") {
		t.Errorf("react root template is missing the Fast Refresh preamble; got:\n%s", out)
	}
	if strings.Index(out, "@react-refresh") > strings.Index(out, "@vite/client") {
		t.Errorf("preamble must come before @vite/client; got:\n%s", out)
	}
}

// rootTemplateFromStub copies a root-template stub into a temporary directory, which
// petaki needs as a real file path; rewrite may edit the body before it is written.
func rootTemplateFromStub(t *testing.T, stub string, rewrite func(string) string) string {
	t.Helper()

	body, err := os.ReadFile(stub)
	if err != nil {
		t.Fatal(err)
	}

	if rewrite != nil {
		body = []byte(rewrite(string(body)))
	}

	tmpl := filepath.Join(t.TempDir(), "app.gohtml")
	if err := os.WriteFile(tmpl, body, 0o644); err != nil {
		t.Fatal(err)
	}

	return tmpl
}
