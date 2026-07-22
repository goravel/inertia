package controllers

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	inertiafacades "github.com/goravel/inertia/facades"
)

type HomeController struct{}

func NewHomeController() *HomeController {
	return &HomeController{}
}

func (r *HomeController) Index(ctx http.Context) http.Response {
	// Deferred prop (Inertia v3): not evaluated on the initial load. The page
	// paints immediately and the client fetches "stats" in a follow-up request.
	inertiafacades.Inertia().Defer(ctx, "stats", func() any {
		time.Sleep(800 * time.Millisecond) // simulate a slow query
		return map[string]any{
			"users":    1280,
			"sessions": 342,
			"uptime":   "99.98%",
		}
	})

	return inertiafacades.Inertia().Render(ctx, "Home", map[string]any{
		"message": "Goravel + Inertia",
	})
}
