package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
	"github.com/stretchr/testify/assert"
)

type mockMyDetailsClient struct {
	count            int
	permissionsCount int
	lastCtx          sirius.Context
	err              error
	permissionsErr   error
	data             sirius.MyDetails
	hasPermission    bool
	lastPermission   struct {
		Group  string
		Method string
	}
}

func (m *mockMyDetailsClient) MyDetails(ctx sirius.Context) (sirius.MyDetails, error) {
	m.count += 1
	m.lastCtx = ctx

	return m.data, m.err
}

func (m *mockMyDetailsClient) HasPermission(ctx sirius.Context, group string, method string) (bool, error) {
	m.permissionsCount += 1
	m.lastCtx = ctx
	m.lastPermission.Group = group
	m.lastPermission.Method = method

	return m.hasPermission, m.permissionsErr
}

func TestGetMyDetails(t *testing.T) {
	assert := assert.New(t)

	data := sirius.MyDetails{
		ID:          123,
		Firstname:   "John",
		Surname:     "Doe",
		Email:       "john@doe.com",
		PhoneNumber: "123",
		Roles:       []string{"A", "COP User", "B"},
		Teams: []sirius.MyDetailsTeam{
			{DisplayName: "A Team"},
		},
	}
	client := &mockMyDetailsClient{data: data}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler := myDetails(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(getContext(r), client.lastCtx)

	assert.Equal(1, client.count)
	assert.Equal(1, client.permissionsCount)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(myDetailsVars{
		Path:               "/path",
		SiriusURL:          "http://sirius",
		ID:                 123,
		Firstname:          "John",
		Surname:            "Doe",
		Email:              "john@doe.com",
		PhoneNumber:        "123",
		Organisation:       "COP User",
		Roles:              []string{"A", "B"},
		Teams:              []string{"A Team"},
		CanEditPhoneNumber: false,
	}, template.lastVars)
}

func TestGetMyDetailsUsesPermission(t *testing.T) {
	assert := assert.New(t)

	data := sirius.MyDetails{
		ID:          123,
		Firstname:   "John",
		Surname:     "Doe",
		Email:       "john@doe.com",
		PhoneNumber: "123",
		Roles:       []string{"A", "COP User", "B"},
		Teams: []sirius.MyDetailsTeam{
			{DisplayName: "A Team"},
		},
	}
	client := &mockMyDetailsClient{data: data, hasPermission: true}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler := myDetails(client, template, "http://sirius")
	err := handler(w, r)

	assert.Nil(err)

	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(getContext(r), client.lastCtx)
	assert.Equal("user", client.lastPermission.Group)
	assert.Equal("patch", client.lastPermission.Method)

	assert.Equal(1, client.count)
	assert.Equal(1, client.permissionsCount)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(myDetailsVars{
		Path:               "/path",
		SiriusURL:          "http://sirius",
		ID:                 123,
		Firstname:          "John",
		Surname:            "Doe",
		Email:              "john@doe.com",
		PhoneNumber:        "123",
		Organisation:       "COP User",
		Roles:              []string{"A", "B"},
		Teams:              []string{"A Team"},
		CanEditPhoneNumber: true,
	}, template.lastVars)
}

func TestGetMyDetailsUnauthenticated(t *testing.T) {
	assert := assert.New(t)

	client := &mockMyDetailsClient{err: sirius.ErrUnauthorized}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "", nil)

	handler := myDetails(client, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(sirius.ErrUnauthorized, err)

	assert.Equal(0, template.count)
}

func TestGetMyDetailsSiriusErrors(t *testing.T) {
	assert := assert.New(t)

	client := &mockMyDetailsClient{err: errors.New("err")}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "", nil)

	handler := myDetails(client, template, "http://sirius")
	err := handler(w, r)

	assert.Equal("err", err.Error())

	assert.Equal(0, template.count)
}

func TestPostMyDetails(t *testing.T) {
	assert := assert.New(t)
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	handler := myDetails(nil, template, "http://sirius")
	err := handler(w, r)

	assert.Equal(StatusError(http.StatusMethodNotAllowed), err)

	assert.Equal(0, template.count)
}
