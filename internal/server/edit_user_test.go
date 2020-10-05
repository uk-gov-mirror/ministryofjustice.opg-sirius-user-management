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

type mockEditUserClient struct {
	user struct {
		count       int
		lastCookies []*http.Cookie
		lastID      int
		data        sirius.AuthUser
		err         error
	}

	editUser struct {
		count       int
		lastCookies []*http.Cookie
		lastUser    sirius.AuthUser
		err         error
	}
}

func (m *mockEditUserClient) User(ctx context.Context, cookies []*http.Cookie, id int) (sirius.AuthUser, error) {
	m.user.count += 1
	m.user.lastCookies = cookies
	m.user.lastID = id

	return m.user.data, m.user.err
}

func (m *mockEditUserClient) EditUser(ctx context.Context, cookies []*http.Cookie, user sirius.AuthUser) error {
	m.editUser.count += 1
	m.editUser.lastCookies = cookies
	m.editUser.lastUser = user

	return m.editUser.err
}

func TestGetEditUser(t *testing.T) {
	assert := assert.New(t)

	client := &mockEditUserClient{}
	client.user.data = sirius.AuthUser{Firstname: "test"}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/edit-user/123", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := editUser(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.user.count)
	assert.Equal(123, client.user.lastID)
	assert.Equal(0, client.editUser.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(editUserVars{
		Path:      "/edit-user/123",
		SiriusURL: "http://sirius",
		User:      client.user.data,
	}, template.lastVars)
}

func TestGetEditUserBadPath(t *testing.T) {
	for name, path := range map[string]string{
		"empty":       "/edit-user/",
		"non-numeric": "/edit-user/hello",
		"suffixed":    "/edit-user/123/no",
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &mockEditUserClient{}
			template := &mockTemplate{}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", path, nil)
			r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

			err := editUser(client, template, "http://sirius")(w, r)
			assert.Equal(StatusError(http.StatusNotFound), err)

			assert.Equal(0, client.user.count)
			assert.Equal(0, client.editUser.count)
			assert.Equal(0, template.count)
		})
	}
}

func TestPostEditUser(t *testing.T) {
	assert := assert.New(t)

	client := &mockEditUserClient{}
	client.user.data = sirius.AuthUser{Firstname: "test"}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/edit-user/123", strings.NewReader("email=a&firstname=b&surname=c&organisation=d&roles=e&roles=f&locked=Yes&suspended=No"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := editUser(client, template, "http://sirius")(w, r)
	assert.Equal(RedirectError("/users"), err)

	assert.Equal(1, client.editUser.count)
	assert.Equal(r.Cookies(), client.editUser.lastCookies)
	assert.Equal(sirius.AuthUser{
		ID:           123,
		Firstname:    "b",
		Surname:      "c",
		Organisation: "d",
		Roles:        []string{"e", "f"},
		Locked:       true,
		Suspended:    false,
	}, client.editUser.lastUser)

	assert.Equal(0, client.user.count)
	assert.Equal(0, template.count)
}

func TestPostEditUserClientError(t *testing.T) {
	assert := assert.New(t)

	client := &mockEditUserClient{}
	client.editUser.err = sirius.ClientError("something")
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/edit-user/123", strings.NewReader("email=a&firstname=b&surname=c&organisation=d&roles=e&roles=f&locked=Yes&suspended=No"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	err := editUser(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.editUser.count)
	assert.Equal(0, client.user.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(editUserVars{
		Path:      "/edit-user/123",
		SiriusURL: "http://sirius",
		User: sirius.AuthUser{
			ID:           123,
			Firstname:    "b",
			Surname:      "c",
			Organisation: "d",
			Roles:        []string{"e", "f"},
			Locked:       true,
			Suspended:    false,
		},
		Errors: sirius.ValidationErrors{
			"firstname": {
				"": "something",
			},
		},
	}, template.lastVars)
}

func TestPostEditUserOtherError(t *testing.T) {
	assert := assert.New(t)

	expectedErr := errors.New("oops")
	client := &mockEditUserClient{}
	client.editUser.err = expectedErr
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/edit-user/123", nil)

	err := editUser(client, template, "http://sirius")(w, r)
	assert.Equal(expectedErr, err)

	assert.Equal(1, client.editUser.count)
	assert.Equal(0, client.user.count)
	assert.Equal(0, template.count)
}