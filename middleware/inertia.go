package middleware

import (
	stdhttp "net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"github.com/goravel/inertia/facades"
)

// Options configures the Inertia middleware. It is the Go equivalent of the
// overridable hooks on Laravel's base Inertia\Middleware class: an application
// supplies its own Share callback (the analogue of overriding share()) while the
// protocol itself stays in the package and is inherited via Handle.
type Options struct {
	// Share returns the props shared with every Inertia response. It runs once per
	// request (ctx-aware), mirroring Laravel's HandleInertiaRequests::share().
	// Returned values are applied as per-request props before the handler runs.
	Share func(ctx contractshttp.Context) map[string]any
}

// inertiaMiddleware carries the Inertia protocol as a contractshttp.Middleware
// implementation (Goravel v1.18 changed Middleware from a func type to an
// interface with Handle and Signature methods).
type inertiaMiddleware struct {
	opts Options
}

// Signature identifies this middleware within Goravel's HTTP pipeline.
func (m *inertiaMiddleware) Signature() string {
	return "goravel-inertia:handle_inertia_requests"
}

// Handle drives the Inertia protocol for a single request.
//
// Non-Inertia requests pass straight through. For Inertia GET requests whose asset
// version is stale it replies 409 Conflict + X-Inertia-Location so the client does a
// full reload. Per-request v3 props (Defer/Optional/Merge/...) need no setup here:
// they lazily initialise from the request context the first time one is set.
func (m *inertiaMiddleware) Handle(ctx contractshttp.Context) {
	inertia := facades.Inertia()

	if ctx.Request().Header("X-Inertia") != "" &&
		ctx.Request().Method() == stdhttp.MethodGet &&
		ctx.Request().Header("X-Inertia-Version") != inertia.Version() {
		ctx.Response().Header("X-Inertia-Location", ctx.Request().FullUrl())
		ctx.Request().Abort(stdhttp.StatusConflict)
		return
	}

	// Application-defined shared props (the share() hook).
	if m.opts.Share != nil {
		for key, value := range m.opts.Share(ctx) {
			inertia.Prop(ctx, key, value)
		}
	}

	// Mirror session flash + validation errors into props for both the initial
	// HTML load and X-Inertia visits.
	inertia.ShareSession(ctx)

	ctx.Request().Next()
}

// Handle returns the Inertia middleware driving the protocol, parameterised by
// opts. Applications register a thin wrapper (scaffolded by `inertia:install` as
// app/http/middleware/handle_inertia_requests.go) that calls Handle with their own
// Share callback — keeping the protocol in the package (so upstream fixes flow)
// while the application owns only its shared props.
func Handle(opts Options) contractshttp.Middleware {
	return &inertiaMiddleware{opts: opts}
}

// Inertia returns the protocol middleware with no application share callback. It is
// equivalent to Handle(Options{}) and kept for convenience / backwards compatibility;
// applications that want customisable shared props register the scaffolded
// HandleInertiaRequests middleware instead.
//
//	facades.Route().Middleware(middleware.Inertia()).Group(func(r route.Router) { ... })
func Inertia() contractshttp.Middleware {
	return Handle(Options{})
}
