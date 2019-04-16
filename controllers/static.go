package controllers

import "github.com/ruckuus/dojo1/views"

type Static struct {
	Home        *views.View
	Contact *views.View
	FAQ     *views.View
}

func NewStatic() *Static  {
	return &Static{
		Home:        views.NewView("bootstrap", "views/static/home.gohtml"),
		Contact: views.NewView("bootstrap", "views/static/contact.gohtml"),
		FAQ:     views.NewView("bootstrap", "views/static/faq.gohtml"),
	}
}
