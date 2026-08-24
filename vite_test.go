package inertia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViteDevTagsFromDevURL(t *testing.T) {
	v := NewVite("public", "build", filepath.Join(t.TempDir(), "hot"), "http://localhost:5173")

	html := string(v.Render("resources/js/app.ts"))

	if !strings.Contains(html, `src="http://localhost:5173/@vite/client"`) {
		t.Errorf("missing @vite/client tag: %s", html)
	}
	if !strings.Contains(html, `src="http://localhost:5173/resources/js/app.ts"`) {
		t.Errorf("missing entry tag: %s", html)
	}
}

func TestViteDevTagsReactRefreshPreamble(t *testing.T) {
	v := NewVite("public", "build", filepath.Join(t.TempDir(), "hot"), "http://localhost:5173")

	// React (.tsx) entry → preamble present, before @vite/client.
	react := string(v.Render("resources/js/app.tsx"))
	if !strings.Contains(react, "/@react-refresh") {
		t.Errorf("missing react-refresh preamble for tsx entry: %s", react)
	}
	if !strings.Contains(react, "__vite_plugin_react_preamble_installed__") {
		t.Errorf("preamble missing install flag: %s", react)
	}
	if strings.Index(react, "@react-refresh") > strings.Index(react, "@vite/client") {
		t.Errorf("preamble must come before @vite/client: %s", react)
	}

	// Vue (.ts) entry → no React preamble.
	vue := string(v.Render("resources/js/app.ts"))
	if strings.Contains(vue, "@react-refresh") {
		t.Errorf("vue entry should not emit react preamble: %s", vue)
	}
}

func TestViteHotFileTakesPrecedence(t *testing.T) {
	hot := filepath.Join(t.TempDir(), "hot")
	if err := os.WriteFile(hot, []byte("http://127.0.0.1:5199\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	v := NewVite("public", "build", hot, "http://localhost:5173")
	html := string(v.Render("resources/js/app.ts"))

	if !strings.Contains(html, "http://127.0.0.1:5199/@vite/client") {
		t.Errorf("hot file URL should win over dev_url: %s", html)
	}
	if strings.Contains(html, "localhost:5173") {
		t.Errorf("dev_url leaked while hot file present: %s", html)
	}
}

func TestViteProdTagsFromManifest(t *testing.T) {
	public := t.TempDir()
	manifestDir := filepath.Join(public, "build", ".vite")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"resources/js/app.ts": {
			"file": "assets/app-abc123.js",
			"src": "resources/js/app.ts",
			"isEntry": true,
			"css": ["assets/app-def456.css"]
		}
	}`
	if err := os.WriteFile(filepath.Join(manifestDir, "manifest.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	// hotFile points nowhere and devURL empty -> production path.
	v := NewVite(public, "build", filepath.Join(public, "hot"), "")
	html := string(v.Render("resources/js/app.ts"))

	if !strings.Contains(html, `href="/build/assets/app-def456.css"`) {
		t.Errorf("missing hashed css link: %s", html)
	}
	if !strings.Contains(html, `src="/build/assets/app-abc123.js"`) {
		t.Errorf("missing hashed js script: %s", html)
	}
}

func writeManifest(t *testing.T, public, body string) {
	t.Helper()
	dir := filepath.Join(public, "build", ".vite")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestViteVersionFromManifestHash(t *testing.T) {
	public := t.TempDir()
	writeManifest(t, public, `{"resources/js/app.ts":{"file":"assets/app-aaa.js"}}`)

	v := NewVite(public, "build", filepath.Join(public, "hot"), "")
	version := v.Version()
	if version == "" {
		t.Fatal("expected non-empty version from manifest")
	}

	// Same content -> same version (stable).
	if v2 := NewVite(public, "build", filepath.Join(public, "hot"), "").Version(); v2 != version {
		t.Errorf("version not stable: %q vs %q", version, v2)
	}

	// Changed content -> different version (cache bust).
	writeManifest(t, public, `{"resources/js/app.ts":{"file":"assets/app-bbb.js"}}`)
	if changed := NewVite(public, "build", filepath.Join(public, "hot"), "").Version(); changed == version {
		t.Errorf("version should change when manifest changes, still %q", version)
	}
}

func TestViteVersionEmptyWithoutManifest(t *testing.T) {
	public := t.TempDir()
	v := NewVite(public, "build", filepath.Join(public, "hot"), "")
	if got := v.Version(); got != "" {
		t.Errorf("Version() = %q, want empty without manifest", got)
	}
}

func TestViteProdMissingManifestRendersComment(t *testing.T) {
	public := t.TempDir()
	v := NewVite(public, "build", filepath.Join(public, "hot"), "")

	html := string(v.Render("resources/js/app.ts"))
	if !strings.Contains(html, "<!-- vite:") {
		t.Errorf("expected error comment when manifest missing, got: %s", html)
	}
}
