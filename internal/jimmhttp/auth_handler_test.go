// Copyright 2025 Canonical.

package jimmhttp_test

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/coreos/go-oidc/v3/oidc"
	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"github.com/canonical/jimm/v3/internal/auth"
	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/jimmhttp"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
	"github.com/canonical/jimm/v3/internal/testutils/testdb"
	"github.com/canonical/jimm/v3/pkg/api/params"
)

func setupDbAndSessionStore(c *qt.C) (*db.Database, sessions.Store) {
	// Setup db ahead of time so we have access to session store
	db := &db.Database{
		DB: testdb.PostgresDB(c, time.Now),
	}
	c.Assert(db.Migrate(context.Background()), qt.IsNil)

	sqlDb, err := db.DB.DB()
	c.Assert(err, qt.IsNil)

	store, err := pgstore.NewPGStoreFromPool(sqlDb, []byte("secretsecretdigletts"))
	c.Assert(err, qt.IsNil)

	return db, store
}

func createClientWithStateCookie(c *qt.C, s *httptest.Server) *http.Client {
	jar, err := cookiejar.New(nil)
	c.Assert(err, qt.IsNil)
	jimmURL, err := url.Parse(s.URL)
	c.Assert(err, qt.IsNil)
	stateCookie := http.Cookie{Name: auth.StateKey, Value: "123"}
	jar.SetCookies(jimmURL, []*http.Cookie{&stateCookie})
	return &http.Client{Jar: jar}
}

type stubBrowserOAuthAuthenticator struct {
	authCodeURL string
	state       string
}

func (s stubBrowserOAuthAuthenticator) AuthCodeURL() (string, string, error) {
	return s.authCodeURL, s.state, nil
}

func (s stubBrowserOAuthAuthenticator) Exchange(context.Context, string) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "access-token"}, nil
}

func (s stubBrowserOAuthAuthenticator) ExtractAndVerifyIDToken(context.Context, *oauth2.Token) (*oidc.IDToken, error) {
	return nil, nil
}

func (s stubBrowserOAuthAuthenticator) Email(*oidc.IDToken) (string, error) {
	return "jimm-test@canonical.com", nil
}

func (s stubBrowserOAuthAuthenticator) UpdateIdentity(context.Context, string, *oauth2.Token) error {
	return nil
}

func (s stubBrowserOAuthAuthenticator) CreateBrowserSession(context.Context, http.ResponseWriter, *http.Request, string) error {
	return nil
}

func (s stubBrowserOAuthAuthenticator) Logout(context.Context, http.ResponseWriter, *http.Request) error {
	return nil
}

func (s stubBrowserOAuthAuthenticator) AuthenticateBrowserSession(ctx context.Context, w http.ResponseWriter, req *http.Request) (context.Context, error) {
	return ctx, nil
}

func (s stubBrowserOAuthAuthenticator) Whoami(context.Context) (*params.WhoamiResponse, error) {
	return &params.WhoamiResponse{
		DisplayName: "jimm-test",
		Email:       "jimm-test@canonical.com",
	}, nil
}

func newStubOAuthServer(c *qt.C, defaultRedirectURL string, allowedRedirectOrigins []string) *httptest.Server {
	handler, err := jimmhttp.NewOAuthHandler(jimmhttp.OAuthHandlerParams{
		Authenticator: stubBrowserOAuthAuthenticator{
			authCodeURL: "https://idp.example.com/auth",
			state:       "state-123",
		},
		DashboardFinalRedirectURL:   defaultRedirectURL,
		AllowedFinalRedirectOrigins: allowedRedirectOrigins,
	})
	c.Assert(err, qt.IsNil)
	mux := chi.NewMux()
	mux.Mount(jimmhttp.AuthResourceBasePath, handler.Routes())
	return httptest.NewServer(mux)
}

