package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type Client interface {
	AddUserClient
	MyDetailsClient
	EditMyDetailsClient
	ChangePasswordClient
	ListUsersClient
}

type Template interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

func New(logger *log.Logger, client Client, templates map[string]*template.Template, prefix, siriusURL, siriusPublicURL string, webDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.RedirectHandler(prefix+"/my-details", http.StatusFound))
	mux.Handle("/health-check", healthCheck())
	mux.Handle("/users",
		errorHandler("listUsers", logger, templates["error.gotmpl"], prefix, siriusPublicURL,
			listUsers(logger, client, templates["users.gotmpl"], siriusURL)))
	mux.Handle("/my-details",
		errorHandler("myDetails", logger, templates["error.gotmpl"], prefix, siriusPublicURL,
			myDetails(logger, client, templates["my-details.gotmpl"], siriusURL)))
	mux.Handle("/my-details/edit",
		errorHandler("editMyDetails", logger, templates["error.gotmpl"], prefix, siriusPublicURL,
			editMyDetails(logger, client, templates["edit-my-details.gotmpl"], siriusURL)))
	mux.Handle("/change-password",
		errorHandler("changePassword", logger, templates["error.gotmpl"], prefix, siriusPublicURL,
			changePassword(logger, client, templates["change-password.gotmpl"], siriusURL)))
	mux.Handle("/add-user",
		errorHandler("addUser", logger, templates["error.gotmpl"], prefix, siriusPublicURL,
			addUser(logger, client, templates["add-user.gotmpl"], siriusURL)))

	static := http.FileServer(http.Dir(webDir + "/static"))
	mux.Handle("/assets/", static)
	mux.Handle("/javascript/", static)
	mux.Handle("/stylesheets/", static)

	return http.StripPrefix(prefix, mux)
}

type RedirectError string

func (e RedirectError) Error() string {
	return "redirect to " + string(e)
}

func (e RedirectError) To() string {
	return string(e)
}

type StatusError int

func (e StatusError) Error() string {
	code := e.Code()

	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func (e StatusError) Code() int {
	return int(e)
}

type Handler func(w http.ResponseWriter, r *http.Request) error

type errorVars struct {
	SiriusURL string
	Path      string

	Code  int
	Error string
}

func errorHandler(name string, logger *log.Logger, tmplError Template, prefix, siriusURL string, next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			if err == sirius.ErrUnauthorized {
				http.Redirect(w, r, siriusURL+"/auth", http.StatusFound)
				return
			}

			if redirect, ok := err.(RedirectError); ok {
				http.Redirect(w, r, prefix+redirect.To(), http.StatusFound)
				return
			}

			code := http.StatusInternalServerError

			if status, ok := err.(StatusError); ok {
				if status.Code() == http.StatusForbidden || status.Code() == http.StatusNotFound {
					code = status.Code()
				}
			}

			logger.Printf("%s: %v\n", name, err)

			w.WriteHeader(code)
			err = tmplError.ExecuteTemplate(w, "page", errorVars{
				SiriusURL: siriusURL,
				Path:      "",
				Code:      code,
				Error:     err.Error(),
			})

			if err != nil {
				logger.Printf("%s: %v\n", name, err)
				http.Error(w, "Could not generate error template", http.StatusInternalServerError)
			}
		}
	})
}
