package controllers

import (
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

type Galleries struct {
	NewView *views.View
}

func NewGalleries() *Galleries {
	return &Galleries{
		NewView: views.NewView("bootstrap", "galleries/new"),
	}
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request)  {
	g.NewView.Render(w, nil)
}