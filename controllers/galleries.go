package controllers

import (
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/views"
	"net/http"
	"strconv"
)

type Galleries struct {
	NewView  *views.View
	ShowView *views.View
	gs       models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

func NewGalleries(services models.GalleryService) *Galleries {
	return &Galleries{
		NewView:  views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		gs:       services,
	}
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, nil)
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	user := context.User(r.Context())

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		UserID: user.ID, // Hardcoded for now
		Title:  form.Title,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	vd.SetSuccessMessage("Successfully created gallery.")
	g.NewView.Render(w, vd)
	return
}

func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	vars := mux.Vars(r)

	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		vd.SetAlert(err)
		g.ShowView.Render(w, vd)
		return
	}

	_ = id

	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.SetAlert(err)
		default:
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
		}
		g.ShowView.Render(w, vd)
		return
	}
	vd.Yield = gallery

	g.ShowView.Render(w, vd)

}
