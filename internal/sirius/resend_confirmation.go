package sirius

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ResendConfirmation(ctx context.Context, cookies []*http.Cookie, email string) error {
	form := url.Values{
		"email": {email},
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/auth/resend-confirmation", strings.NewReader(form.Encode()), cookies)
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
		return newStatusError(resp)
	}

	return nil
}