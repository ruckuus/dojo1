package controllers

import (
	"github.com/gorilla/schema"
	"net/http"
	"net/url"
)

func parseForm(r *http.Request, dst interface{}) error {

	if err := r.ParseForm(); err != nil {
		return err
	}

	return parseValues(r.PostForm, dst)

}

func parseValues(values url.Values, dst interface{}) error {

	decoder := schema.NewDecoder()

	// This is to tell gorilla/schema to ignore unknown key in the request
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(dst, values)

	if err != nil {
		return err
	}

	return nil
}

func parseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	return parseValues(r.Form, dst)
}
