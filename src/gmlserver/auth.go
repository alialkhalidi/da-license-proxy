package gmlserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func GetAccessToken(scope string, userID string, password string, clientID string) (string, error) {
	codeVerifier, codeChallenge, err := generateCodeVerifierAndCaculateCodeChallenge()
	if err != nil {
		return "", err
	}

	authcode, err := getAuthCode(scope, authlevelCLB, userID, password, codeChallenge, clientID)
	if err != nil {
		return "", err
	}
	payload := &AccessTokenReqBody{
		Provider:     Config.CorrectProviderURL,
		AuthCode:     authcode,
		CodeVerifier: codeVerifier,
		ClientID:     clientID,
	}
	var postbody = new(AccessTokenReq)
	postbody.Body.AccessTokenBody = payload
	var expected = new(AccessTokenResp)

	err = SendRequestAndCheckResponse(accessTokenRequestMethod, postbody.Body, http.StatusAccepted, &expected.Body)
	if err != nil {
		return "", nil
	}

	/*
	   if deleteflag {
	           err = AdminDeleteLockboxForUserID(userID)
	           if err != nil {
	                   return "", fmt.Errorf("Error getting Access Token for test: %v", err)
	           }
	   }
	*/
	return expected.Body.AccessToken, nil
}

func getAuthCode(scope, authlevel string, userID string, password string, codeChallenge string, clientID string) (string, error) {

	payload := &RequestObjectReqBody{
		Provider:            Config.CorrectProviderURL,
		Audience:            Config.CorrectAudience,
		State:               correctState,
		Scopes:              scope,
		UILocales:           Config.UILocales,
		AcrValues:           authlevel,
		CodeChallengeMethod: "S256",
		CodeChallenge:       codeChallenge,
		ClientID:            clientID,
	}

	var postbody = new(RequestObjectReq)
	postbody.Body.RequestObjBody = payload
	var expected = new(RequestObjectResp)
	err := SendRequestAndCheckResponse(requestObjectRequestMethod, postbody.Body, http.StatusAccepted, &expected.Body)
	if err != nil {
		return "", err
	}

	if expected.Body.LoginURL == "" {
		return "", fmt.Errorf("handler returned unexpected body: loginurl is empty")
	}
	urlRef, err := url.Parse(expected.Body.LoginURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse the login url returned by the simulator")
	}
	parameters := urlRef.Query()

	// in case of basic auth, make sure ui_locales is present in the url
	if parameters["client_id"][0] == "myClientIDbasic" {
		//make sure ui_locales is present in the url
		if len(parameters["ui_locales"]) == 0 {
			return "", fmt.Errorf("ui_locales is not present in the url")
		}
		//make sure ui_locales matches the simulator locales
		if parameters["ui_locales"][0] != Config.UILocales {
			return "", fmt.Errorf("the locales param does not match")
		}
	}

	var authcode string
	authcode, err = MyAMGetOIDCAuthCode(userID, password, expected.Body.LoginURL)
	if err != nil {
		return "", fmt.Errorf("failed to get authcode for user %s :: %v", userID, err)
	}

	if err != nil || authcode == "" {
		return "", fmt.Errorf("failed to get authcode, err: %v", err)
	}
	return authcode, nil
}

func MyAMGetOIDCAuthCode(userID, password, loginurl string) (string, error) {
	oidcAuthURL, err := url.Parse(loginurl)
	if err != nil {
		return "", fmt.Errorf("failed to parse loginurl %v, url: %s", err, loginurl)
	}
	authenticator := NewMyAMAuthenticator(userID, password, oidcAuthURL)
	return authenticator.GetOIDCAuthCode()
}

func NewMyAMAuthenticator(userID, password string, oidcAuthURL *url.URL) *MyAMAuthenticator {
	return &MyAMAuthenticator{
		userID:      userID,
		password:    password,
		oidcAuthURL: oidcAuthURL,
	}
}

