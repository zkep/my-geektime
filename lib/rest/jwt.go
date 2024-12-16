package rest

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTConfig struct {
	SigningMethod  string
	SigningKey     any
	AuthScheme     string
	TokenLookup    string
	Expires        time.Duration
	SigningKeys    map[string]any
	Claims         jwt.Claims
	KeyFunc        jwt.Keyfunc
	ParseTokenFunc func(auth string) (*jwt.Token, error)
}

const (
	AlgorithmHS256 = "HS256"
)

var (
	DefaultJWTConfig = JWTConfig{
		SigningMethod: AlgorithmHS256,
		TokenLookup:   "header:Authorization,cookie:analogjwt",
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
		SigningKeys:   map[string]any{},
	}
)

func JWT(key string, expires int64) *JWTConfig {
	c := DefaultJWTConfig
	c.SigningKey = []byte(key)
	c.Expires = time.Second * time.Duration(expires)
	if c.AuthScheme == "" {
		c.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	if c.KeyFunc == nil {
		c.KeyFunc = c.DefaultKeyFunc
	}
	if c.ParseTokenFunc == nil {
		c.ParseTokenFunc = c.DefaultParseToken
	}
	return &c
}

func (config *JWTConfig) DefaultParseToken(tokenstr string) (token *jwt.Token, err error) {
	if _, ok := config.Claims.(jwt.MapClaims); ok {
		token, err = jwt.Parse(tokenstr, config.KeyFunc)
	} else {
		t := reflect.ValueOf(config.Claims).Type().Elem()
		claims := reflect.New(t).Interface().(jwt.Claims)
		token, err = jwt.ParseWithClaims(tokenstr, claims, config.KeyFunc)
	}
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

func (config *JWTConfig) DefaultKeyFunc(t *jwt.Token) (any, error) {
	if t.Method.Alg() != config.SigningMethod {
		return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
	}
	if len(config.SigningKeys) > 0 {
		if kid, ok := t.Header["kid"].(string); ok {
			if key, ok := config.SigningKeys[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
	}
	return config.SigningKey, nil
}

func (config *JWTConfig) DefaultTokenGenerator(fn func() (jwt.MapClaims, error)) (string, time.Time, error) {
	claims, err := fn()
	if err != nil {
		return "", time.Time{}, err
	}
	if claims == nil {
		claims = jwt.MapClaims{}
	}
	now := time.Now().UTC()
	expire := now.Add(config.Expires)
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = now.Unix()
	token := jwt.NewWithClaims(jwt.GetSigningMethod(config.SigningMethod), claims)
	tokenString, err := token.SignedString(config.SigningKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}
