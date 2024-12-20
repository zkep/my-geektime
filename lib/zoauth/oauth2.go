package zoauth

import (
	"bytes"
	"net/url"
	"strings"
)

type Config struct {
	// Oauth2 kind
	Kind string `json:"kind" yaml:"kind"`

	// ClientID is the application's ID.
	ClientID string `json:"client_id" yaml:"client_id"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"client_secret" yaml:"client_secret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint Endpoint `json:"endpoint" yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirect_url" yaml:"redirect_url"`

	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`
}

// Endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type Endpoint struct {
	AuthURL       string `json:"auth_url" yaml:"auth_url"`
	DeviceAuthURL string `json:"device_auth_url" yaml:"device_auth_url"`
	TokenURL      string `json:"token_url" yaml:"token_url"`
	UserInfoURL   string `json:"user_info_url" yaml:"user_info_url"`
	// AuthStyle optionally specifies how the endpoint wants the
	// client ID & client secret sent. The zero value means to
	// auto-detect.
	AuthStyle AuthStyle `json:"auth_style" yaml:"auth_style"`
}

// AuthStyle represents how requests for tokens are authenticated
// to the server.
type AuthStyle int

const (
	// AuthStyleAutoDetect means to auto-detect which authentication
	// style the provider wants by trying both ways and caching
	// the successful way for the future.
	AuthStyleAutoDetect AuthStyle = 0

	// AuthStyleInParams sends the "client_id" and "client_secret"
	// in the POST body as application/x-www-form-urlencoded parameters.
	AuthStyleInParams AuthStyle = 1

	// AuthStyleInHeader sends the client_id and client_password
	// using HTTP Basic Authorization. This is an optional style
	// described in the OAuth2 RFC 6749 section 2.3.1.
	AuthStyleInHeader AuthStyle = 2
)

// An AuthCodeOption is passed to Config.AuthCodeURL.
type AuthCodeOption interface {
	setValue(url.Values)
}

type setParam struct{ k, v string }

func (p setParam) setValue(m url.Values) { m.Set(p.k, p.v) }

// SetAuthURLParam builds an AuthCodeOption which passes key/value parameters
// to a provider's authorization endpoint.
func SetAuthURLParam(key, value string) AuthCodeOption {
	return setParam{key, value}
}

func (c *Config) AuthCodeURL(state string, opts ...AuthCodeOption) string {
	var buf bytes.Buffer
	buf.WriteString(c.Endpoint.AuthURL)
	v := url.Values{
		"response_type": {"code"},
		"client_id":     {c.ClientID},
	}
	if c.RedirectURL != "" {
		v.Set("redirect_uri", c.RedirectURL)
	}
	if len(c.Scopes) > 0 {
		v.Set("scope", strings.Join(c.Scopes, " "))
	}
	if state != "" {
		v.Set("state", state)
	}
	for _, opt := range opts {
		opt.setValue(v)
	}
	if strings.Contains(c.Endpoint.AuthURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(v.Encode())
	return buf.String()
}
