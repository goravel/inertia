package inertia

// response implements goravel contracts/http.Response. The actual write to the
// underlying http.ResponseWriter is deferred until Goravel calls Render() during
// the request lifecycle, so any errors surface to the framework instead of being
// swallowed in the controller.
type response struct {
	render func() error
}

func newResponse(render func() error) *response {
	return &response{render: render}
}

func (r *response) Render() error {
	return r.render()
}
