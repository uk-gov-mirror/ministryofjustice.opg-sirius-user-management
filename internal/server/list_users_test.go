package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
	"github.com/stretchr/testify/assert"
)

type mockListUsersClient struct {
	count      int
	lastCtx    sirius.Context
	lastSearch string
	err        error
	data       []sirius.User
}

func (m *mockListUsersClient) SearchUsers(ctx sirius.Context, search string) ([]sirius.User, error) {
	m.count += 1
	m.lastCtx = ctx
	m.lastSearch = search

	return m.data, m.err
}

func TestListUsers(t *testing.T) {
	assert := assert.New(t)

	data := []sirius.User{
		{
			ID:          29,
			DisplayName: "Milo Nihei",
			Email:       "milo.nihei@opgtest.com",
			Status:      "Active",
		},
	}
	client := &mockListUsersClient{
		data: data,
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path?search=milo", nil)

	handler := listUsers(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(getContext(r), client.lastCtx)

	assert.Equal(1, client.count)
	assert.Equal("milo", client.lastSearch)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(listUsersVars{
		Path:      "/path",
		SiriusURL: "http://sirius",

		Search: "milo",
		Users: []sirius.User{
			{
				ID:          29,
				DisplayName: "Milo Nihei",
				Email:       "milo.nihei@opgtest.com",
				Status:      "Active",
			},
		},
	}, template.lastVars)
}

func TestListUsersRequiresSearch(t *testing.T) {
	assert := assert.New(t)

	data := []sirius.User{
		{
			ID:          29,
			DisplayName: "Milo Nihei",
			Email:       "milo.nihei@opgtest.com",
			Status:      "Active",
		},
	}
	client := &mockListUsersClient{
		data: data,
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler := listUsers(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)

	assert.Equal(0, client.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(listUsersVars{
		Path:      "/path",
		SiriusURL: "http://sirius",

		Search: "",
		Users:  nil,
	}, template.lastVars)
}

func TestListUsersClientError(t *testing.T) {
	assert := assert.New(t)

	client := &mockListUsersClient{
		err: sirius.ClientError("problem"),
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path?search=m", nil)

	handler := listUsers(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)

	assert.Equal(1, client.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(listUsersVars{
		Path:      "/path",
		SiriusURL: "http://sirius",

		Search: "m",
		Users:  nil,
		Errors: sirius.ValidationErrors{
			"search": {
				"": "problem",
			},
		},
	}, template.lastVars)
}

func TestListUsersSiriusErrors(t *testing.T) {
	assert := assert.New(t)

	expectedErr := errors.New("err")
	client := &mockListUsersClient{err: expectedErr}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/?search=long", nil)

	handler := listUsers(client, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(expectedErr, err)
	assert.Equal(0, template.count)
}

func TestPostListUsers(t *testing.T) {
	assert := assert.New(t)
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	handler := listUsers(nil, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(StatusError(http.StatusMethodNotAllowed), err)

	assert.Equal(0, template.count)
}
