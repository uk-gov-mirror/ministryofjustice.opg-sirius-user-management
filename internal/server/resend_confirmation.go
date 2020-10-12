package server

import (
	"context"
	"net/http"
)

type ResendConfirmationClient interface {
	ResendConfirmation(context.Context, []*http.Cookie, string) error
}

type resendConfirmationVars struct {
	Path      string
	SiriusURL string
	ID        string
	Email     string
}

func resendConfirmation(client ResendConfirmationClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return RedirectError("/users")

		case http.MethodPost:
			vars := resendConfirmationVars{
				Path:      r.URL.Path,
				SiriusURL: siriusURL,
				ID:        r.PostFormValue("id"),
				Email:     r.PostFormValue("email"),
			}

			err := client.ResendConfirmation(r.Context(), r.Cookies(), vars.Email)
			if err != nil {
				return err
			}

			return tmpl.ExecuteTemplate(w, "page", vars)

		default:
			return StatusError(http.StatusMethodNotAllowed)
		}
	}
}