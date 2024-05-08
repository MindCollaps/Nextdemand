package router

import (
	"NextDemand/main/core/kubernetes"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

var RequestedIps = make(map[string]time.Time)
var RequestedIpsInstance = make(map[string]string)

func InitRouter(r *gin.Engine) {
	r.GET("/spawn", func(c *gin.Context) {
		// Check if the IP has requested a deployment in the last 5 minutes
		if RequestedIps[c.ClientIP()] != (time.Time{}) {
			t := RequestedIps[c.ClientIP()]

			if time.Since(t) < 10*time.Minute {
				c.JSON(429, gin.H{
					"message": "Rate limit exceeded",
				})
				return
			}
		}

		uid, err := kubernetes.GetRandomId()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{
				"message": "Internal server error - kubernetes error",
			})
			return
		}

		RequestedIps[c.ClientIP()] = time.Now()

		password, err := kubernetes.SpawnNewNextcloudDeployment(uid)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{
				"message": "Internal server error - kubernetes error - could not spawn deployment",
			})
			return
		}
		RequestedIpsInstance[c.ClientIP()] = uid
		c.JSON(200, gin.H{
			"message": "Spawned new deployment",
			"uid":     uid,
			"pass":    password,
		})
	})
}
