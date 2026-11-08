package inertia

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goravel/framework/support"
	frameworkvite "github.com/goravel/framework/vite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestService returns the framework Vite service pointed at a temporary project,
// optionally with a build manifest in place.
func newTestService(t *testing.T, manifest string) *frameworkvite.Vite {
	t.Helper()

	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	if manifest != "" {
		path := filepath.Join(support.RelativePath, "public", "build", ".vite", "manifest.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return frameworkvite.NewVite(nil)
}

// render executes tmpl with funcs, failing the test on a parse or execute error.
func render(t *testing.T, funcs template.FuncMap, tmpl string) string {
	t.Helper()

	parsed, err := template.New("t").Funcs(funcs).Parse(tmpl)
	require.NoError(t, err)

	var b strings.Builder
	require.NoError(t, parsed.Execute(&b, nil))

	return b.String()
}

func TestViteFuncsRenderManifestTags(t *testing.T) {
	service := newTestService(t, `{"resources/inertia/app.ts":{"file":"assets/app-abc123.js","css":["assets/app-def456.css"]}}`)

	out := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, out, `<link rel="stylesheet" href="/build/assets/app-def456.css">`)
	assert.Contains(t, out, `<script type="module" src="/build/assets/app-abc123.js"></script>`)
}

func TestViteFuncsRegisterViteAsset(t *testing.T) {
	service := newTestService(t, `{"resources/images/logo.svg":{"file":"assets/logo-11223344.svg"}}`)

	out := render(t, viteFuncs(service), `{{ viteAsset "resources/images/logo.svg" }}`)

	assert.Equal(t, "/build/assets/logo-11223344.svg", out)
}

func TestViteFuncsDegradeToComment(t *testing.T) {
	service := newTestService(t, "")

	out := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, out, "<!-- vite:")
	assert.Contains(t, out, "vite manifest not found")

	// A missing asset must cost the URL, not the page.
	assert.Empty(t, render(t, viteFuncs(service), `{{ viteAsset "resources/images/logo.svg" }}`))
}

func TestViteFuncsNilServiceRendersComment(t *testing.T) {
	out := render(t, viteFuncs(nil), `{{ vite "resources/inertia/app.ts" }}{{ viteAsset "resources/images/logo.svg" }}{{ viteReactRefresh }}`)

	assert.Contains(t, out, "<!-- vite: ")
	assert.Contains(t, out, "vite facade is not initialized")
}

func TestViteFuncsDevURLFromFramework(t *testing.T) {
	// No hot file and no manifest: the dev URL is the only source of tags.
	service := newTestService(t, "").UseDevURL("http://localhost:5173")

	out := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, out, `src="http://localhost:5173/@vite/client"`)
	assert.Contains(t, out, `src="http://localhost:5173/resources/inertia/app.ts"`)
}

func TestViteFuncsDevURLReactRefresh(t *testing.T) {
	dev := newTestService(t, "").UseDevURL("http://localhost:5173")
	assert.Contains(t, render(t, viteFuncs(dev), `{{ viteReactRefresh }}`), `http://localhost:5173/@react-refresh`)

	built := newTestService(t, `{"resources/inertia/app.ts":{"file":"assets/app-abc123.js"}}`)
	assert.Empty(t, render(t, viteFuncs(built), `{{ viteReactRefresh }}`))
}

func TestViteFuncsMatchFrameworkView(t *testing.T) {
	service := newTestService(t, `{"resources/inertia/app.ts":{"file":"assets/app-abc123.js"}}`)

	adapter := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)
	view := render(t, frameworkvite.Funcs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, view, `src="/build/assets/app-abc123.js"`)
	assert.Equal(t, view, adapter)
}
