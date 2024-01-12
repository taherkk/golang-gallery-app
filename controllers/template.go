package controllers

import "net/http"

// if we are to use a different template engine nothing in our controllers
// change
type Template interface {
	Execute(w http.ResponseWriter, r *http.Request, data interface{})
}
