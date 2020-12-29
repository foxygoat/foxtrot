package foxtrot

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var apiURLFlag = flag.String("api-base-url", "", "base URL of API")

type APITestSuite struct {
	suite.Suite
	baseURL string
	server  *httptest.Server
}

const (
	testSemver    = "v0.0.0-test"
	testCommitSha = "123456789abcdef"
)

func (s *APITestSuite) SetupSuite() {
	t := s.T()
	s.baseURL = *apiURLFlag
	if s.baseURL == "" {
		mux := http.NewServeMux()
		cfg := &Config{
			DSN:     ":memory:",
			Version: Version{Semver: testSemver, CommitSha: testCommitSha},
		}
		_, err := NewApp(cfg, mux)
		require.NoError(t, err)
		s.server = httptest.NewServer(mux)
		s.baseURL = s.server.URL
	}
}

func (s *APITestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
}

func TestAPI(t *testing.T) {
	suite.Run(t, &APITestSuite{})
}

func (s *APITestSuite) TestAPIHistory() {
	t := s.T()
	tests := map[string]string{
		"/api/history?room=$Kitchen": `[
		  {"id":5,"createdAt":"2020-11-22T13:00:42Z","room":"$Kitchen","author":"$Goat","content":"Ok, bye."},
		  {"id":4,"createdAt":"2020-11-22T12:32:42Z","room":"$Kitchen","author":"$Fox","content":"Yes."},
		  {"id":3,"createdAt":"2020-11-22T12:22:42Z","room":"$Kitchen","author":"$Goat","content":"Are you hungry?"},
		  {"id":2,"createdAt":"2020-11-22T11:12:12Z","room":"$Kitchen","author":"$Fox","content":"Hallo"},
		  {"id":1,"createdAt":"2020-11-22T11:11:11Z","room":"$Kitchen","author":"$Goat","content":"Hi"}
		]`,
		"/api/history?room=$Kitchen&before=3": `[
		  {"id":2,"createdAt":"2020-11-22T11:12:12Z","room":"$Kitchen","author":"$Fox","content":"Hallo"},
		  {"id":1,"createdAt":"2020-11-22T11:11:11Z","room":"$Kitchen","author":"$Goat","content":"Hi"}
		]`,
		"/api/history?room=$MissingRoom": "[]",
	}

	for relURL, wantBody := range tests {
		relURL, wantBody := relURL, wantBody
		t.Run(relURL, func(t *testing.T) {
			body, status := httpGet(t, s.baseURL+relURL)
			require.Equal(t, http.StatusOK, status)
			require.JSONEq(t, wantBody, body)
		})
	}
}

func (s *APITestSuite) TestAPIHistory400() {
	t := s.T()
	relURL := "/api/history?room=$Kitchen&before=NOT_A_NUMBER"
	body, status := httpGet(t, s.baseURL+relURL)
	require.Equal(t, http.StatusBadRequest, status)
	want := http.StatusText(http.StatusBadRequest) + "\n"
	require.Equal(t, want, body)
}

func (s *APITestSuite) TestLogin() {
	t := s.T()
	relURL := "/api/login"
	payload := `{"name": "$Fox", "password": "Pa$$w0rd"}`
	body, status := httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusOK, status)
	u := User{}
	err := json.Unmarshal([]byte(body), &u)
	require.NoError(t, err, body)
	want := User{Name: "$Fox", JWT: u.JWT}
	require.Equal(t, want, u)
	require.NotEmpty(t, u.JWT)
	require.Equal(t, 2, strings.Count(u.JWT, "."))
}

func (s *APITestSuite) TestLoginErr() {
	t := s.T()
	relURL := "/api/login"
	payload := `{"name": "$MISSING_USER", "password": "Pa$$w0rd"}`
	_, status := httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusUnauthorized, status)

	payload = `{"name": "$Fox", "password": "WRONG_PASSWORD"}`
	_, status = httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusUnauthorized, status)

	payload = `{"BAD_JSON`
	_, status = httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusBadRequest, status)
}

func (s *APITestSuite) TestRegister() {
	t := s.T()
	relURL := "/api/register"
	payload := fmt.Sprintf(`{"name": "%s", "password": "Pa$$w0rd"}`, testUser)
	body, status := httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusOK, status)
	u := User{}
	err := json.Unmarshal([]byte(body), &u)
	require.NoError(t, err, body)
	want := User{Name: testUser, JWT: u.JWT}
	require.Equal(t, want, u)
	require.NotEmpty(t, u.JWT)
	require.Equal(t, 2, strings.Count(u.JWT, "."))

	relURL = "/api/_test_cleanup"
	_, status = httpDelete(t, s.baseURL+relURL)
	require.Equal(t, http.StatusOK, status)
}

func (s *APITestSuite) TestVersion() {
	t := s.T()
	body, status := httpGet(t, s.baseURL+"/api/version")
	require.Equal(t, http.StatusOK, status)
	version := Version{}
	err := json.Unmarshal([]byte(body), &version)
	require.NoError(t, err, body)
	require.NotEmpty(t, version.Semver)
	require.NotEmpty(t, version.CommitSha)
	require.NotEqual(t, "undefined", version.CommitSha)
	require.NotEqual(t, "undefined", version.Semver)
	if s.server != nil {
		require.Equal(t, testCommitSha, version.CommitSha)
		require.Equal(t, testSemver, version.Semver)
	}
}

func (s *APITestSuite) TestRegisterErr() {
	t := s.T()
	relURL := "/api/register"
	payload := `{"name": "$Fox", "password": "Pa$$w0rd"}` // user already exists
	body, status := httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusBadRequest, status, body)

	payload = `{"BAD_JSON`
	body, status = httpPost(t, s.baseURL+relURL, payload)
	require.Equal(t, http.StatusBadRequest, status, body)
}

func httpGet(t *testing.T, url string) (string, int) {
	t.Helper()
	return httpDo(t, http.MethodGet, url, "")
}

func httpPost(t *testing.T, url, body string) (string, int) {
	t.Helper()
	return httpDo(t, http.MethodPost, url, body)
}

func httpDelete(t *testing.T, url string) (string, int) {
	t.Helper()
	return httpDo(t, http.MethodDelete, url, "")
}

func httpDo(t *testing.T, method, url, body string) (string, int) {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req) //nolint:gosec, noctx
	require.NoErrorf(t, err, "cannot %s %s", method, url)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "cannot read body "+url)
	return string(b), resp.StatusCode
}
