package inertia

import (
	stdhttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/color"

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
//	import "github.com/goravel/inertia"
//	&inertia.ServiceProvider{}
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

		if config != nil {
			rootView = config.GetString("inertia.root_view", "app")
			version = config.GetString("inertia.version", "")
			ssr = config.GetBool("inertia.ssr", false)
			ssrURL = config.GetString("inertia.ssr_url", "http://127.0.0.1:13714/render")
			ssrTimeout = config.GetInt("inertia.ssr_timeout", 5)
			url = config.GetString("app.url", "")
			flashKeys = toStringSlice(config.Get("inertia.flash_keys"))
		}

		viteService := app.MakeVite() // nil when the framework vite module is not installed

		// A pinned inertia.version wins; otherwise the Vite manifest hash seeds the
		// version, so each build busts the client cache. The config key stays because a
		// missing or unreadable manifest yields "", which matches only clients rendered
		// with it: once those reload, the server can no longer spot a stale build.
		if version == "" && viteService != nil {
			version = viteService.ManifestHash()
		}

		engine := petaki.New(url, rootView, version)
		shareViteFuncs(engine, viteService)

		adapter := NewAdapter(engine)

		if ssr {
			engine.EnableSsr(ssrURL, &stdhttp.Client{Timeout: time.Duration(ssrTimeout) * time.Second})

			// CSR fallback engine: identical config but SSR disabled. The manager
			// renders with it when an SSR render fails, so an unreachable SSR server
			// degrades to client-side rendering instead of a blank page.
			csr := petaki.New(url, rootView, version)
			shareViteFuncs(csr, viteService)
			adapter.SetCSR(csr)
		}

		manager := NewInertiaManager(adapter, url, version, flashKeys...)

		return manager, nil
	})
}

// Boot resolves the manager, registers default shared props, and exposes the facade.
func (p *ServiceProvider) Boot(app foundation.Application) {
	// The Inertia root template gets {{ vite }}/{{ viteAsset }} from the framework
	// vite service, so without it the page renders none of its assets. Name the
	// missing provider here instead of leaving a silently asset-less page.
	if app.MakeVite() == nil {
		color.Warningln(errViteServiceMissing)
	}

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
