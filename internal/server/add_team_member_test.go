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

type mockAddTeamMemberClient struct {
	team struct {
		count       int
		lastCookies []*http.Cookie
		lastID      int
		data        sirius.Team
		err         error
	}
	editTeam struct {
		count       int
		lastCookies []*http.Cookie
		lastTeam    sirius.Team
		err         error
	}
	searchUsers struct {
		count       int
		lastCookies []*http.Cookie
		lastSearch  string
		data        []sirius.User
		err         error
	}
}

func (c *mockAddTeamMemberClient) Team(ctx context.Context, cookies []*http.Cookie, id int) (sirius.Team, error) {
	c.team.count += 1
	c.team.lastCookies = cookies
	c.team.lastID = id

	return c.team.data, c.team.err
}

func (c *mockAddTeamMemberClient) EditTeam(ctx context.Context, cookies []*http.Cookie, team sirius.Team) error {
	c.editTeam.count += 1
	c.editTeam.lastCookies = cookies
	c.editTeam.lastTeam = team

	return c.editTeam.err
}

func (c *mockAddTeamMemberClient) SearchUsers(ctx context.Context, cookies []*http.Cookie, search string) ([]sirius.User, error) {
	c.searchUsers.count += 1
	c.searchUsers.lastCookies = cookies
	c.searchUsers.lastSearch = search

	return c.searchUsers.data, c.searchUsers.err
}

func TestGetAddTeamMember(t *testing.T) {
	assert := assert.New(t)

	client := &mockAddTeamMemberClient{}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/teams/add-member/123", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(r.Cookies(), client.team.lastCookies)
	assert.Equal(123, client.team.lastID)

	assert.Equal(0, client.editTeam.count)
	assert.Equal(0, client.searchUsers.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
	}, template.lastVars)
}

func TestGetAddTeamMemberSearch(t *testing.T) {
	assert := assert.New(t)

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 5},
		},
	}
	client.searchUsers.data = []sirius.User{
		{ID: 6},
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/teams/add-member/123?search=admin", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(r.Cookies(), client.team.lastCookies)
	assert.Equal(123, client.team.lastID)

	assert.Equal(0, client.editTeam.count)

	assert.Equal(1, client.searchUsers.count)
	assert.Equal(r.Cookies(), client.searchUsers.lastCookies)
	assert.Equal("admin", client.searchUsers.lastSearch)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
		Search:    "admin",
		Team:      client.team.data,
		Users:     client.searchUsers.data,
		Members:   map[int]bool{5: true},
	}, template.lastVars)
}

func TestGetAddTeamMemberTeamError(t *testing.T) {
	assert := assert.New(t)

	expectedError := errors.New("oops")

	client := &mockAddTeamMemberClient{}
	client.team.err = expectedError
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/teams/add-member/123", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Equal(expectedError, err)

	assert.Equal(1, client.team.count)
	assert.Equal(0, client.editTeam.count)
	assert.Equal(0, client.searchUsers.count)
	assert.Equal(0, template.count)
}

func TestGetAddTeamMemberSearchClientError(t *testing.T) {
	assert := assert.New(t)

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 5},
		},
	}
	client.searchUsers.err = sirius.ClientError("problem")
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/teams/add-member/123?search=admin", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(1, client.searchUsers.count)
	assert.Equal(0, client.editTeam.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
		Search:    "admin",
		Team:      client.team.data,
		Users:     nil,
		Errors: sirius.ValidationErrors{
			"search": {
				"": "problem",
			},
		},
	}, template.lastVars)
}

func TestGetAddTeamMemberSearchError(t *testing.T) {
	assert := assert.New(t)

	expectedError := errors.New("oops")

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 5},
		},
	}
	client.searchUsers.err = expectedError
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/teams/add-member/123?search=admin", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Equal(expectedError, err)

	assert.Equal(1, client.team.count)
	assert.Equal(1, client.searchUsers.count)
	assert.Equal(0, client.editTeam.count)
	assert.Equal(0, template.count)
}

