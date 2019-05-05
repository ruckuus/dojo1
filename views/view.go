package views

import (
	"bytes"
	"errors"
	"github.com/gorilla/csrf"
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/models"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

var (
	TemplateDir = "views/"
	LayoutDir   = TemplateDir + "layouts/"
	TemplateExt = ".gohtml"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if problem persists."
)

type Data struct {
	Alert *Alert
	Yield interface{}
	User  *models.User
}

type Alert struct {
	Level   string
	Message string
}

type PublicError interface {
	error
	Public() string
}

func (d *Data) SetAlert(err error) {
	var msg string
	if publicError, ok := err.(PublicError); ok {
		msg = publicError.Public()
	} else {
		log.Println(err)
		msg = err.Error()
	}

	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

func (d *Data) SetSuccessMessage(message string) {
	d.Alert = &Alert{
		Level:   AlertLvlSuccess,
		Message: message,
	}
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}

	return files
}

func NewView(layout string, files ...string) *View {

	for i, f := range files {
		files[i] = TemplateDir + f + TemplateExt
	}

	files = append(files, layoutFiles()...)
	t, err := template.
		New("").Funcs(template.FuncMap{
		// csrfField stub, this is required in compile time.
		// it will eventually be replaced in Render() method
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not implemented yet.")
		},
	}).
		ParseFiles(files...)

	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yield: data,
		}
	}
	vd.User = context.User(r.Context())
	var buf bytes.Buffer

	// actual implementation of csrfField
	// create csrfField from the current http request
	csrfField := csrf.TemplateField(r)
	// this v.Template.Funcs return template
	tpl := v.Template.Funcs(template.FuncMap{
		// we don't need to worry about error here
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	// Execute the template, now it contains csrf field
	err := tpl.ExecuteTemplate(&buf, v.Layout, vd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}
