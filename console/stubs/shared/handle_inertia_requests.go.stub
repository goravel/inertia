package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	inertiamw "github.com/goravel/inertia/middleware"
)

// HandleInertiaRequests is the application's Inertia middleware — the Go analogue of
// Laravel's App\Http\Middleware\HandleInertiaRequests. It is the single middleware you
// register on your web routes. The Inertia protocol (version checks, flash/validation
// error bridging, redirects) is inherited from the package via inertiamw.Handle; you
// own only the shared props returned by share() below.
//
// Inertia is session-based and meant for web routes only — never API routes. Register
// it together with the session middleware on your web route group (routes/web.go):
//
//	facades.Route().Middleware(
//	    sessionmiddleware.StartSession(),
//	    appmiddleware.HandleInertiaRequests(),
//	).Group(func(router route.Router) { ... })
func HandleInertiaRequests() http.Middleware {
	return inertiamw.Handle(inertiamw.Options{
		Share: share,
	})
}

// share defines the props shared with every Inertia response. Validation errors and
// session flash are already shared by the package; add your own props here. It runs
// once per request and is available on the client via usePage().props.
func share(ctx http.Context) map[string]any {
	return map[string]any{
		// Available on every page as usePage().props.appName.
		"appName": facades.Config().GetString("app.name"),
		// The auth object, available on every page as usePage().props.auth.user
		// (nil when nobody is logged in). Mirrors the Inertia/Laravel convention.
		"auth": map[string]any{
			"user": authUser(ctx),
		},
	}
}

// authUser returns the authenticated user from the session, or nil when nobody is
// logged in. Inertia apps are session-based: persist the user at login with
//
//	ctx.Request().Session().Put("user", user)
//
// and clear it on logout with ctx.Request().Session().Forget("user").
func authUser(ctx http.Context) any {
	if !ctx.Request().HasSession() {
		return nil
	}
	return ctx.Request().Session().Get("user")
}
