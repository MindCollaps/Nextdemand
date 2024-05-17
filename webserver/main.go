package main

import (
	"NextDemand/main/core"
	"NextDemand/main/core/kubernetes"
	"NextDemand/main/router"
	"NextDemand/main/tasks"
	"NextDemand/main/web/env"
	"embed"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

//go:embed main/web/*
var Files embed.FS

func main() {
	log.Println("--------------------------------")
	log.Println("Starting NextDemand")
	_, err := Files.ReadDir("main/web")
	if err != nil {
		log.Println("Failed to read public files - this is likely a problem during compilation. Exiting...")
		return
	}

	env.Files = Files

	// command line arguments
	flags()

	//Environment setup
	environmentSetup()
	printEnv()

	//Kubernetes setup
	if env.Testing {
		fmt.Println("!!! Server is running in testing mode - kubernetes functions are disabled!")
		kubernetes.Test()
	} else {
		kubernetes.Init()
	}

	//Checker tasks
	tasks.StartRepeatingTasks()

	//Gin setup
	r := gin.Default()
	core.LoadTemplates(r)
	core.LoadServerAssets(r)

	router.InitRouter(r)

	r.Run(":" + strconv.Itoa(env.Port))
}

func flags() {
	flag.StringVar(&env.NameSpace, "namespace", "default", "namespace to use")
	flag.StringVar(&env.Host, "host", "127.0.0.1", "host to use")
	flag.BoolVar(&env.Testing, "test", false, "testing mode")
	flag.IntVar(&env.Port, "port", 8080, "port to use")

	flag.Parse()
}

func environmentSetup() {
	if env.UseEnvFile {
		envLocation := ".env"
		if env.UNIX {
			envLocation = "/etc/nxdemand/.env"
		}

		if err := godotenv.Load(envLocation); err != nil {
			log.Println("No .env file found")
			return
		}
	}

	if !isFlagPassed("namespace") && os.Getenv("NXDEMAND_NAMESPACE") != "" {
		env.NameSpace = os.Getenv("NXDEMAND_NAMESPACE")
	}

	if !isFlagPassed("host") && os.Getenv("NXDEMAND_HOST") != "" {
		env.Host = os.Getenv("NXDEMAND_HOST")
	}

	if !isFlagPassed("test") && os.Getenv("TESTING") == "true" {
		env.Testing = true
	}

	if os.Getenv("CHECK_TIME") == "true" {
		env.CheckTime = true
	}

	if os.Getenv("CHECK_SIMULTANEOUS") == "true" {
		env.CheckSimultaneous = true
	}

	if os.Getenv("TIME_ALIVE") != "" {
		timeAlive, err := strconv.Atoi(os.Getenv("TIME_ALIVE"))
		if err != nil {
			log.Println("Invalid time alive specified in .env file")
		} else {
			env.TimeAlive = timeAlive
		}
	}

	if os.Getenv("SIMULTANEOUS_INSTANCES") != "" {
		simultaneousInstances, err := strconv.Atoi(os.Getenv("SIMULTANEOUS_INSTANCES"))
		if err != nil {
			log.Println("Invalid simultaneous instances specified in .env file")
		} else {
			env.SimultaneousInstances = simultaneousInstances
		}
	}

	if !isFlagPassed("port") && os.Getenv("NXDEMAND_PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("NXDEMAND_PORT"))
		if err != nil {
			log.Println("Invalid port specified in .env file")
		} else {
			env.Port = port
		}
	}
}

func printEnv() {
	fmt.Println("Namespace:", env.NameSpace)
	fmt.Println("Host:", env.Host)
	fmt.Println("Testing:", env.Testing)
	fmt.Println("CheckTime:", env.CheckTime)
	fmt.Println("CheckSimultaneous:", env.CheckSimultaneous)
	fmt.Println("TimeAlive:", env.TimeAlive)
	fmt.Println("SimultaneousInstances:", env.SimultaneousInstances)
	fmt.Println("Port:", env.Port)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
