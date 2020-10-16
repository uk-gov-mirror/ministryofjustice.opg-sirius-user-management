package sirius

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

func TestTeams(t *testing.T) {
	pact := &dsl.Pact{
		Consumer:          "sirius-user-management",
		Provider:          "sirius",
		Host:              "localhost",
		PactFileWriteMode: "merge",
		LogDir:            "../../logs",
		PactDir:           "../../pacts",
	}
	defer pact.Teardown()

	testCases := []struct {
		name             string
		setup            func()
		cookies          []*http.Cookie
		expectedResponse []Team
		expectedError    error
	}{
		{
			name: "OK",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request for teams").
					WithRequest(dsl.Request{
						Method: http.MethodGet,
						Path:   dsl.String("/api/v1/teams"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status:  http.StatusOK,
						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
						Body: dsl.EachLike(map[string]interface{}{
							"id":          dsl.Like(123),
							"displayName": dsl.Like("Cool Team"),
							"members": dsl.EachLike(map[string]interface{}{
								"displayName": dsl.Like("John"),
								"email":       dsl.Like("john@opgtest.com"),
							}, 1),
							"teamType": dsl.Like(map[string]interface{}{
								"handle": "ALLOCATIONS",
								"label":  "Allocations",
							}),
						}, 1),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			expectedResponse: []Team{
				{
					ID:          123,
					DisplayName: "Cool Team",
					Members: []TeamMember{
						{
							DisplayName: "John",
							Email:       "john@opgtest.com",
						},
					},
					Type:      "ALLOCATIONS",
					TypeLabel: "Supervision — Allocations",
				},
			},
		},

		{
			name: "OKWithLpaTeams",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists and teams have no type").
					UponReceiving("A request for teams").
					WithRequest(dsl.Request{
						Method: http.MethodGet,
						Path:   dsl.String("/api/v1/teams"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status:  http.StatusOK,
						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
						Body: dsl.EachLike(map[string]interface{}{
							"id":          dsl.Like(123),
							"displayName": dsl.Like("Cool Team"),
							"members": dsl.EachLike(map[string]interface{}{
								"displayName": dsl.Like("John"),
								"email":       dsl.Like("john@opgtest.com"),
							}, 1),
						}, 1),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			expectedResponse: []Team{
				{
					ID:          123,
					DisplayName: "Cool Team",
					Members: []TeamMember{
						{
							DisplayName: "John",
							Email:       "john@opgtest.com",
						},
					},
					Type:      "",
					TypeLabel: "LPA",
				},
			},
		},

		{
			name: "Unauthorized",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request for teams without cookies").
					WithRequest(dsl.Request{
						Method: http.MethodGet,
						Path:   dsl.String("/api/v1/teams"),
						Headers: dsl.MapMatcher{
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusUnauthorized,
					})
			},
			expectedError: ErrUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			assert.Nil(t, pact.Verify(func() error {
				client, _ := NewClient(http.DefaultClient, fmt.Sprintf("http://localhost:%d", pact.Server.Port))

				users, err := client.Teams(context.Background(), tc.cookies)
				assert.Equal(t, tc.expectedResponse, users)
				assert.Equal(t, tc.expectedError, err)
				return nil
			}))
		})
	}
}
