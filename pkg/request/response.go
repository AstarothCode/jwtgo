package request

import (
	"jwtgo/pkg/request/schema"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetCookies(c *gin.Context, cookies []schema.Cookie) {
	for _, cookieData := range cookies {
		cookie := &http.Cookie{
			Name:     cookieData.Name,
			Value:    cookieData.Value,
			Path:     "/",
			Domain:   "",
			Expires:  time.Now().UTC().Add(cookieData.Duration),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(c.Writer, cookie)
	}
}