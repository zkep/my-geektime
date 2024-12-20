package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/types/oauth"
	"go.uber.org/zap"
)

type Oauth struct{}

func NewOauth() *Oauth {
	return &Oauth{}
}

func (*Oauth) Authorize(c *gin.Context) {
	var r oauth.Authorize
	err := c.ShouldBind(&r)
	if err != nil {
		global.LOG.Error("Authorize.ShouldBind", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}

	client := http.DefaultClient

	var (
		clientId, clientSecret, tokenURL, userInfoURL, redirectUri string
	)

	for _, provider := range global.CONF.Oauth2 {
		if strings.HasPrefix(r.State, provider.Kind) {
			clientId = provider.ClientID
			clientSecret = provider.ClientSecret
			tokenURL = provider.Endpoint.TokenURL
			userInfoURL = provider.Endpoint.UserInfoURL
			redirectUri = provider.RedirectURL
			break
		}
	}

	if len(clientId) == 0 || len(clientSecret) == 0 {
		global.LOG.Error("Authorize not support kind", zap.Any("req", r))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	// from code get token
	accessTokenReq := oauth.AuthorizeToken{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Code:         r.Code,
		GrantType:    "authorization_code",
		RedirectUri:  redirectUri,
	}
	raw, _ := json.Marshal(accessTokenReq)
	httpReq, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(raw))
	if err != nil {
		global.LOG.Error("Authorize.NewRequest", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")

	response, err := client.Do(httpReq)
	if err != nil {
		global.LOG.Error("Authorize.Do", zap.Error(err), zap.String("req", string(raw)))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	defer func() { _ = response.Body.Close() }()
	ret, err := httputil.DumpResponse(response, true)
	if err != nil {
		global.LOG.Error("Authorize.DumpResponse", zap.Error(err), zap.String("req", string(raw)))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	global.LOG.Debug("Authorize.Dump",
		zap.String("req", string(raw)), zap.String("ret", string(ret)))
	var access oauth.AuthorizeTokenResponse
	if err = json.NewDecoder(response.Body).Decode(&access); err != nil {
		global.LOG.Error("Authorize.Decode", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	// get user info
	if strings.HasPrefix(r.State, "gitee") {
		userInfoURL += "?access_token=" + access.AccessToken
	}
	userInfoHttpReq, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		global.LOG.Error("Authorize.NewRequest", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	userInfoHttpReq.Header.Add("Content-Type", "application/json")
	userInfoHttpReq.Header.Add("Authorization", "Bearer "+access.AccessToken)

	userInfoResponse, err := client.Do(userInfoHttpReq)
	if err != nil {
		global.LOG.Error("Authorize.Do", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	defer func() { _ = userInfoResponse.Body.Close() }()
	userRet, err := httputil.DumpResponse(userInfoResponse, true)
	if err != nil {
		global.LOG.Error("Authorize.userInfoResponse", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	global.LOG.Debug("Authorize.userInfo", zap.String("userRet", string(userRet)))

	var userInfo oauth.AuthorizeUserInfo
	if err = json.NewDecoder(userInfoResponse.Body).Decode(&userInfo); err != nil {
		global.LOG.Error("Authorize.Decode", zap.Error(err))
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
		return
	}
	if access.ExpiresIn == 0 {
		access.ExpiresIn = int64(time.Hour * 24)
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", access.AccessToken,
		int(access.ExpiresIn), "/", "", true, false)

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?t=%d", time.Now().Unix()))
}
