package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/views"
	"net/http"
	"strconv"
)

const (
	IndexProperties = "index_properties"
)

// Properties defines the Properties controller
// It ties its models and views in one place
type Properties struct {
	NewView   *views.View
	IndexView *views.View
	ShowView  *views.View
	EditView  *views.View
	ps        models.PropertyService
	r         *mux.Router
}

// PropertyForm defines schema for form input
type PropertyForm struct {
	Name       string `schema:"name"`
	Address    string `schema:"address"`
	PostalCode string `schema:"postal_code"`
}

// NewProperties returns new Properties object,
// it instantiates all the necessary elements to
// be used by every controller methods
func NewProperties(services models.PropertyService, r *mux.Router) *Properties {
	return &Properties{
		NewView:   views.NewView("bootstrap", "properties/new"),
		IndexView: views.NewView("bootstrap", "properties/index"),
		ShowView:  views.NewView("bootstrap", "properties/show"),
		EditView:  views.NewView("bootstrap", "properties/edit"),
		ps:        services,
		r:         r,
	}
}

// New renders the view for GET /properties/new
func (p *Properties) New(w http.ResponseWriter, r *http.Request) {
	p.NewView.Render(w, r, nil)
}

// Create handles property creation POST /properties
func (p *Properties) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form PropertyForm

	user := context.User(r.Context())

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		p.NewView.Render(w, r, vd)
		return
	}

	property := models.Property{
		UserID:     user.ID,
		Name:       form.Name,
		Address:    form.Address,
		PostalCode: form.PostalCode,
	}

	if err := p.ps.Create(&property); err != nil {
		vd.SetAlert(err)
		p.NewView.Render(w, r, vd)
		return
	}

	views.RedirectAlert(w, r, fmt.Sprintf("/properties/%d", property.ID), http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Successfully created property",
	})
}

func (p *Properties) Index(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	user := context.User(r.Context())
	properties, err := p.ps.ByUserID(user.ID)
	if err != nil {
		vd.SetAlert(err)
		p.IndexView.Render(w, r, vd)
		return
	}
	vd.Yield = properties
	p.IndexView.Render(w, r, vd)
}

// propertyByID helper to fetch property by id
func (p *Properties) propertyByID(w http.ResponseWriter, r *http.Request) (*models.Property, error) {
	vars := mux.Vars(r)

	paramID := vars["id"]

	id, err := strconv.Atoi(paramID)
	if err != nil {
		return nil, err
	}

	property, err := p.ps.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Property not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return nil, err
	}
	return property, nil
}

// Show single property
func (p *Properties) Show(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	property, err := p.propertyByID(w, r)
	if err != nil {
		return
	}
	vd.Yield = property
	p.ShowView.Render(w, r, vd)
}

// Edit handles GET /properties/:id/edit
func (p *Properties) Edit(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	property, err := p.propertyByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())

	if property.UserID != user.ID {
		http.Error(w, "Error updating property, user mismatch", http.StatusForbidden)
		return
	}

	vd.Yield = property

	p.EditView.Render(w, r, vd)
}

// Update handles POST /properties/:id/update
func (p *Properties) Update(w http.ResponseWriter, r *http.Request) {
	property, err := p.propertyByID(w, r)
	if err != nil {
		http.Error(w, "error updating property", http.StatusBadRequest)
		return
	}

	user := context.User(r.Context())

	if property.UserID != user.ID {
		http.Error(w, "error updating property, user mismatch", http.StatusForbidden)
		return
	}

	var vd views.Data
	vd.Yield = property

	var form PropertyForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		p.EditView.Render(w, r, vd)
		return
	}

	property.Name = form.Name
	property.Address = form.Address
	property.PostalCode = form.PostalCode

	err = p.ps.Update(property)
	if err != nil {
		vd.SetAlert(err)
		p.EditView.Render(w, r, vd)
		return
	}

	views.RedirectAlert(w, r, fmt.Sprintf("/properties/%d", property.ID), http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Property updated successfully.",
	})
}

// Delete handles POST /properties/:id/delete
func (p *Properties) Delete(w http.ResponseWriter, r *http.Request) {
	property, err := p.propertyByID(w, r)
	if err != nil {
		http.Error(w, "Error deleting property", http.StatusBadRequest)
		return
	}

	user := context.User(r.Context())
	if property.UserID != user.ID {
		http.Error(w, "Error deleting property, user mismatch", http.StatusForbidden)
		return
	}

	var vd views.Data
	err = p.ps.Delete(property.ID)

	if err != nil {
		vd.SetAlert(err)
		vd.Yield = property
		p.EditView.Render(w, r, vd)
		return
	}

	// FIXME: alert is not shown in /properties page
	views.RedirectAlert(w, r, "/properties", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Successfully deleted property.",
	})
}
