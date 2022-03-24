package clarify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/clarify/clarify-go/jsonrpc"
	"golang.org/x/oauth2/clientcredentials"
)

// Credentials contain a data-structure with Clarify integration credentials.
type Credentials struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`

	APIURL      string          `json:"apiUrl"`
	Integration string          `json:"integration"`
	Credentials CredentialsAuth `json:"credentials"`
}

// CredentialsAuth contains the information that is used to authenticate
// credentials against Clarify.
type CredentialsAuth struct {
	Type         string `json:"type"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// Supported credentials types.
const (
	TypeBasicAuth         = "basic-auth"
	TypeClientCredentials = "client-credentials"
)

const (
	defaultAPIURL = "https://api.clarify.io/v1/"
)

// BasicAuthCredentials returns basic auth credentials for use with Clarify.
func BasicAuthCredentials(integrationID, password string) *Credentials {
	c := Credentials{
		APIURL:      defaultAPIURL,
		Integration: integrationID,
	}
	c.Credentials.Type = TypeBasicAuth
	c.Credentials.ClientID = integrationID
	c.Credentials.ClientSecret = password
	return &c
}

// CredentialsFromFile parse Clarify Credentials from the passed in filename,
// and return either valid credentials or an error.
func CredentialsFromFile(name string) (creds *Credentials, err error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer appendOnError(&err, r.Close, "; ")
	return CredentialsFromReader(r)
}

// CredentialsFromReader parse Clarify Credentials from the passed in reader,
// and return either valid credentials or an error.
func CredentialsFromReader(r io.Reader) (*Credentials, error) {
	var creds Credentials
	dec := json.NewDecoder(r)

	if err := dec.Decode(&creds); err != nil {
		return nil, err
	}
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Validate returns an error if the credentials are invalid.
func (creds *Credentials) Validate() error {
	issues := map[string][]string{}
	if creds.APIURL == "" {
		issues["apiUrl"] = []string{"required"}
	} else if u, err := url.Parse(creds.APIURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		issues["apiUrl"] = []string{"must be a valid HTTP(S) URL"}
	}
	if creds.Integration == "" {
		issues["integration"] = []string{"required"}
	}
	if creds.Credentials.ClientID == "" {
		issues["credentials.clientId"] = []string{"required"}
	}
	if creds.Credentials.ClientSecret == "" {
		issues["credentials.clientSecret"] = []string{"required"}
	}
	switch creds.Credentials.Type {
	case TypeBasicAuth, TypeClientCredentials:
		// pass
	case "":
		issues["credentials.type"] = []string{"required"}
	default:
		issues["credentials.type"] = []string{"not in [basic client-credentials]"}
	}
	if len(issues) > 0 {
		return joinErrors(ErrBadCredentials, PathErrors(issues), ": ")
	}
	return nil
}

// Client returns a new Clarify client for the current credentials, assuming the
// client credentials to be valid. If the credentials are invalid, this method
// will return a non-functional client where all requests result return the
// ErrBadCredentials error.
func (creds Credentials) Client(ctx context.Context) *Client {
	var h jsonrpc.Handler

	h, err := creds.HTTPHandler(ctx)
	if err != nil {
		h = invalidRPCHandler{err: err}
	}

	return &Client{integration: creds.Integration, h: h}
}

// HTTPHandler returns a low-level RPC handler that communicates over HTTP using
// the credentials in creds.
func (creds Credentials) HTTPHandler(ctx context.Context) (*jsonrpc.HTTPHandler, error) {
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	apiURL := strings.TrimRight(creds.APIURL, "/") + "/"

	var c http.Client
	switch creds.Credentials.Type {
	case TypeBasicAuth:
		c.Transport = basicAuthTransport{
			parent: http.DefaultTransport,
			user:   creds.Credentials.ClientID,
			pass:   creds.Credentials.ClientSecret,
		}
	case TypeClientCredentials:
		cfg := clientcredentials.Config{
			ClientID:     creds.Credentials.ClientID,
			ClientSecret: creds.Credentials.ClientSecret,
			TokenURL:     apiURL + "oauth/token",
			EndpointParams: url.Values{
				"audience": {apiURL},
			},
		}
		c = *cfg.Client(ctx)
	default:
		// This code-path is impossible because creds.Validate() should have
		// returned an error.
		panic(ErrBadCredentials)
	}
	c.Timeout = 20 * time.Second

	return &jsonrpc.HTTPHandler{Client: c, URL: apiURL + "rpc"}, nil
}

var _ http.RoundTripper = basicAuthTransport{}

type basicAuthTransport struct {
	parent http.RoundTripper
	user   string
	pass   string
}

func (t basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.user, t.pass)
	return t.parent.RoundTrip(req)
}

func (t basicAuthTransport) CloseIdleConnections() {
	if i, ok := t.parent.(interface{ CloseIdleConnections() }); ok {
		i.CloseIdleConnections()
	}
}

var _ jsonrpc.Handler = invalidRPCHandler{}

type invalidRPCHandler struct {
	err error
}

func (h invalidRPCHandler) Do(ctx context.Context, req jsonrpc.Request, result any) error {
	return h.err
}
