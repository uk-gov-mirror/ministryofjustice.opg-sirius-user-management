package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type AddMemberClient interface {
	SearchUsers(context.Context, []*http.Cookie, string) ([]sirius.User, error)
	Team(ctx context.Context, cookies []*http.Cookie, id int) (sirius.Team, error)
}

type addMemberVars struct {
	Path         string
	SiriusURL    string
	Search       string
	Users        []sirius.User
	Success      bool
	Errors       sirius.ValidationErrors
	Team         sirius.Team
	TeamUsersMap map[int]bool
}

func addMember(client AddMemberClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {

		search := r.FormValue("search")

		vars := addMemberVars{
			Path:      r.URL.Path,
			SiriusURL: siriusURL,
			Search:    search,
		}

		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/teams/add-member/"))
		if err != nil {
			return StatusError(http.StatusNotFound)
		}

		team, err := client.Team(r.Context(), r.Cookies(), id)
		if err != nil {
			return err
		}
		vars.Team = team

		if len(search) >= 3 {
			users, err := client.SearchUsers(r.Context(), r.Cookies(), search)
			if err != nil {
				return err
			}
			vars.Users = users
		} else if search != "" {
			vars.Errors = sirius.ValidationErrors{
				"search": {
					"": "Search term must be at least three characters",
				},
			}
		}

		vars.TeamUsersMap = map[int]bool{}

		for _, u := range vars.Users {
			vars.TeamUsersMap[u.ID] = false
		}
		for _, u := range vars.Team.Members {
			vars.TeamUsersMap[u.ID] = true
		}

		return tmpl.ExecuteTemplate(w, "page", vars)
	}
}
