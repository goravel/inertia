// Command setup wires goravel-inertia into a Goravel application when installed
// via `./artisan package:install github.com/goravel/inertia`.
//
// It registers the Inertia ServiceProvider in the application (bootstrap/providers.go
// for the modern bootstrap setup, or config/app.go otherwise). Frontend scaffolding
// (Vue 3 / Vite, demo pages, config) is handled separately by `./artisan inertia:install`.
package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)

	moduleImport := setup.Paths().Module().Import()
	// The ServiceProvider lives in the package root (package goravelinertia), so it
	// is referenced by the module's package name — matching the convention of other
	// official packages (e.g. &gin.ServiceProvider{}) and avoiding a clash with the
	// application's own "providers" package in bootstrap/providers.go.
	provider := "&goravelinertia.ServiceProvider{}"
	appConfigPath := path.Config("app.go")

	setup.Install(
		// Non-bootstrap setup: register the provider in config/app.go.
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport, "goravelinertia")).
			Find(match.Providers()).Modify(modify.Register(provider))),

		// Bootstrap setup: register the provider in bootstrap/providers.go.
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RegisterProvider(moduleImport, provider)),
	).Uninstall(
		// Non-bootstrap setup: remove the provider from config/app.go.
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Providers()).Modify(modify.Unregister(provider)).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),

		// Bootstrap setup: remove the provider from bootstrap/providers.go.
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.UnregisterProvider(moduleImport, provider)),
	).Execute()
}
