<p align="center">
  <img src="goravel-inertia.png" alt="Goravel Inertia" width="180" />
</p>

<h1 align="center">Goravel Inertia</h1>

<p align="center">
  Build server-driven single-page apps in <a href="https://github.com/goravel/framework">Goravel</a>
  with <a href="https://inertiajs.com">Inertia.js</a> — no API layer, no client routing.
</p>

<p align="center">
  <a href="https://github.com/goravel/inertia/actions/workflows/ci.yml"><img src="https://github.com/goravel/inertia/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://pkg.go.dev/github.com/goravel/inertia"><img src="https://pkg.go.dev/badge/github.com/goravel/inertia.svg" alt="Go Reference" /></a>
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26" />
  <img src="https://img.shields.io/badge/Inertia.js-v3-8B5CF6" alt="Inertia v3" />
  <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT" />
</p>

---

`goravel/inertia` is the official [Inertia.js](https://inertiajs.com) adapter for
the [Goravel](https://github.com/goravel/framework) framework. It exposes a
Laravel-style API on top of [`petaki/inertia-go`](https://github.com/petaki/inertia-go)
(which implements the Inertia v3 protocol) and wires it into Goravel's HTTP
lifecycle, session, validation and routing.

```go
func (c *HomeController) Index(ctx http.Context) http.Response {
    return facades.Inertia().Render(ctx, "Home", map[string]any{
        "message": "Hello from Goravel + Inertia",
    })
}
```

## Features

- 🧩 **Inertia v3 props** — deferred, optional, always, merge / deep-merge / prepend, scroll, once.
- ⚡ **Vite integration** — HMR dev server (Laravel-style `public/hot`) and hashed production builds.
- 🖥️ **SSR** with an **automatic CSR fallback** when the SSR server is unreachable (no blank pages).
- 💬 **Flash & validation** bridged from Goravel's session into `props.flash` / `props.errors`.
- 🔁 **Inertia-aware redirects** (303 on mutating methods) and external `Location` redirects.
- 🏷️ **Asset versioning** auto-derived from the Vite manifest hash for cache busting.
- 🛠️ **`inertia:install`** artisan command that scaffolds a full demo app — **Vue 3 or React**.

## Requirements

- Go **1.26+**
- Goravel **v1.18+**
- Node 18+ (for the Vite frontend)

> **Versioning.** As an official Goravel package, releases track the framework's
> minor line — Goravel `v1.18.x` → this package `v1.18.x`. Pick the release
> matching your Goravel version.

## Installation

Install the package and register its service provider automatically:

```bash
go run . artisan package:install github.com/goravel/inertia
```

This runs the package's `setup`, which adds `&goravelinertia.ServiceProvider{}`
to `bootstrap/providers.go` (or `config/app.go` on a non-bootstrap setup) for you.

<details>
<summary>Manual registration</summary>

```bash
go get github.com/goravel/inertia
```

```go
import goravelinertia "github.com/goravel/inertia"

var Providers = []foundation.ServiceProvider{
    // ...
    &goravelinertia.ServiceProvider{},
}
```
</details>

Then scaffold the frontend, config, root template and a demo app. Pick a stack
with `--stack` (`vue` is the default):

```bash
go run . artisan inertia:install                 # Vue 3 (default)
go run . artisan inertia:install --stack=react   # React 19
```

This creates `config/inertia.go`, `resources/inertia/app.gohtml`, the JS app
under `resources/js/`, demo pages (Home / Feed / Contact / About) with their
controllers, the `app/http/middleware/handle_inertia_requests.go` middleware (your
shared-props entry point), `vite.config.ts`, `tsconfig.json` and `package.json`.
It also wires `routes/web.go` (session + `HandleInertiaRequests` middleware, demo
routes) and removes the default Goravel welcome view. Pass `--force` to overwrite
existing files.

The Go controllers, middleware, routes and config are identical across stacks —
only the `resources/js/` frontend (`.vue` vs `.tsx`) and the JS toolchain
(`package.json` / `vite.config.ts`) differ.

Finally:

```bash
npm install
npm run dev   # writes public/hot, then in another shell:
go run .
```

Open <http://localhost:3000>.

## Configuration

`config/inertia.go`:

| Key | Env | Default | Description |
|-----|-----|---------|-------------|
| `root_view` | — | `resources/inertia/app.gohtml` | Root Blade-like template. |
| `version` | `INERTIA_VERSION` | manifest hash | Asset version for the version check. |
| `ssr` | `INERTIA_SSR` | `false` | Enable server-side rendering. |
| `ssr_url` | `INERTIA_SSR_URL` | `http://127.0.0.1:13714/render` | SSR Node endpoint. |
| `ssr_timeout` | `INERTIA_SSR_TIMEOUT` | `5` | Seconds before SSR is abandoned for CSR. |
| `flash_keys` | — | `success,error,warning,info,message` | Session keys mirrored into `props.flash`. |
| `vite.public_path` | — | `public` | Public web root served to clients. |
| `vite.build_dir` | — | `build` | Production build dir under `public_path`. |
| `vite.hot_file` | — | `public/hot` | Hot file written by `npm run dev`. |
| `vite.dev_url` | `VITE_DEV_URL` | `` | Dev-server URL (usually set via `public/hot`). |

## Usage

Access the manager via the facade:

```go
import "github.com/goravel/inertia/facades"
```

### Render

```go
return facades.Inertia().Render(ctx, "Users/Index", map[string]any{"users": users})
```

### Shared props

Shared props are sent on **every** Inertia response. This adapter offers a
**hybrid** model — pick by lifecycle:

| Mechanism | Where | Lifecycle | Use for |
|-----------|-------|-----------|---------|
| `share()` middleware | `app/http/middleware/handle_inertia_requests.go` | Per-request (ctx-aware) | Auth user, anything request-dependent. **Recommended default.** |
| `Inertia().ShareFunc(key, fn)` | Provider `Boot` | Per-request (ctx-aware) | Same as above, but registered from a provider/package instead of the app middleware. |
| `Inertia().Share(key, value)` | Provider `Boot` | Static, set once | Constants known at boot (app name, build info). No request access. |

The underlying [`petaki/inertia-go`](https://github.com/petaki/inertia-go) engine
only provides the static `Share(key, value)`. The per-request pieces —
`HandleInertiaRequests` and `ShareFunc` — are added by this adapter so shared
props can read the request (e.g. the authenticated user), the Go analogue of
Laravel's `HandleInertiaRequests::share()`.

**1. Per-request via the middleware (recommended).** `inertia:install` scaffolds
`handle_inertia_requests.go`; its `share()` runs once per request and you own it:

```go
// app/http/middleware/handle_inertia_requests.go
func share(ctx http.Context) map[string]any {
    return map[string]any{
        "appName": facades.Config().GetString("app.name"),
        "auth": map[string]any{
            "user": authUser(ctx),
        },
    }
}
```

Validation errors and session flash are shared by the package automatically — you
only own `share()`. Internally each entry is applied as a per-request prop.

**2. From a provider via the facade.** Useful when a provider or reusable package
must contribute shared props without touching the app middleware. `Share` is
static; `ShareFunc` is resolved per request from the context:

```go
facades.Inertia().Share("appName", "My App")          // static, set once
facades.Inertia().ShareFunc("user", func(ctx http.Context) any { // per-request
    if !ctx.Request().HasSession() {
        return nil
    }
    return ctx.Request().Session().Get("user")
})
```

> Inertia here is session-based: persist the user with
> `ctx.Request().Session().Put("user", user)` at login and
> `ctx.Request().Session().Forget("user")` at logout.

### Inertia v3 props

| Method | Behaviour |
|--------|-----------|
| `Defer(ctx, key, fn, group...)` | Loaded after the initial render (`<Deferred>`). |
| `Optional(ctx, key, fn)` | Only evaluated on partial reloads. |
| `Always(ctx, key, fn)` | Always present, even on partial reloads. |
| `Merge(ctx, key, fn, matchOn...)` | Client shallow-merges (e.g. pagination "load more"). |
| `DeepMerge(ctx, key, fn, matchOn...)` | Client deep-merges. |
| `Prepend(ctx, key, fn, matchOn...)` | Merge, prepending new values. |
| `Scroll(ctx, key, prop)` | Infinite-scroll / pagination metadata. |
| `Once(ctx, key, fn)` | Sent once, then cached client-side. |
| `Prop(ctx, key, value)` | Eager per-request prop. |
| `PreserveFragment(ctx)` | Keep the URL fragment (`#hash`) across the visit. |

```go
func (c *HomeController) Index(ctx http.Context) http.Response {
    facades.Inertia().Defer(ctx, "stats", func() any { return loadStats() })
    return facades.Inertia().Render(ctx, "Home", nil)
}
```

### Flash & validation

The middleware mirrors session flash and validation errors into props
(`props.flash`, `props.errors`) automatically. On a failed validation, flash the
errors and redirect back:

```go
func (c *ContactController) Store(ctx http.Context) http.Response {
    validator, err := ctx.Request().Validate(map[string]any{
        "email": "required|email",
    })
    if err != nil || validator.Fails() {
        facades.Inertia().FlashErrors(ctx, validator.Errors())
        return facades.Inertia().Redirect(ctx, "/contact")
    }

    ctx.Request().Session().Flash("success", "Saved!")
    return facades.Inertia().Redirect(ctx, "/contact")
}
```

On the client: `usePage().props.flash` and `usePage().props.errors` (or `useForm`).

You can also push props imperatively: `Flash(ctx, data)` merges a map into
`props.flash`, and `Error(ctx, key, value)` adds a single validation error.

### Redirects

```go
facades.Inertia().Redirect(ctx, "/dashboard")          // 303 on PUT/PATCH/DELETE, 302 otherwise
facades.Inertia().Location(ctx, "https://example.com") // full-page / external redirect
```

### History

```go
facades.Inertia().ClearHistory(ctx)
facades.Inertia().EncryptHistory(ctx)
```

## Frontend (Vite)

- **Development:** `npm run dev` runs Vite and writes `public/hot`; the backend
  loads assets from the dev server with HMR — no env var needed.
- **Production:** `npm run build` emits hashed assets + a manifest under
  `public/build`; the backend serves them via the manifest. The asset version is
  derived from the manifest hash, so a new build busts the client cache.

The `{{ vite "resources/js/app.ts" }}` template helper picks dev vs prod
automatically (`public/hot` → `VITE_DEV_URL` → manifest).

## SSR

```bash
npm run build:ssr          # client build + SSR bundle (bootstrap/ssr/ssr.js)
npm run ssr                # SSR Node server on :13714 (keeps running, in its own shell)
```

Then, in another shell:

```bash
INERTIA_SSR=true go run .
```

If the SSR server is unreachable, the adapter **falls back to client-side
rendering** instead of returning a blank page (a warning is logged).

## Credits

Built on the shoulders of:

- [Inertia.js](https://github.com/inertiajs/inertia) — the protocol & client (MIT).
- [petaki/inertia-go](https://github.com/petaki/inertia-go) — the Go server engine (MIT).
- [Goravel](https://github.com/goravel/framework) — the Go web framework (MIT).
- [Vue](https://github.com/vuejs/core) & [React](https://github.com/facebook/react) — the scaffolded frontend stacks (MIT).
- [Vite](https://github.com/vitejs/vite) — dev server & production builds (MIT).

## Attribution

The logo features the **Go gopher**, designed by [Renée French](https://reneefrench.blogspot.com)
and licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
It also incorporates the Inertia.js and Goravel marks to identify the projects
this adapter integrates; those marks belong to their respective owners and are
used here only for identification, not endorsement.

## License

Code is released under the [MIT License](LICENSE). Third-party dependencies keep
their own licenses (MIT / BSD / Apache-2.0) — see
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
