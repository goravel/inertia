package controllers

import (
	"github.com/goravel/framework/contracts/http"

	inertiafacades "github.com/goravel/inertia/facades"
)

type AboutController struct{}

func NewAboutController() *AboutController {
	return &AboutController{}
}

func (r *AboutController) Index(ctx http.Context) http.Response {
	return inertiafacades.Inertia().Render(ctx, "About", map[string]any{
		"title":   "About",
		"content": "An example app built with Goravel and Inertia.js.",
	})
}
