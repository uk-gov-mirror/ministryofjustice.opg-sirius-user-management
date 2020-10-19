package server

import (
	"net/http"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type AddMemberClient interface {
}

type addMemberVars struct {
	Path      string
	SiriusURL string
	Search    string
	Users     []sirius.User
	Success   bool
	Errors    sirius.ValidationErrors
}

func addMember(client AddMemberClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		vars := addMemberVars{
			Path:      r.URL.Path,
			SiriusURL: siriusURL,
		}

		return tmpl.ExecuteTemplate(w, "page", vars)
	}
}
