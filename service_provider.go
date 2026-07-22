package goravelinertia

import (
	stdhttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"

	petaki "github.com/petaki/inertia-go"

	inertiaconsole "github.com/goravel/inertia/console"
	"github.com/goravel/inertia/contracts"
	"github.com/goravel/inertia/facades"
)

// ServiceProvider registers and boots the Inertia singleton and facade.
//
// Registered in a Goravel application via package:install (which wires it into
// bootstrap/providers.go automatically) or manually:
//
//	import goravelinertia "github.com/goravel/inertia"
//	&goravelinertia.ServiceProvider{}
type ServiceProvider struct {
}

// toStringSlice converts a config value (which may be []string or []any) into a
// []string, returning nil so the manager falls back to its default flash keys.
func toStringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

// Register binds the Inertia manager as the "goravel.inertia" singleton.
func (p *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("goravel.inertia", func(app foundation.Application) (any, error) {
		config := app.MakeConfig()

		rootView := "app"
		version := ""
		ssr := false
		ssrURL := "http://127.0.0.1:13714/render"
		ssrTimeout := 5
		url := ""
		var flashKeys []string

		vitePublic := "public"
		viteBuild := "build"
		viteHot := ""
		viteDevURL := ""

		if config != nil {
			rootView = config.GetString("inertia.root_view", "app")
			version = config.GetString("inertia.version", "")
			ssr = config.GetBool("inertia.ssr", false)
			ssrURL = config.GetString("inertia.ssr_url", "http://127.0.0.1:13714/render")
			ssrTimeout = config.GetInt("inertia.ssr_timeout", 5)
			url = config.GetString("app.url", "")
			flashKeys = toStringSlice(config.Get("inertia.flash_keys"))

			vitePublic = config.GetString("inertia.vite.public_path", vitePublic)
			viteBuild = config.GetString("inertia.vite.build_dir", viteBuild)
			viteHot = config.GetString("inertia.vite.hot_file", viteHot)
			viteDevURL = config.GetString("inertia.vite.dev_url", viteDevURL)
		}

		vite := NewVite(vitePublic, viteBuild, viteHot, viteDevURL)

		// Auto-derive the asset version from the build manifest when not pinned in
		// config, so a new build invalidates the client cache automatically.
		if version == "" {
			version = vite.Version()
		}

		inertia := petaki.New(url, rootView, version)
		inertia.ShareFunc("vite", vite.TemplateFunc())

		adapter := NewAdapter(inertia)

		if ssr {
			inertia.EnableSsr(ssrURL, &stdhttp.Client{Timeout: time.Duration(ssrTimeout) * time.Second})

			// CSR fallback engine: identical config but SSR disabled. The manager
			// renders with it when an SSR render fails, so an unreachable SSR server
			// degrades to client-side rendering instead of a blank page.
			csr := petaki.New(url, rootView, version)
			csr.ShareFunc("vite", vite.TemplateFunc())
			adapter.SetCSR(csr)
		}

		manager := NewInertiaManager(adapter, url, version, flashKeys...)

		return manager, nil
	})
}

// Boot resolves the manager, registers default shared props, and exposes the facade.
func (p *ServiceProvider) Boot(app foundation.Application) {
	instance, err := app.Make("goravel.inertia")
	if err != nil {
		panic("Failed to resolve goravel.inertia: " + err.Error())
	}

	inertia := instance.(contracts.Inertia)

	config := app.MakeConfig()
	if config != nil {
		inertia.Share("appName", config.GetString("app.name"))
	}

	inertia.ShareFunc("timestamp", func(_ contractshttp.Context) any {
		return time.Now().Format("2006-01-02 15:04:05")
	})

	facades.RegisterInertia(inertia)

	app.MakeArtisan().Register([]console.Command{
		inertiaconsole.NewInstallCommand(),
	})
}
