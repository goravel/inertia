package contracts

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// ScrollProp describes an infinite-scroll/pagination prop. Mirrors the petaki
// ScrollPageProp so consumers don't depend on the underlying library.
type ScrollProp struct {
	PageName     string
	CurrentPage  any
	PreviousPage any
	NextPage     any
	Reset        bool
}

type Inertia interface {
	Render(ctx http.Context, component string, props map[string]any) http.Response
	Redirect(ctx http.Context, url string) http.Response
	Location(ctx http.Context, url string) http.Response

	Share(key string, value any)
	ShareFunc(key string, fn func(ctx http.Context) any)

	// Per-request props.
	Prop(ctx http.Context, key string, value any)
	Defer(ctx http.Context, key string, fn func() any, group ...string)
	Optional(ctx http.Context, key string, fn func() any)
	Always(ctx http.Context, key string, fn func() any)
	Merge(ctx http.Context, key string, fn func() any, matchOn ...string)
	DeepMerge(ctx http.Context, key string, fn func() any, matchOn ...string)
	Prepend(ctx http.Context, key string, fn func() any, matchOn ...string)
	Once(ctx http.Context, key string, fn func() any)
	Scroll(ctx http.Context, key string, prop ScrollProp)
	Error(ctx http.Context, key string, value any)
	Flash(ctx http.Context, data map[string]any)
	FlashErrors(ctx http.Context, errors validation.Errors)
	ShareSession(ctx http.Context)
	ClearHistory(ctx http.Context)
	EncryptHistory(ctx http.Context)
	PreserveFragment(ctx http.Context)

	// Version returns the Inertia asset version — the value the client echoes in
	// X-Inertia-Version and the middleware compares; a mismatch answers 409 Conflict +
	// X-Inertia-Location, reloading the page. It defaults to the Vite manifest hash
	// (contracts/vite.Vite.ManifestHash) but is not that hash: the hash is empty while
	// the dev server runs or the manifest is unreadable, which silently disables the
	// check, so a deploy that must bust client caches pins inertia.version.
	Version() string
	URL() string
}
