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

	Version() string
	URL() string
}
