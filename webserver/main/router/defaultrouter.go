package router

import (
	"NextDemand/main/core/kubernetes"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.GET("/spawn", func(c *gin.Context) {
		kubernetes.SpawnNewNextcloudDeployment("testing")
		c.JSON(200, gin.H{
			"message": "Spawned new deployment",
		})
	})
}
