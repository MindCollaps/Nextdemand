package middleware

import (
	"github.com/gin-gonic/gin"
)

//check if the header has the api key

func LoginToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func VerifyAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		/*if c.GetBool("loggedIn") {
			user, ok := c.Get("user")
			if ok {
				dUser := user.(models.User)
				if dUser.IsAdmin {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatus(401)
		*/
		c.AbortWithStatus(401)
	}
}
