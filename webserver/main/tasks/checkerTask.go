package tasks

import (
	"NextDemand/main/core/kubernetes"
	"NextDemand/main/router"
	"NextDemand/main/web/env"
	"fmt"
	"time"
)

func initCheckerTasks() {
	_, err := Scheduler.Every(3).Minutes().Do(func() {
		fmt.Println("Checking for expired IPs")
		for ip, t := range router.RequestedIps {
			if time.Since(t) > time.Duration(env.TimeAlive)*time.Minute {
				kubernetes.DeleteInstance(router.RequestedIpsInstance[ip])
				delete(router.RequestedIps, ip)
				delete(router.RequestedIpsInstance, ip)
			}
		}
	})

	if err != nil {
		fmt.Println("Failed to start checker task")
		fmt.Println(err)
	}
}
