package server

import "github.com/gorilla/mux"

// NewRouter creates new router instance that will serve DFA grading requests
func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	dfaHandler := newDFAHandler()
	dfaHandler.register(r)

	return r
}
