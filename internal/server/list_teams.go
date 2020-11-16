package server

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
)

type ListTeamsClient interface {
	Teams(sirius.Context) ([]sirius.Team, error)
}

type listTeamsVars struct {
	Path      string
	SiriusURL string
	XSRFToken string
	Search    string
	Teams     []sirius.Team
}

func listTeams(client ListTeamsClient, tmpl Template, siriusURL string) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return StatusError(http.StatusMethodNotAllowed)
		}

		ctx := getContext(r)

		teams, err := client.Teams(ctx)
		if err != nil {
			return err
		}

		search := r.FormValue("search")
		if search != "" {
			searchLower := strings.ToLower(search)

			var matchingTeams []sirius.Team
			for _, t := range teams {
				if strings.Contains(strings.ToLower(t.DisplayName), searchLower) {
					matchingTeams = append(matchingTeams, t)
				}
			}

			teams = matchingTeams
		}

		vars := listTeamsVars{
			Path:      r.URL.Path,
			SiriusURL: siriusURL,
			XSRFToken: ctx.XSRFToken,
			Search:    search,
			Teams:     teams,
		}

		return tmpl.ExecuteTemplate(w, "page", vars)
	}
}
