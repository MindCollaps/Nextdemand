package main

import (
	"NextDemand/main/core"
	"NextDemand/main/core/kubernetes"
	"NextDemand/main/router"
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

	//Kubernetes setup
	if env.Testing {
		fmt.Println("!!! Server is running in testing mode - kubernetes functions are disabled!")
		kubernetes.Test()
	} else {
		kubernetes.Init()
	}

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
	envLocation := ".env"
	if env.UNIX {
		envLocation = "/etc/nxdemand/.env"
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

	if !isFlagPassed("port") && os.Getenv("NXDEMAND_PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("NXDEMAND_PORT"))
		if err != nil {
			log.Println("Invalid port specified in .env file")
		} else {
			env.Port = port
		}
	}

	if err := godotenv.Load(envLocation); err != nil {
		log.Println("No .env file found")
		return
	}
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
