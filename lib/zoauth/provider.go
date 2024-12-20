package zoauth

// Identity represents the ID Token claims supported by the server.
type Identity struct {
	UserID            string
	Username          string
	PreferredUsername string
	Email             string
	EmailVerified     bool
	Groups            []string
	ConnectorData     []byte
}

type Authorize struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type Provider interface {
	AuthorizeToken(auth Authorize) (*AccessToken, error)
	Identity() (*Identity, error)
}