func (t *MyAMAuthenticator) GetOIDCAuthCode() (string, error) {
	t.client = &http.Client{Timeout: defaultTimeout}
	var err error
	t.client.Jar, err = cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create cookiejar for http client: %v", err)
	}

	// step 1. visit login URL
	loginPageURL := t.oidcAuthURL.String()
	loginPageURLResp, err := t.client.Get(loginPageURL)
	_, err = t.checkResponse("authorize", loginPageURL, loginPageURLResp, err)
	if err != nil {
		return "", err
	}
	defer loginPageURLResp.Body.Close()

	// step 2. submit userid/password
	authPayload := fmt.Sprintf(`
                {
                        "username": "%s",


                        "password": "%s",
                        "rememberMe": false
                }
        `, t.userID, t.password)
	authenticateResp, err := t.postRequest("authenticate", "application/json", authPayload)
	if err != nil {
		return "", err
	}
	defer authenticateResp.Body.Close()

	// from this point, on we might get redirect at any of the operations
	// depending on OIDC scopes... (ie. consent is only necessary for lockbox_creation)
	//
	// set up redirect handler here...
	t.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		query := req.URL.Query()
		if query != nil {
			t.authCode = query.Get("code")
		}
		t.lastRedirect = req.URL.String()
		return fmt.Errorf("Redirect intercepted %s", t.lastRedirect)
	}

	// step 3. submit login form
	payload := fmt.Sprintf(`username=%s&password=%s`, t.userID, t.password)
	loginResp, err := t.postRequest("login", "application/x-www-form-urlencoded", payload)
	if err != nil {
		return "", err
	}
	defer loginResp.Body.Close()
	if t.authCode != "" {
		return t.authCode, nil
	}
	if loginResp.ContentLength > 0 {
		respbody, err := ioutil.ReadAll(loginResp.Body)
		if err != nil {
			return "", err
		}

		if strings.Contains(string(respbody), `action="/myam/oidc/stepup"`) {
			// step 3.1  - step up authentication, send bogus pin
			stepUpResp, err := t.sendGetRequest("stepup", "code=1234")
			if err != nil {
				return "", err
			}
			defer stepUpResp.Body.Close()
			if t.authCode != "" {
				return t.authCode, nil
			}
		}
	}

	// step 4. submit consent
	consentResp, err := t.sendGetRequest("consent", "")
	if err != nil {
		return "", err
	}
	defer consentResp.Body.Close()
	if t.authCode != "" {
		return t.authCode, nil
	}

	return "", fmt.Errorf("No auth code obtained after successfully sending all necessary requests to MyAM")
}

func (t *MyAMAuthenticator) postRequest(operation, contentType, payload string) (*http.Response, error) {
	var body io.Reader
	if payload != "" {
		body = strings.NewReader(payload)
	}

	urlStr := t.getURL(operation)
	resp, err := t.client.Post(urlStr, contentType, body)
	return t.checkResponse(operation, urlStr, resp, err)
}

func (t *MyAMAuthenticator) sendGetRequest(operation, query string) (*http.Response, error) {
	urlStr := t.getURL(operation)
	if query != "" {
		urlStr += "?" + query
	}
	resp, err := t.client.Get(urlStr)
	return t.checkResponse(operation, urlStr, resp, err)
}

func (t *MyAMAuthenticator) getURL(operation string) string {
	opURL := url.URL{
		Scheme: t.oidcAuthURL.Scheme,
		Host:   t.oidcAuthURL.Host,
		Path:   strings.Replace(t.oidcAuthURL.Path, "/authorize", "/"+operation, 1),
	}

	return opURL.String()
}

func (t *MyAMAuthenticator) checkResponse(operation string, urlStr string, resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		// check if we already got a redirect with authCode
		if t.authCode != "" {
			// done
			return resp, nil
		}
		return nil, fmt.Errorf("Request to %s failed, error: %v, url: %s", operation, err, urlStr)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respbody, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		return resp, fmt.Errorf("%s endpoint returns non-200, status: %v, response: %s, url: %s", operation, resp.Status, respbody, urlStr)
	}
	return resp, nil
}
