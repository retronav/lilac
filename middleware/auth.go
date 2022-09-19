package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// IndieauthResponse is the response returned by an IndieAuth endpoint for
// verifying a token.
type IndieauthResponse struct {
	Me    string         `json:"me"`
	Scope IndieauthScope `json:"scope"`
}

// IndieauthScope is a string wrapper intended for use with IndieAuth token
// scopes.
type IndieauthScope string

func (i IndieauthScope) Has(scope string) bool {
	return strings.Contains(string(i), scope)
}

func Indieauth(me string, tokenEndpoint string) gin.HandlerFunc {
	log := logrus.WithField("middleware", "indieauth")
	return func(ctx *gin.Context) {
		body, exists := ctx.Get("body")
		if !exists {
			log.Error("body is not parsed yet")
		}

		tokenFromHeader, headerOk := getTokenFromHeader(ctx.Request.Header)
		tokenFromBody, bodyOk := getTokenFromBody(body.(map[string]interface{}))

		if !headerOk && !bodyOk {
			log.Error("no auth token")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if headerOk && bodyOk {
			log.Error("auth token given in both header and body")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var token string

		if tokenFromHeader != "" {
			token = tokenFromHeader
		} else {
			token = tokenFromBody
		}
		req, err := http.NewRequest("GET", tokenEndpoint, nil)
		if err != nil {
			log.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		req.Header.Set("Accept", gin.MIMEJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var authStatus IndieauthResponse
		if err = json.NewDecoder(resp.Body).Decode(&authStatus); err != nil {
			log.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		authMeUrl, _ := url.Parse(authStatus.Me)
		meUrl, _ := url.Parse(me)

		if authMeUrl.Hostname() != meUrl.Hostname() {
			log.Errorf("expected %s to be authenticated, instead got %s",
				me, authStatus.Me)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Set("auth", authStatus)
	}
}

func getTokenFromHeader(headers http.Header) (string, bool) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", false
	}
	if !strings.HasPrefix(token, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(token, "Bearer "), true
}

func getTokenFromBody(body map[string]interface{}) (string, bool) {
	token, ok := body["access_token"].(string)
	return token, ok
}
