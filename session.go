package inertia

import (
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// sessionErrorsKey is the session key under which validation errors are flashed
// so they survive the redirect back to the form, matching the Inertia convention.
const sessionErrorsKey = "errors"

// ShareSession mirrors flash messages and validation errors stored in the session
// into the Inertia props for this request. It runs in the middleware before the
// handler, so both the initial HTML load and X-Inertia visits pick them up:
// props.flash for the configured flash keys, props.errors for validation errors.
func (m *InertiaManager) ShareSession(ctx contractshttp.Context) {
	if !ctx.Request().HasSession() {
		return
	}

	s := ctx.Request().Session()

	flash := make(map[string]any)
	for _, key := range m.flashKeys {
		if s.Has(key) {
			flash[key] = s.Get(key)
		}
	}
	if len(flash) > 0 {
		m.Flash(ctx, flash)
	}

	for field, msg := range readErrors(s.Get(sessionErrorsKey)) {
		m.Error(ctx, field, msg)
	}
}

// FlashErrors flattens Goravel validation errors to one message per field and
// flashes them to the session. Call it before redirecting back from a failed
// validation; ShareSession then exposes them as props.errors on the next request.
func (m *InertiaManager) FlashErrors(ctx contractshttp.Context, errors validation.Errors) {
	if errors == nil || !ctx.Request().HasSession() {
		return
	}

	flat := make(map[string]any, len(errors.All()))
	for field := range errors.All() {
		flat[field] = errors.One(field)
	}

	ctx.Request().Session().Flash(sessionErrorsKey, flat)
}

// readErrors normalises the session-stored errors value, which may round-trip as
// either map[string]any or map[string]string depending on the session driver.
func readErrors(v any) map[string]any {
	switch errs := v.(type) {
	case map[string]any:
		return errs
	case map[string]string:
		out := make(map[string]any, len(errs))
		for k, val := range errs {
			out[k] = val
		}
		return out
	default:
		return nil
	}
}
