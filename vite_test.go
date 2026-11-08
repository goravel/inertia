package inertia

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	contractsvite "github.com/goravel/framework/contracts/vite"
	"github.com/goravel/framework/support"
	frameworkvite "github.com/goravel/framework/vite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestService returns the framework Vite service pointed at a temporary project,
// optionally with a build manifest in place.
func newTestService(t *testing.T, manifest string) contractsvite.Vite {
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
}

func TestViteFuncsNilServiceRendersComment(t *testing.T) {
	out := render(t, viteFuncs(nil), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, out, "<!-- vite: ")
	assert.Contains(t, out, "vite facade is not initialized")
}

func TestViteFuncsDevURLFromFramework(t *testing.T) {
	service := frameworkvite.NewVite(nil).
		UseHotFile(filepath.Join(t.TempDir(), "hot")).
		UseDevURL("http://localhost:5173")

	out := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, out, `src="http://localhost:5173/@vite/client"`)
}

func TestViteFuncsMatchFrameworkView(t *testing.T) {
	service := newTestService(t, `{"resources/inertia/app.ts":{"file":"assets/app-abc123.js"}}`)

	adapter := render(t, viteFuncs(service), `{{ vite "resources/inertia/app.ts" }}`)
	view := render(t, frameworkvite.Funcs(service), `{{ vite "resources/inertia/app.ts" }}`)

	assert.Contains(t, view, `src="/build/assets/app-abc123.js"`)
	assert.Equal(t, view, adapter)
}
