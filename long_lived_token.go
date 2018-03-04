package fbvideo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type LongLivedTokenGenerater struct {
	longLivedToken string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
}

func NewLongLivedTokenGenerater(clientID, clientSecret, redirectURL string) *LongLivedTokenGenerater {
	return &LongLivedTokenGenerater{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

// GenerateLongLivedToken generate a long-lived token from the short token.
func (g *LongLivedTokenGenerater) GenerateLongLivedToken(shortToken string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", "fb_exchange_token")
	form.Add("client_id", g.ClientID)
	form.Add("client_secret", g.ClientSecret)
	form.Add("fb_exchange_token", shortToken)
	endpointURL := "https://graph.facebook.com/oauth/access_token?" + form.Encode()

	resp, err := http.Get(endpointURL)
	if err != nil {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	return NewLongLivedTokenFromBody(resp.Body).AccessToken, nil
}

//RefreshLongLivedToken get a new long-lived token from an old long-lived token.
func (g *LongLivedTokenGenerater) RefreshLongLivedToken(longLivedToken string) (string, error) {
	code, err := g.getClientCode(longLivedToken)
	if err != nil {
		return "", err
	}

	return g.redeemClientCode(code)
}

func (g *LongLivedTokenGenerater) getClientCode(longLivedToken string) (string, error) {
	form := url.Values{}
	form.Add("access_token", longLivedToken)
	form.Add("client_id", g.ClientID)
	form.Add("client_secret", g.ClientSecret)
	form.Add("redirect_uri", g.RedirectURL)
	endpointURL := "https://graph.facebook.com/oauth/client_code?" + form.Encode()

	resp, err := http.Get(endpointURL)
	if err != nil {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	var jsonObject map[string]string
	json.NewDecoder(resp.Body).Decode(&jsonObject)

	return jsonObject["code"], nil
}

func (g *LongLivedTokenGenerater) redeemClientCode(code string) (string, error) {
	form := url.Values{}
	form.Add("code", code)
	form.Add("client_id", g.ClientID)
	form.Add("redirect_uri", g.RedirectURL)
	endpointURL := "https://graph.facebook.com/oauth/access_token?" + form.Encode()

	resp, err := http.Get(endpointURL)
	if err != nil {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	return NewLongLivedTokenFromBody(resp.Body).AccessToken, nil
}
