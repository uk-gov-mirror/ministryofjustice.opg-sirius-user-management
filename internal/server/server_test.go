package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-user-management/internal/sirius"
	"github.com/stretchr/testify/assert"
)

type mockTemplate struct {
	count    int
	lastName string
	lastVars interface{}
}

func (m *mockTemplate) ExecuteTemplate(w io.Writer, name string, vars interface{}) error {
	m.count += 1
	m.lastName = name
	m.lastVars = vars
	return nil
}

func TestErrorHandler(t *testing.T) {
	assert := assert.New(t)

	tmplError := &mockTemplate{}

	wrap := errorHandler(nil, tmplError, "/prefix", "http://sirius")
	handler := wrap("handler", func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusTeapot)
		return nil
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(http.StatusTeapot, resp.StatusCode)
	assert.Equal(0, tmplError.count)
}

func TestErrorHandlerUnauthorized(t *testing.T) {
	assert := assert.New(t)

	tmplError := &mockTemplate{}

	wrap := errorHandler(nil, tmplError, "/prefix", "http://sirius")
	handler := wrap("handler", func(w http.ResponseWriter, r *http.Request) error {
		return sirius.ErrUnauthorized
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(http.StatusFound, resp.StatusCode)
	assert.Equal("http://sirius/auth", resp.Header.Get("Location"))

	assert.Equal(0, tmplError.count)
}

func TestErrorHandlerRedirect(t *testing.T) {
	assert := assert.New(t)

	tmplError := &mockTemplate{}

	wrap := errorHandler(nil, tmplError, "/prefix", "http://sirius")
	handler := wrap("handler", func(w http.ResponseWriter, r *http.Request) error {
		return RedirectError("/here")
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(http.StatusFound, resp.StatusCode)
	assert.Equal("/prefix/here", resp.Header.Get("Location"))

	assert.Equal(0, tmplError.count)
}

func TestErrorHandlerStatus(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	tmplError := &mockTemplate{}

	wrap := errorHandler(logger, tmplError, "/prefix", "http://sirius")
	handler := wrap("handler", func(w http.ResponseWriter, r *http.Request) error {
		return StatusError(http.StatusTeapot)
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)

	handler.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(http.StatusInternalServerError, resp.StatusCode)

	assert.Equal(1, tmplError.count)
	assert.Equal(errorVars{SiriusURL: "http://sirius", Code: http.StatusInternalServerError, Error: "418 I'm a teapot"}, tmplError.lastVars)

	assert.Equal("handler: 418 I'm a teapot\n", buf.String())
}

func TestErrorHandlerStatusKnown(t *testing.T) {
	for name, code := range map[string]int{
		"Forbidden": http.StatusForbidden,
		"Not Found": http.StatusNotFound,
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			tmplError := &mockTemplate{}

			wrap := errorHandler(logger, tmplError, "/prefix", "http://sirius")
			handler := wrap("handler", func(w http.ResponseWriter, r *http.Request) error {
				return StatusError(code)
			})

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/path", nil)

			handler.ServeHTTP(w, r)

			resp := w.Result()
			assert.Equal(code, resp.StatusCode)

			assert.Equal(1, tmplError.count)
			assert.Equal(errorVars{SiriusURL: "http://sirius", Code: code, Error: fmt.Sprintf("%d %s", code, name)}, tmplError.lastVars)

			assert.Equal(fmt.Sprintf("handler: %d %s\n", code, name), buf.String())
		})
	}
}

type mockSystemAdminOnlyClient struct {
	count int
	roles []string
	err   error
}

func (m *mockSystemAdminOnlyClient) MyDetails(ctx context.Context, cookies []*http.Cookie) (sirius.MyDetails, error) {
	m.count += 1

	return sirius.MyDetails{Roles: m.roles}, m.err
}

func TestSystemAdminOnly(t *testing.T) {
	assert := assert.New(t)

	client := &mockSystemAdminOnlyClient{}
	client.roles = []string{"System Admin"}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := systemAdminOnly(client)(func(w http.ResponseWriter, r *http.Request) error {
		return StatusError(http.StatusTeapot)
	})(w, r)

	assert.Equal(StatusError(http.StatusTeapot), err)
	assert.Equal(1, client.count)
}

func TestSystemAdminOnlyMissingRole(t *testing.T) {
	assert := assert.New(t)

	client := &mockSystemAdminOnlyClient{}
	client.roles = []string{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/path", nil)
	r.AddCookie(&http.Cookie{Name: "test", Value: "val"})

	err := systemAdminOnly(client)(func(w http.ResponseWriter, r *http.Request) error {
		return StatusError(http.StatusTeapot)
	})(w, r)

	assert.Equal(StatusError(http.StatusForbidden), err)
	assert.Equal(1, client.count)
}

func TestSystemAdminOnlyMyDetailsError(t *testing.T) {
	assert := assert.New(t)

	expectedErr := errors.New("oops")
	client := &mockSystemAdminOnlyClient{}
	client.err = expectedErr

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/path", nil)

	err := systemAdminOnly(client)(func(w http.ResponseWriter, r *http.Request) error {
		return StatusError(http.StatusTeapot)
	})(w, r)

	assert.Equal(expectedErr, err)
	assert.Equal(1, client.count)
}