// TestBrowserLoginAndLogout goes through the flow of a browser logging in, simulating
// the cookie state and handling the callbacks are as expected. Additionally handling
// the final callback to the dashboard emulating an endpoint. See RunBrowserLogin
// where we create an additional handler to simulate the final callback to the dashboard
// from JIMM.
//
// Finally, it calls the logout using the cookie containing the identity we wish to logout.
func TestBrowserLoginAndLogout(t *testing.T) {
	c := qt.New(t)

	// Login
	db, sessionStore := setupDbAndSessionStore(c)

	cookie, jimmHTTPServer, err := jimmtest.RunBrowserLoginAndKeepServerRunning(
		db,
		sessionStore,
		jimmtest.HardcodedSafeUsername,
		jimmtest.HardcodedSafePassword,
	)
	c.Assert(err, qt.IsNil)
	defer jimmHTTPServer.Close()
	c.Assert(cookie, qt.Not(qt.Equals), "")

	// Run a whoami logged in
	req, err := http.NewRequest("GET", jimmHTTPServer.URL+jimmhttp.AuthResourceBasePath+jimmhttp.WhoAmIEndpoint, nil)
	c.Assert(err, qt.IsNil)
	parsedCookies := jimmtest.ParseCookies(cookie)
	c.Assert(parsedCookies, qt.HasLen, 1)
	req.AddCookie(parsedCookies[0])

	res, err := http.DefaultClient.Do(req)
	c.Assert(err, qt.IsNil)
	defer res.Body.Close()
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	b, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(b), qt.JSONEquals, &params.WhoamiResponse{
		DisplayName: "jimm-test",
		Email:       "jimm-test@canonical.com",
	})

	// Logout
	req, err = http.NewRequest("GET", jimmHTTPServer.URL+jimmhttp.AuthResourceBasePath+jimmhttp.LogOutEndpoint, nil)
	c.Assert(err, qt.IsNil)
	req.AddCookie(parsedCookies[0])

	res, err = http.DefaultClient.Do(req)
	c.Assert(err, qt.IsNil)
	defer res.Body.Close()
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

	// Run a whoami logged out
	req, err = http.NewRequest("GET", jimmHTTPServer.URL+jimmhttp.AuthResourceBasePath+jimmhttp.WhoAmIEndpoint, nil)
	c.Assert(err, qt.IsNil)
	parsedCookies = jimmtest.ParseCookies(cookie)
	c.Assert(parsedCookies, qt.HasLen, 1)
	req.AddCookie(parsedCookies[0])

	res, err = http.DefaultClient.Do(req)
	c.Assert(err, qt.IsNil)
	defer res.Body.Close()
	c.Assert(res.StatusCode, qt.Equals, http.StatusForbidden)

	// Run a logout with no identity
	req, err = http.NewRequest("GET", jimmHTTPServer.URL+jimmhttp.AuthResourceBasePath+jimmhttp.LogOutEndpoint, nil)
	c.Assert(err, qt.IsNil)
	res, err = http.DefaultClient.Do(req)
	c.Assert(err, qt.IsNil)
	defer res.Body.Close()
	c.Assert(res.StatusCode, qt.Equals, http.StatusForbidden)
}

func TestCallbackFailsNoState(t *testing.T) {
	c := qt.New(t)

	db, sessionStore := setupDbAndSessionStore(c)
	s, err := jimmtest.SetupTestDashboardCallbackHandler("<no dashboard needed for this test>", db, sessionStore)
	c.Assert(err, qt.IsNil)
	defer s.Close()

	u, err := url.Parse(s.URL)
	c.Assert(err, qt.IsNil)
	u = u.JoinPath(jimmhttp.AuthResourceBasePath, jimmhttp.CallbackEndpoint)
	res, err := http.Get(u.String())
	c.Assert(err, qt.IsNil)

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(b), qt.Equals, http.StatusText(http.StatusForbidden)+" - no state cookie present\n")
}

func TestCallbackFailsStateNoMatch(t *testing.T) {
	c := qt.New(t)

	db, sessionStore := setupDbAndSessionStore(c)
	s, err := jimmtest.SetupTestDashboardCallbackHandler("<no dashboard needed for this test>", db, sessionStore)
	c.Assert(err, qt.IsNil)
	defer s.Close()

	client := createClientWithStateCookie(c, s)
	callbackURL := s.URL + jimmhttp.AuthResourceBasePath + jimmhttp.CallbackEndpoint
	res, err := client.Get(callbackURL + "?state=567")
	c.Assert(err, qt.IsNil)

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(b), qt.Equals, http.StatusText(http.StatusForbidden)+" - state does not match\n")
}

