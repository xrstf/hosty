package oauth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/xrstf/hosty/session"
)

type googleProvider struct {
	clientID     string
	clientSecret string
	callbackURL  string
	scopes       []string
}

type googleAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type googleProfileResponse struct {
	Emails []struct {
		Value string
	}
	DisplayName string
}

func NewGoogleProvider(clientID string, clientSecret string, callbackURL string, scopes []string) Provider {
	return &googleProvider{clientID, clientSecret, callbackURL, scopes}
}

func (p *googleProvider) Start(sess *session.Session) (string, error) {
	state, err := session.RandomString(42)
	if err != nil {
		return "", errors.New("Could not create OAuth state: " + err.Error())
	}

	sess.SetOAuthState(state)

	query := url.Values{}
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.callbackURL)
	query.Set("state", state)
	query.Set("scope", strings.Join(p.scopes, " "))
	query.Set("response_type", "code")
	query.Set("approval_prompt", "auto")

	return "https://accounts.google.com/o/oauth2/auth?" + query.Encode(), nil
}

func (p *googleProvider) Finish(sess *session.Session, request *http.Request) (string, error) {
	query := request.URL.Query()
	providedState := query.Get("state")
	providedCode := query.Get("code")
	storedState := sess.OAuthState()

	if storedState == "" || providedState != storedState || providedCode == "" {
		return "", errors.New("Bad session state or invalid query string parameters.")
	}

	// exchange the temp. code for a more long-lived access token,
	// which we will need to just confirm that the code was valid.
	qs := url.Values{}
	qs.Set("client_id", p.clientID)
	qs.Set("client_secret", p.clientSecret)
	qs.Set("redirect_uri", p.callbackURL)
	qs.Set("grant_type", "authorization_code")
	qs.Set("code", providedCode)

	u := "https://accounts.google.com/o/oauth2/token"
	ct := "application/x-www-form-urlencoded"
	body := strings.NewReader(qs.Encode())

	response, err := http.Post(u, ct, body)
	if err != nil {
		return "", errors.New("Failed requesting the access token.")
	}
	defer response.Body.Close()

	r := googleAccessTokenResponse{}
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return "", errors.New("Bad access token response.")
	}

	return r.AccessToken, nil
}

func (p *googleProvider) UserProfile(accessToken string) (string, string, error) {
	query := url.Values{}
	query.Set("fields", "displayName,emails/value")
	query.Set("alt", "json")
	query.Set("access_token", accessToken)

	u := "https://www.googleapis.com/plus/v1/people/me?" + query.Encode()

	response, err := http.Get(u)
	if err != nil {
		return "", "", errors.New("Could not fetch user profile.")
	}
	defer response.Body.Close()

	profile := googleProfileResponse{}
	err = json.NewDecoder(response.Body).Decode(&profile)
	if err != nil {
		return "", "", errors.New("Bad profile response.")
	}

	email := ""

	if len(profile.Emails) > 0 {
		email = profile.Emails[0].Value
	}

	return email, profile.DisplayName, nil
}
