package inertia

import (
	"context"

	contractshttp "github.com/goravel/framework/contracts/http"
	petaki "github.com/petaki/inertia-go"

	"github.com/goravel/inertia/contracts"
)

// Ensure the manager satisfies the public contract at compile time.
var _ contracts.Inertia = (*InertiaManager)(nil)

// contextKey namespaces values stored on the Goravel http.Context.
type contextKey string

// ctxKeyProps holds the accumulated context.Context that carries the per-request
// props (deferred, optional, merge, ...). petaki reads these from the
// *http.Request context, so the accumulated context is injected into the request
// just before Render runs.
const ctxKeyProps = contextKey("goravel-inertia.props")

// propsContext returns the accumulated props context for this request, falling
// back to the underlying request context the first time a prop is set.
func (m *InertiaManager) propsContext(ctx contractshttp.Context) context.Context {
	if v := ctx.Value(ctxKeyProps); v != nil {
		if c, ok := v.(context.Context); ok {
			return c
		}
	}

	return m.adapter.Request(ctx).Context()
}

// storePropsContext persists the accumulated props context back onto the
// Goravel http.Context so chained With* calls and Render see the same context.
func (m *InertiaManager) storePropsContext(ctx contractshttp.Context, c context.Context) {
	ctx.WithValue(ctxKeyProps, c)
}

// Prop sets an eagerly-evaluated per-request prop.
func (m *InertiaManager) Prop(ctx contractshttp.Context, key string, value any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithProp(m.propsContext(ctx), key, value))
}

// Defer registers a prop evaluated lazily by the client after the initial load.
func (m *InertiaManager) Defer(ctx contractshttp.Context, key string, fn func() any, group ...string) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithDeferredProp(m.propsContext(ctx), key, fn, group...))
}

// Optional registers a prop only evaluated when explicitly requested (partial reload).
func (m *InertiaManager) Optional(ctx contractshttp.Context, key string, fn func() any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithOptionalProp(m.propsContext(ctx), key, fn))
}

// Always registers a prop included on every response, even on partial reloads.
func (m *InertiaManager) Always(ctx contractshttp.Context, key string, fn func() any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithAlwaysProp(m.propsContext(ctx), key, fn))
}

// Merge registers a prop the client shallow-merges with existing data.
func (m *InertiaManager) Merge(ctx contractshttp.Context, key string, fn func() any, matchOn ...string) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithMergeProp(m.propsContext(ctx), key, fn, matchOn...))
}

// DeepMerge registers a prop the client deep-merges with existing data.
func (m *InertiaManager) DeepMerge(ctx contractshttp.Context, key string, fn func() any, matchOn ...string) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithDeepMergeProp(m.propsContext(ctx), key, fn, matchOn...))
}

// Prepend registers a merge prop whose values are prepended instead of appended.
func (m *InertiaManager) Prepend(ctx contractshttp.Context, key string, fn func() any, matchOn ...string) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithPrependProp(m.propsContext(ctx), key, fn, matchOn...))
}

// Once registers a prop sent only once and then cached by the client.
func (m *InertiaManager) Once(ctx contractshttp.Context, key string, fn func() any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithOnceProp(m.propsContext(ctx), key, fn))
}

// Scroll registers an infinite-scroll/pagination prop.
func (m *InertiaManager) Scroll(ctx contractshttp.Context, key string, prop contracts.ScrollProp) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithScrollProp(m.propsContext(ctx), key, petaki.ScrollPageProp{
		PageName:     prop.PageName,
		CurrentPage:  prop.CurrentPage,
		PreviousPage: prop.PreviousPage,
		NextPage:     prop.NextPage,
		Reset:        prop.Reset,
	}))
}

// Error attaches a single validation error to the response.
func (m *InertiaManager) Error(ctx contractshttp.Context, key string, value any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithErrorProp(m.propsContext(ctx), key, value))
}

// Flash attaches flash data under props.flash, matching the Inertia + Laravel
// convention where flash is a shared prop read via usePage().props.flash. This
// keeps flash consistent with props.errors instead of petaki's top-level page.flash.
func (m *InertiaManager) Flash(ctx contractshttp.Context, data map[string]any) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithProp(m.propsContext(ctx), "flash", data))
}

// ClearHistory instructs the client to clear its history state.
func (m *InertiaManager) ClearHistory(ctx contractshttp.Context) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithClearHistory(m.propsContext(ctx)))
}

// EncryptHistory instructs the client to encrypt its history state.
func (m *InertiaManager) EncryptHistory(ctx contractshttp.Context) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithEncryptHistory(m.propsContext(ctx)))
}

// PreserveFragment keeps the URL fragment across the visit.
func (m *InertiaManager) PreserveFragment(ctx contractshttp.Context) {
	m.storePropsContext(ctx, m.adapter.Inertia().WithPreserveFragment(m.propsContext(ctx)))
}
