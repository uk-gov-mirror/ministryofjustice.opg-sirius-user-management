package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type Logger interface {
	Request(*http.Request, error)
}

type Client interface {
	AddTeamClient
	AddUserClient
	AllowRolesClient
	ChangePasswordClient
	DeleteUserClient
	EditMyDetailsClient
	EditTeamClient
	EditUserClient
	ListTeamsClient
	ListUsersClient
	MyDetailsClient
	ResendConfirmationClient
	UnlockUserClient
	ViewTeamClient
}

type Template interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

func New(logger Logger, client Client, templates map[string]*template.Template, prefix, siriusURL, siriusPublicURL, webDir string) http.Handler {
	wrap := errorHandler(logger, templates["error.gotmpl"], prefix, siriusPublicURL)
	systemAdminOnly := allowRoles(client, "System Admin")

	deleteUserEnabled := siriusPublicURL != "https://live.sirius-opg.uk"

	mux := http.NewServeMux()
	mux.Handle("/", http.RedirectHandler(prefix+"/my-details", http.StatusFound))
	mux.Handle("/health-check", healthCheck())

	mux.Handle("/users",
		wrap(
			systemAdminOnly(
				listUsers(client, templates["users.gotmpl"]))))

	mux.Handle("/teams",
		wrap(
			allowRoles(client, "System Admin", "Manager")(
				listTeams(client, templates["teams.gotmpl"]))))

	mux.Handle("/teams/",
		wrap(
			allowRoles(client, "System Admin", "Manager")(
				viewTeam(client, templates["team.gotmpl"]))))

	mux.Handle("/teams/add",
		wrap(
			systemAdminOnly(
				addTeam(client, templates["team-add.gotmpl"]))))

	mux.Handle("/teams/edit/",
		wrap(
			allowRoles(client, "System Admin", "Manager")(
				editTeam(client, templates["team-edit.gotmpl"]))))

	mux.Handle("/teams/add-member/",
		wrap(
			allowRoles(client, "System Admin", "Manager")(
				addTeamMember(client, templates["team-add-member.gotmpl"]))))

	mux.Handle("/teams/remove-member/",
		wrap(
			allowRoles(client, "System Admin", "Manager")(
				removeTeamMember(client, templates["team-remove-member.gotmpl"]))))

	mux.Handle("/my-details",
		wrap(
			myDetails(client, templates["my-details.gotmpl"])))

	mux.Handle("/my-details/edit",
		wrap(
			editMyDetails(client, templates["edit-my-details.gotmpl"])))

	mux.Handle("/change-password",
		wrap(
			changePassword(client, templates["change-password.gotmpl"])))

	mux.Handle("/add-user",
		wrap(
			systemAdminOnly(
				addUser(client, templates["add-user.gotmpl"]))))

	mux.Handle("/edit-user/",
		wrap(
			systemAdminOnly(
				editUser(client, templates["edit-user.gotmpl"], deleteUserEnabled))))

	mux.Handle("/unlock-user/",
		wrap(
			systemAdminOnly(
				unlockUser(client, templates["unlock-user.gotmpl"]))))

	mux.Handle("/delete-user/",
		wrap(
			systemAdminOnly(
				deleteUser(client, templates["delete-user.gotmpl"], deleteUserEnabled))))

	mux.Handle("/resend-confirmation",
		wrap(
			systemAdminOnly(
				resendConfirmation(client, templates["resend-confirmation.gotmpl"]))))

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

func errorHandler(logger Logger, tmplError Template, prefix, siriusURL string) func(next Handler) http.Handler {
	return func(next Handler) http.Handler {
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

				logger.Request(r, err)

				code := http.StatusInternalServerError
				if status, ok := err.(StatusError); ok {
					if status.Code() == http.StatusForbidden || status.Code() == http.StatusNotFound {
						code = status.Code()
					}
				}

				w.WriteHeader(code)
				err = tmplError.ExecuteTemplate(w, "page", errorVars{
					SiriusURL: siriusURL,
					Path:      "",
					Code:      code,
					Error:     err.Error(),
				})

				if err != nil {
					logger.Request(r, err)
					http.Error(w, "Could not generate error template", http.StatusInternalServerError)
				}
			}
		})
	}
}

func getContext(r *http.Request) sirius.Context {
	token := ""

	if r.Method == http.MethodGet {
		if cookie, err := r.Cookie("XSRF-TOKEN"); err == nil {
			token, _ = url.QueryUnescape(cookie.Value)
		}
	} else {
		token = r.FormValue("xsrfToken")
	}

	return sirius.Context{
		Context:   r.Context(),
		Cookies:   r.Cookies(),
		XSRFToken: token,
	}
}
