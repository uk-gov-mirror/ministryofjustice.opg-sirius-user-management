package sirius

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

type editMyDetailsBadRequestResponse struct {
	Status           int    `json:"status" pact:"example=400"`
	Detail           string `json:"detail" pact:"example=Payload failed validation"`
	ValidationErrors *struct {
		PhoneNumber *struct {
			StringLengthTooLong string `json:"stringLengthTooLong" pact:"example=The input is more than 255 characters long"`
		} `json:"phoneNumber"`
	} `json:"validation_errors"`
}

func TestMyDetails(t *testing.T) {
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
		name              string
		setup             func()
		cookies           []*http.Cookie
		expectedMyDetails MyDetails
		expectedError     error
	}{
		{
			name: "OK",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request to get my details").
					WithRequest(dsl.Request{
						Method: http.MethodGet,
						Path:   dsl.String("/api/v1/users/current"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status:  http.StatusOK,
						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
						Body: dsl.Like(map[string]interface{}{
							"id":          dsl.Like(47),
							"name":        dsl.Like("system"),
							"phoneNumber": dsl.Like("03004560300"),
							"teams": dsl.EachLike(map[string]interface{}{
								"displayName": dsl.Like("Allocations - (Supervision)"),
							}, 1),
							"displayName": dsl.Like("system admin"),
							"deleted":     dsl.Like(false),
							"email":       dsl.Like("system.admin@opgtest.com"),
							"firstname":   dsl.Like("system"),
							"surname":     dsl.Like("admin"),
							"roles":       dsl.EachLike("System Admin", 1),
							"locked":      dsl.Like(false),
							"suspended":   dsl.Like(false),
						}),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			expectedMyDetails: MyDetails{
				ID:          47,
				Name:        "system",
				PhoneNumber: "03004560300",
				Teams: []MyDetailsTeam{
					{DisplayName: "Allocations - (Supervision)"},
				},
				DisplayName: "system admin",
				Deleted:     false,
				Email:       "system.admin@opgtest.com",
				Firstname:   "system",
				Surname:     "admin",
				Roles:       []string{"string"},
				Locked:      false,
				Suspended:   false,
			},
		},

		{
			name: "Unauthorized",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request to get my details without cookies").
					WithRequest(dsl.Request{
						Method: http.MethodGet,
						Path:   dsl.String("/api/v1/users/current"),
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

				myDetails, err := client.MyDetails(context.Background(), tc.cookies)
				assert.Equal(t, tc.expectedMyDetails, myDetails)
				assert.Equal(t, tc.expectedError, err)
				return nil
			}))
		})
	}
}

func TestEditMyDetails(t *testing.T) {
	pact := &dsl.Pact{
		Consumer:          "sirius-user-management",
		Provider:          "sirius",
		Host:              "localhost",
		PactFileWriteMode: "merge",
		LogDir:            "../../logs",
		PactDir:           "../../pacts",
	}
	defer pact.Teardown()

	testCases := map[string]struct {
		phoneNumber   string
		setup         func()
		cookies       []*http.Cookie
		expectedError error
	}{
		"OK": {
			phoneNumber: "01210930320",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request to change my phone number").
					WithRequest(dsl.Request{
						Method: http.MethodPut,
						Path:   dsl.String("/api/v1/users/47/updateTelephoneNumber"),
						Headers: dsl.MapMatcher{
							"Content-type":        dsl.String("application/json"),
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
						Body: map[string]string{
							"phoneNumber": "01210930320",
						},
					}).
					WillRespondWith(dsl.Response{
						Status:  http.StatusOK,
						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
						Body: dsl.Like(map[string]interface{}{
							"id":          dsl.Like(47),
							"name":        dsl.Like("system"),
							"phoneNumber": dsl.Like("03004560300"),
							"teams": dsl.EachLike(map[string]interface{}{
								"displayName": dsl.Like("Allocations - (Supervision)"),
							}, 1),
							"displayName": dsl.Like("system admin"),
							"deleted":     dsl.Like(false),
							"email":       dsl.Like("system.admin@opgtest.com"),
							"firstname":   dsl.Like("system"),
							"surname":     dsl.Like("admin"),
							"roles":       dsl.EachLike("System Admin", 1),
							"locked":      dsl.Like(false),
							"suspended":   dsl.Like(false),
						}),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
		},

		"BadRequest": {
			phoneNumber: "invalid phone number",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("An invalid request to change my phone number").
					WithRequest(dsl.Request{
						Method: http.MethodPut,
						Path:   dsl.String("/api/v1/users/47/updateTelephoneNumber"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
						Body: map[string]string{
							"phoneNumber": "invalid phone number",
						},
					}).
					WillRespondWith(dsl.Response{
						Status:  http.StatusBadRequest,
						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/problem+json")},
						Body:    dsl.Match(editMyDetailsBadRequestResponse{}),
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			expectedError: &ValidationError{
				Message: "Payload failed validation",
				Errors: ValidationErrors{
					"phoneNumber": {
						"stringLengthTooLong": "The input is more than 255 characters long",
					},
				},
			},
		},

		"Unauthorized": {
			phoneNumber: "01210930320",
			setup: func() {
				pact.
					AddInteraction().
					Given("User exists").
					UponReceiving("A request to change my phone number without cookies").
					WithRequest(dsl.Request{
						Method: http.MethodPut,
						Path:   dsl.String("/api/v1/users/47/updateTelephoneNumber"),
						Headers: dsl.MapMatcher{
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
						Body: map[string]string{
							"phoneNumber": "01210930320",
						},
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusUnauthorized,
					})
			},
			expectedError: ErrUnauthorized,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.setup()

			assert.Nil(t, pact.Verify(func() error {
				client, _ := NewClient(http.DefaultClient, fmt.Sprintf("http://localhost:%d", pact.Server.Port))

				err := client.EditMyDetails(context.Background(), tc.cookies, 47, tc.phoneNumber)
				assert.Equal(t, tc.expectedError, err)
				return nil
			}))
		})
	}
}