func TestGetAddTeamMemberBadPath(t *testing.T) {
	for name, path := range map[string]string{
		"empty":       "/teams/add-member/",
		"non-numeric": "/teams/add-member/hello",
		"suffixed":    "/teams/add-member/123/no",
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r, _ := http.NewRequest("GET", path, nil)
			err := editTeam(nil, nil, "http://sirius")(nil, r)

			assert.Equal(StatusError(http.StatusNotFound), err)
		})
	}
}

func TestPostAddTeamMember(t *testing.T) {
	assert := assert.New(t)

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 4},
		},
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/teams/add-member/123", strings.NewReader("id=5&search=admin&email=system.admin@opgtest.com"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(r.Cookies(), client.team.lastCookies)
	assert.Equal(123, client.team.lastID)

	newTeam := sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 4},
			{ID: 5},
		},
	}

	assert.Equal(1, client.editTeam.count)
	assert.Equal(r.Cookies(), client.editTeam.lastCookies)
	assert.Equal(newTeam, client.editTeam.lastTeam)

	assert.Equal(1, client.searchUsers.count)
	assert.Equal(r.Cookies(), client.searchUsers.lastCookies)
	assert.Equal("admin", client.searchUsers.lastSearch)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
		Search:    "admin",
		Team:      client.team.data,
		Users:     client.searchUsers.data,
		Members:   map[int]bool{4: true, 5: true},
		Success:   "system.admin@opgtest.com",
	}, template.lastVars)
}

func TestPostAddTeamMemberClientError(t *testing.T) {
	assert := assert.New(t)

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 4},
		},
	}
	client.editTeam.err = sirius.ClientError("problem")
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/teams/add-member/123", strings.NewReader("id=5&search=admin&email=system.admin@opgtest.com"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(1, client.editTeam.count)
	assert.Equal(1, client.searchUsers.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
		Search:    "admin",
		Team:      client.team.data,
		Users:     client.searchUsers.data,
		Members:   map[int]bool{4: true, 5: true},
		Errors: sirius.ValidationErrors{
			"search": {
				"": "problem",
			},
		},
	}, template.lastVars)
}

func TestPostAddTeamMemberValidationError(t *testing.T) {
	assert := assert.New(t)

	validationErrors := sirius.ValidationErrors{
		"teamType": {
			"teamTypeInUse": "This team type is already in use",
		},
	}

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 4},
		},
	}
	client.editTeam.err = &sirius.ValidationError{
		Errors: validationErrors,
	}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/teams/add-member/123", strings.NewReader("id=5&search=admin&email=system.admin@opgtest.com"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Nil(err)

	assert.Equal(1, client.team.count)
	assert.Equal(1, client.editTeam.count)
	assert.Equal(1, client.searchUsers.count)

	assert.Equal(1, template.count)
	assert.Equal("page", template.lastName)
	assert.Equal(addTeamMemberVars{
		Path:      "/teams/add-member/123",
		SiriusURL: "http://sirius",
		Search:    "admin",
		Team:      client.team.data,
		Users:     client.searchUsers.data,
		Members:   map[int]bool{4: true, 5: true},
		Errors:    validationErrors,
	}, template.lastVars)
}

func TestPostAddTeamMemberOtherError(t *testing.T) {
	assert := assert.New(t)

	expectedError := errors.New("oops")

	client := &mockAddTeamMemberClient{}
	client.team.data = sirius.Team{
		Members: []sirius.TeamMember{
			{ID: 4},
		},
	}
	client.editTeam.err = expectedError
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/teams/add-member/123", strings.NewReader("id=5&search=admin&email=system.admin@opgtest.com"))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := addTeamMember(client, template, "http://sirius")(w, r)
	assert.Equal(expectedError, err)

	assert.Equal(1, client.team.count)
	assert.Equal(1, client.editTeam.count)
	assert.Equal(0, client.searchUsers.count)
	assert.Equal(0, template.count)
}

func TestPutAddTeamMemberTeam(t *testing.T) {
	assert := assert.New(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/teams/add-member/123", nil)

	err := addTeamMember(nil, nil, "http://sirius")(w, r)
	assert.Equal(StatusError(http.StatusMethodNotAllowed), err)
}
