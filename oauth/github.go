package oauth

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"go.xrstf.de/hosty/session"
)

type githubProvider struct {
	clientID     string
	clientSecret string
	callbackURL  string
	scopes       []string
}

type githubProfileResponse struct {
	Login string
	Name  string
}

func NewGithubProvider(clientID string, clientSecret string, callbackURL string, scopes []string) Provider {
	return &githubProvider{clientID, clientSecret, callbackURL, scopes}
}

func (p *githubProvider) Start(sess *session.Session) (string, error) {
	state, err := session.RandomString(42)
	if err != nil {
		return "", errors.New("Could not create OAuth state: " + err.Error())
	}

	sess.SetOAuthState(state)

	query := url.Values{}
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.callbackURL)
	query.Set("state", state)
	query.Set("scope", strings.Join(p.scopes, ","))
	query.Set("response_type", "code")
	query.Set("approval_prompt", "auto")

	return "https://github.com/login/oauth/authorize?" + query.Encode(), nil
}

func (p *githubProvider) Finish(sess *session.Session, request *http.Request) (string, error) {
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
	qs.Set("alt", "json")
	qs.Set("code", providedCode)

	u := "https://github.com/login/oauth/access_token"
	ct := "application/x-www-form-urlencoded"
	body := strings.NewReader(qs.Encode())

	response, err := http.Post(u, ct, body)
	if err != nil {
		return "", errors.New("Failed requesting the access token.")
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.New("Could not read response body.")
	}

	parsed, err := url.ParseQuery(string(responseBody))
	if err != nil {
		return "", errors.New("Response body not a valid query string.")
	}

	token := parsed.Get("access_token")

	if token == "" {
		return "", errors.New("Could not acquire an access token.")
	}

	return token, nil
}

func (p *githubProvider) UserProfile(accessToken string) (string, string, error) {
	query := url.Values{}
	query.Set("access_token", accessToken)

	u := "https://api.github.com/user?" + query.Encode()

	response, err := http.Get(u)
	if err != nil {
		return "", "", errors.New("Could not fetch user profile.")
	}
	defer response.Body.Close()

	profile := githubProfileResponse{}
	err = json.NewDecoder(response.Body).Decode(&profile)
	if err != nil {
		return "", "", errors.New("Bad profile response.")
	}

	return profile.Login, profile.Name, nil
}
