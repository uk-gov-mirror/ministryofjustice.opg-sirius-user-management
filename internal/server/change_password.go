package server

import (
	"net/http"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type ChangePasswordClient interface {
	ChangePassword(ctx sirius.Context, currentPassword, newPassword, newPasswordConfirm string) error
}

type changePasswordVars struct {
	Path      string
	SiriusURL string
	XSRFToken string
	Success   bool
	Errors    sirius.ValidationErrors
}

func changePassword(client ChangePasswordClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := getContext(r)

		vars := changePasswordVars{
			Path:      r.URL.Path,
			SiriusURL: siriusURL,
			XSRFToken: ctx.XSRFToken,
		}

		switch r.Method {
		case http.MethodGet:
			return tmpl.ExecuteTemplate(w, "page", vars)

		case http.MethodPost:
			var (
				currentPassword = r.PostFormValue("currentpassword")
				password1       = r.PostFormValue("password1")
				password2       = r.PostFormValue("password2")
			)

			err := client.ChangePassword(ctx, currentPassword, password1, password2)

			if err == sirius.ErrUnauthorized {
				return err
			}

			if _, ok := err.(sirius.ClientError); ok {
				vars.Errors = sirius.ValidationErrors{
					"currentpassword": {
						"": err.Error(),
					},
				}
				w.WriteHeader(http.StatusBadRequest)
				return tmpl.ExecuteTemplate(w, "page", vars)
			}

			if err != nil {
				return err
			}

			vars.Success = true
			return tmpl.ExecuteTemplate(w, "page", vars)

		default:
			return StatusError(http.StatusMethodNotAllowed)
		}
	}
}
