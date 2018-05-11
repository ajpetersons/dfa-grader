package server

import "github.com/gorilla/mux"

func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	dfaHandler := NewDFAHandler()
	dfaHandler.Register(r)

	return r
}
