package sirius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ChangePassword(ctx context.Context, cookies []*http.Cookie, oldPassword, newPassword, newPasswordConfirm string) error {
	form := url.Values{
		"existingPassword": {oldPassword},
		"password":         {newPassword},
		"confirmPassword":  {newPasswordConfirm},
	}
	body := strings.NewReader(form.Encode())

	req, err := c.newRequest(ctx, http.MethodPost, "/auth/change-password", body, cookies)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		var v struct {
			Errors string `json:"errors"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil {
			return ClientError(v.Errors)
		}

		return fmt.Errorf("returned non-2XX response: %d", resp.StatusCode)
	}

	return nil
}
