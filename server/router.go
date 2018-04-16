package server

import "github.com/gorilla/mux"

func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	return r
}
