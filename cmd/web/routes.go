package main

import "net/http"

func (a *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/", a.parseLoad)
	return router
}
