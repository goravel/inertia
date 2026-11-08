package inertia

import (
	"errors"
	"html/template"

	contractsvite "github.com/goravel/framework/contracts/vite"
	contractserrors "github.com/goravel/framework/errors"
	frameworkvite "github.com/goravel/framework/vite"
	petaki "github.com/petaki/inertia-go"
)

// errViteServiceMissing tells the operator how to fix the only failure mode of the
// adapter: the framework Vite provider was never registered.
var errViteServiceMissing = errors.New("goravel/inertia: the framework vite service is not registered; add &vite.ServiceProvider{} (github.com/goravel/framework/vite) to the providers list in bootstrap/providers.go, otherwise {{ vite }} in the Inertia root template renders an empty comment")

// viteFuncs returns the template functions petaki's engine registers for the Inertia
// root template. "vite", "viteAsset" and "viteReactRefresh" come straight from the
// framework Vite service, so this template and resources/views emit the same tags for
// the same manifest, with the same cache and the same version. "vite" is wrapped
// because html/template aborts a render when a function returns an error: a failed
// build becomes an HTML comment instead of a blank page, as this package has always done.
func viteFuncs(service contractsvite.Vite) template.FuncMap {
	if service == nil {
		return template.FuncMap{"vite": func(_ ...string) template.HTML {
			return viteComment(contractserrors.ViteFacadeNotSet)
		}}
	}

	funcs := frameworkvite.Funcs(service)
	funcs["vite"] = func(entries ...string) template.HTML {
		tags, err := service.Tags(entries...)
		if err != nil {
			return viteComment(err)
		}

		return tags
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
