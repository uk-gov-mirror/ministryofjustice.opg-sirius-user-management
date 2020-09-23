package server

import (
	"io"
	"log"
	"net/http"
)

type Client interface {
	MyDetailsClient
}

type Templates interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

func New(logger *log.Logger, client Client, templates Templates, siriusURL, webDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.RedirectHandler("/my-details", http.StatusFound))
	mux.Handle("/my-details", myDetails(logger, client, templates, siriusURL))

	static := http.FileServer(http.Dir(webDir + "/static"))
	mux.Handle("/assets/", static)
	mux.Handle("/javascript/", static)
	mux.Handle("/stylesheets/", static)

	return mux
}
