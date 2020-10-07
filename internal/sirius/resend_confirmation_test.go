package sirius

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

func TestResendConfirmation(t *testing.T) {
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
		setup         func()
		cookies       []*http.Cookie
		email         string
		expectedError error
	}{
		"Created": {
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to resend a confirmation email").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/auth/resend-confirmation"),
						Headers: dsl.MapMatcher{
							"X-XSRF-TOKEN":        dsl.String("abcde"),
							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
							"OPG-Bypass-Membrane": dsl.String("1"),
							"Content-Type":        dsl.String("application/x-www-form-urlencoded"),
						},
						Body: "email=john.doe@example.com",
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusOK,
					})
			},
			cookies: []*http.Cookie{
				{Name: "XSRF-TOKEN", Value: "abcde"},
				{Name: "Other", Value: "other"},
			},
			email: "john.doe@example.com",
		},

		"Unauthorized": {
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to resend a confirmation email without cookies").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/auth/resend-confirmation"),
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

		"Errors": {
			setup: func() {
				pact.
					AddInteraction().
					Given("An admin user").
					UponReceiving("A request to resend a confirmation email errors").
					WithRequest(dsl.Request{
						Method: http.MethodPost,
						Path:   dsl.String("/auth/resend-confirmation"),
						Headers: dsl.MapMatcher{
							"OPG-Bypass-Membrane": dsl.String("1"),
						},
					}).
					WillRespondWith(dsl.Response{
						Status: http.StatusBadRequest,
					})
			},
			expectedError: errors.New("returned non-200 response: 400"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.setup()

			assert.Nil(t, pact.Verify(func() error {
				client, _ := NewClient(http.DefaultClient, fmt.Sprintf("http://localhost:%d", pact.Server.Port))

				err := client.ResendConfirmation(context.Background(), tc.cookies, tc.email)
				assert.Equal(t, tc.expectedError, err)
				return nil
			}))
		})
	}
}
