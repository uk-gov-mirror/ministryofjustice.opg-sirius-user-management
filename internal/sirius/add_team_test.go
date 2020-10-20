package sirius

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

func TestAddTeam(t *testing.T) {
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
		scenario      string
		setup         func()
		cookies       []*http.Cookie
		name          string
		teamType      string
		phone         string
		email         string
		expectedID    int
		expectedError func(int) error
	}{
		{
			scenario: "Created",
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to add a new team").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/api/team"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
							"Content-Type":        dsl.String("application/x-www-form-urlencoded"),
						},
						Body: "email=a&name=b&phone=c&type=&teamType=",
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusCreated,
						Body: dsl.Like(map[string]interface{}{
							"data": dsl.Like(map[string]interface{}{
								"id": dsl.Like(123),
							}),
						}),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			email:         "a",
			name:          "b",
			phone:         "c",
			teamType:      "",
			expectedID:    123,
			expectedError: func(port int) error { return nil },
		},

		{
			scenario: "CreatedSupervision",
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to add a new supervision team").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/api/team"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
							"Content-Type":        dsl.String("application/x-www-form-urlencoded"),
						},
						Body: "email=a&name=b&phone=c&type=&teamType=&teamType[handle]=WHAT",
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusCreated,
						Body: dsl.Like(map[string]interface{}{
							"data": dsl.Like(map[string]interface{}{
								"id": dsl.Like(123),
							}),
						}),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			email:         "a",
			name:          "b",
			phone:         "c",
			teamType:      "WHAT",
			expectedID:    123,
			expectedError: func(port int) error { return nil },
		},

		{
			scenario: "Unauthorized",
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to add a new team without cookies").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/api/team"),
						Headers: dsl.MapMatcher{
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusUnauthorized,
					})
			},
			expectedError: func(_ int) error { return ErrUnauthorized },
		},

		{
			scenario: "Errors",
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to add a new team errors").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/api/team"),
						Headers: dsl.MapMatcher{
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusBadRequest,
						Body: dsl.Like(map[string]interface{}{
							"data": dsl.Like(map[string]interface{}{
								"errorMessages": dsl.Like(map[string]interface{}{
									"email": dsl.Like(map[string]interface{}{
										"stringLengthTooLong": "The input is more than 255 characters long",
									}),
								}),
							}),
						}),
					})
			},
			expectedError: func(_ int) error {
				return ValidationError{
					Errors: ValidationErrors{
						"email": {
							"stringLengthTooLong": "The input is more than 255 characters long",
						},
					},
				}
			},
		},

		{
			scenario: "BadRequest",
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to add a new team with bad request").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/api/team"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
							"Content-Type":        dsl.String("application/x-www-form-urlencoded"),
						},
						Body: "email=a&name=b&phone=c&teamType=&type=",
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusBadRequest,
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			email: "a",
			name:  "b",
			phone: "c",
			expectedError: func(port int) error {
				return StatusError{
					Code:   http.StatusBadRequest,
					URL:    fmt.Sprintf("http://localhost:%d/api/team", port),
					Method: http.MethodPost,
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.setup()

			assert.Nil(t, pact.Verify(func() error {
				client, _ := NewClient(http.DefaultClient, fmt.Sprintf("http://localhost:%d", pact.Server.Port))

				id, err := client.AddTeam(context.Background(), tc.cookies, tc.name, tc.teamType, tc.phone, tc.email)
				assert.Equal(t, tc.expectedError(pact.Server.Port), err)
				assert.Equal(t, tc.expectedID, id)
				return nil
			}))
		})
	}
}