func TestCallbackFailsNoCodePresent(t *testing.T) {
	c := qt.New(t)

	db, sessionStore := setupDbAndSessionStore(c)
	s, err := jimmtest.SetupTestDashboardCallbackHandler("<no dashboard needed for this test>", db, sessionStore)
	c.Assert(err, qt.IsNil)
	defer s.Close()

	client := createClientWithStateCookie(c, s)

	callbackURL := s.URL + jimmhttp.AuthResourceBasePath + jimmhttp.CallbackEndpoint
	res, err := client.Get(callbackURL + "?state=123")
	c.Assert(err, qt.IsNil)

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(b), qt.Equals, http.StatusText(http.StatusForbidden)+" - missing auth code\n")
}

func TestCallbackFailsExchange(t *testing.T) {
	c := qt.New(t)

	db, sessionStore := setupDbAndSessionStore(c)
	s, err := jimmtest.SetupTestDashboardCallbackHandler("<no dashboard needed for this test>", db, sessionStore)
	c.Assert(err, qt.IsNil)
	defer s.Close()

	client := createClientWithStateCookie(c, s)
	callbackURL := s.URL + jimmhttp.AuthResourceBasePath + jimmhttp.CallbackEndpoint
	c.Assert(err, qt.IsNil)
	res, err := client.Get(callbackURL + "?code=idonotexist&state=123")
	c.Assert(err, qt.IsNil)

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(b), qt.Equals, http.StatusText(http.StatusForbidden)+` - authorisation code exchange failed: oauth2: "invalid_grant" "Code not valid"`+"\n")
}

func TestLoginRejectsDisallowedRedirectURI(t *testing.T) {
	c := qt.New(t)

	defaultRedirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer defaultRedirect.Close()

	authServer := newStubOAuthServer(c, defaultRedirect.URL, []string{defaultRedirect.URL})
	defer authServer.Close()

	response, err := http.Get(authServer.URL + jimmhttp.AuthResourceBasePath + jimmhttp.LoginEndpoint + "?redirect_uri=" + url.QueryEscape("https://pr-123.demo.example.com/models"))
	c.Assert(err, qt.IsNil)
	defer response.Body.Close()
	c.Assert(response.StatusCode, qt.Equals, http.StatusBadRequest)
	body, err := io.ReadAll(response.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(body), qt.Equals, http.StatusText(http.StatusBadRequest)+" - redirect uri origin is not allowed\n")
}

func TestCallbackUsesRequestedRedirectURI(t *testing.T) {
	c := qt.New(t)

	defaultRedirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("default redirect"))
	}))
	defer defaultRedirect.Close()

	requestedRedirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("requested redirect"))
	}))
	defer requestedRedirect.Close()

	authServer := newStubOAuthServer(c, defaultRedirect.URL, []string{requestedRedirect.URL})
	defer authServer.Close()

	jar, err := cookiejar.New(nil)
	c.Assert(err, qt.IsNil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	loginURL := authServer.URL + jimmhttp.AuthResourceBasePath + jimmhttp.LoginEndpoint + "?redirect_uri=" + url.QueryEscape(requestedRedirect.URL+"/models")
	response, err := client.Get(loginURL)
	c.Assert(err, qt.IsNil)
	defer response.Body.Close()
	c.Assert(response.StatusCode, qt.Equals, http.StatusTemporaryRedirect)
	c.Assert(response.Header.Get("Location"), qt.Equals, "https://idp.example.com/auth")

	client.CheckRedirect = nil
	callbackURL := authServer.URL + jimmhttp.AuthResourceBasePath + jimmhttp.CallbackEndpoint + "?state=state-123&code=ok"
	response, err = client.Get(callbackURL)
	c.Assert(err, qt.IsNil)
	defer response.Body.Close()
	c.Assert(response.StatusCode, qt.Equals, http.StatusOK)
	body, err := io.ReadAll(response.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(string(body), qt.Equals, "requested redirect")
}
