package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
	"github.com/stretchr/testify/assert"
)

type mockChangePasswordClient struct {
	count                  int
	lastCookies            []*http.Cookie
	lastExistingPassword   string
	lastNewPassword        string
	lastNewPasswordConfirm string
	err                    error
}

func (m *mockChangePasswordClient) ChangePassword(ctx context.Context, cookies []*http.Cookie, existingPassword, newPassword, newPasswordConfirm string) error {
	m.count += 1
	m.lastCookies = cookies
	m.lastExistingPassword = existingPassword
	m.lastNewPassword = newPassword
	m.lastNewPasswordConfirm = newPasswordConfirm

	return m.err
}

func TestGetChangePassword(t *testing.T) {
	assert := assert.New(t)

	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	handler := changePassword(nil, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(changePasswordVars{
		Path:      "/path",
		SiriusURL: "http://sirius",
	}, template.lastVars)
}

func TestPostChangePassword(t *testing.T) {
	assert := assert.New(t)

	client := &mockChangePasswordClient{}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/path", strings.NewReader("currentpassword=a&password1=b&password2=c"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	handler := changePassword(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	assert.Equal(r.Cookies(), client.lastCookies)
	assert.Equal("a", client.lastExistingPassword)
	assert.Equal("b", client.lastNewPassword)
	assert.Equal("c", client.lastNewPasswordConfirm)

	assert.Equal("page", template.lastName)
	assert.Equal(changePasswordVars{
		Path:      "/path",
		SiriusURL: "http://sirius",
		Success:   true,
	}, template.lastVars)
}

func TestPostChangePasswordUnauthenticated(t *testing.T) {
	assert := assert.New(t)

	client := &mockChangePasswordClient{err: sirius.ErrUnauthorized}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/path", nil)

	handler := changePassword(client, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(sirius.ErrUnauthorized, err)

	assert.Equal(0, template.count)
}

func TestPostChangePasswordSiriusError(t *testing.T) {
	assert := assert.New(t)

	client := &mockChangePasswordClient{err: sirius.ClientError("Something happened")}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/path", nil)

	handler := changePassword(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(changePasswordVars{
		Path:      "/path",
		SiriusURL: "http://sirius",
		Errors: sirius.ValidationErrors{
			"currentpassword": {
				"": "Something happened",
			},
		},
	}, template.lastVars)
}

func TestPostChangePasswordOtherError(t *testing.T) {
	assert := assert.New(t)

	expectedErr := errors.New("oops")
	client := &mockChangePasswordClient{err: expectedErr}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/path", nil)

	handler := changePassword(client, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(expectedErr, err)

	assert.Equal(0, template.count)
}
