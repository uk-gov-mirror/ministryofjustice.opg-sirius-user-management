package server

import (
	"context"
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type AddUserClient interface {
	AddUser(ctx context.Context, cookies []*http.Cookie, email, firstname, surname, organisation string, roles []string) error
	MyDetails(context.Context, []*http.Cookie) (sirius.MyDetails, error)
}

type addUserVars struct {
	Path      string
	SiriusURL string
	Errors    sirius.ValidationErrors
}

func addUser(logger *log.Logger, client AddUserClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		myDetails, err := client.MyDetails(r.Context(), r.Cookies())
		if err != nil {
			return err
		}

		permitted := false
		for _, role := range myDetails.Roles {
			if role == "System Admin" {
				permitted = true
			}
		}

		if !permitted {
			return StatusError(http.StatusForbidden)
		}

		vars := addUserVars{
			Path:      r.URL.Path,
			SiriusURL: siriusURL,
		}

		switch r.Method {
		case http.MethodGet:
			return tmpl.ExecuteTemplate(w, "page", vars)

		case http.MethodPost:
			var (
				email        = r.PostFormValue("email")
				firstname    = r.PostFormValue("firstname")
				surname      = r.PostFormValue("surname")
				organisation = r.PostFormValue("organisation")
				roles        = r.PostForm["roles"]
			)

			err := client.AddUser(r.Context(), r.Cookies(), email, firstname, surname, organisation, roles)

			if verr, ok := err.(sirius.ValidationError); ok {
				vars.Errors = verr.Errors

				w.WriteHeader(http.StatusBadRequest)
				return tmpl.ExecuteTemplate(w, "page", vars)
			}

			if err != nil {
				return err
			}

			return RedirectError("/users")

		default:
			return StatusError(http.StatusMethodNotAllowed)
		}
	}
}