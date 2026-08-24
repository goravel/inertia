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
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "app.gohtml")
	stub, err := os.ReadFile("console/stubs/vue/app.gohtml.stub")
	if err != nil {
		t.Fatal(err)
	}
	// Strip the vite helper (not registered here) so template parses/executes.
	body := strings.ReplaceAll(string(stub), `{{ vite "resources/js/app.ts" }}`, "")
	if err := os.WriteFile(tmpl, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

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
