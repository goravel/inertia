package inertia

import (
	"errors"
	"html/template"

	contractsvite "github.com/goravel/framework/contracts/vite"
	frameworkerrors "github.com/goravel/framework/errors"
	frameworkvite "github.com/goravel/framework/vite"
	petaki "github.com/petaki/inertia-go"
)

// errViteServiceMissing tells the operator how to fix the only failure mode of the
// adapter: the framework Vite provider was never registered.
var errViteServiceMissing = errors.New("goravel/inertia: the framework vite service is not registered; add &vite.ServiceProvider{} (github.com/goravel/framework/vite) to the providers list in bootstrap/providers.go, otherwise {{ vite }} in the Inertia root template renders an empty comment")

// viteFuncs returns the template functions petaki's engine registers for the Inertia
// root template. "vite", "viteAsset" and "viteReactRefresh" come straight from the
// framework Vite service, so this template and resources/views emit the same tags for
// the same manifest, with the same cache and the same version. "vite" and "viteAsset"
// are wrapped because html/template aborts a render when a function returns an error:
// a failed build becomes an HTML comment, or an empty URL, instead of a blank page.
func viteFuncs(service contractsvite.Vite) template.FuncMap {
	if service == nil {
		// The same names as the bound path, so the root template parses either way;
		// only {{ vite }} leaves a trace of the missing provider behind.
		return template.FuncMap{
			"vite": func(_ ...string) template.HTML {
				return viteComment(frameworkerrors.ViteFacadeNotSet)
			},
			"viteAsset":        func(_ string) string { return "" },
			"viteReactRefresh": func() template.HTML { return "" },
		}
	}

	funcs := frameworkvite.Funcs(service)
	funcs["vite"] = func(entries ...string) template.HTML {
		tags, err := service.Tags(entries...)
		if err != nil {
			return viteComment(err)
		}

		return tags
	}
	// An asset error costs the URL, not the page; {{ vite }} in the same template
	// still reports the failure in its comment.
	funcs["viteAsset"] = func(path string) string {
		url, err := service.Asset(path)
		if err != nil {
			return ""
		}

		return url
	}

	return funcs
}

// shareViteFuncs registers viteFuncs on one engine; the SSR engine and the CSR
// fallback engine get the identical function set.
func shareViteFuncs(engine *petaki.Inertia, service contractsvite.Vite) {
	for name, fn := range viteFuncs(service) {
		engine.ShareFunc(name, fn)
	}
}

// viteComment renders a build failure as an inert HTML comment.
func viteComment(err error) template.HTML {
	return template.HTML("<!-- vite: " + template.HTMLEscapeString(err.Error()) + " -->")
}
