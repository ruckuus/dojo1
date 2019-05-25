package controllers

import (
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

	vd.SetSuccessMessage("Successfully created new property")
	p.NewView.Render(w, r, vd)

	//views.RedirectAlert(w, r, "/properties", http.StatusFound, views.Alert{
	//	Level: views.AlertLvlSuccess,
	//	Message: "Successfully creating property",
	//})
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
