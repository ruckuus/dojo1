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
	NewView   *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	r         *mux.Router
	is        models.ImageService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

const (
	ShowGallery     = "show_gallery"
	EditGallery     = "edit_gallery"
	UpdateGallery   = "update_gallery"
	DeleteGallery   = "delete_gallery"
	IndexGalleries  = "index_galleries"
	maxMultipartMem = 1 << 20 // 1 MB
)

func NewGalleries(services models.GalleryService, r *mux.Router, is models.ImageService) *Galleries {
	return &Galleries{
		NewView:   views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        services,
		r:         r,
		is:        is,
	}
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, r, nil)
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	user := context.User(r.Context())

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, r, vd)
		return
	}

	gallery := models.Gallery{
		UserID: user.ID,
		Title:  form.Title,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, r, vd)
		return
	}
	url, err := g.r.Get(EditGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	vd.Yield = gallery

	g.ShowView.Render(w, r, vd)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {

	vars := mux.Vars(r)

	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}

	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
		}
		return nil, err
	}
	images, err := g.is.ByGalleryID(gallery.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	gallery.Images = images
	return gallery, nil
}

// /galleries/:ID/update
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	// Find the user
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to access this gallery.", http.StatusForbidden)
		return
	}

	vd.Yield = gallery

	g.EditView.Render(w, r, vd)
}

func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		http.Error(w, "Error updating gallery", http.StatusBadRequest)
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Error updating gallery, user mismatch", http.StatusForbidden)
		return
	}

	var vd views.Data
	vd.Yield = gallery

	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Title = form.Title

	err = g.gs.Update(gallery)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	vd.SetSuccessMessage("Gallery updated successfully")

	g.EditView.Render(w, r, vd)
}

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)

	if err != nil {
		http.Error(w, "Error deleting gallery", http.StatusBadRequest)
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Error deleting gallery, user mismatch", http.StatusForbidden)
		return
	}

	var vd views.Data
	err = g.gs.Delete(gallery.ID)

	if err != nil {
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		vd.SetAlert(err)
		g.IndexView.Render(w, r, vd)
		return
	}

	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
	return
}

func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()

		err = g.is.Create(gallery.ID, file, f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		vd.SetSuccessMessage("Successfully upload images!")
		g.EditView.Render(w, r, vd)
		return
	}
}
