package main

import (
	"NextDemand/main/core"
	"NextDemand/main/core/kubernetes"
	"NextDemand/main/router"
	"NextDemand/main/web/env"
	"embed"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
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
	kubernetes.Init(env.KubeConfig)

	//Gin setup
	r := gin.Default()
	core.LoadTemplates(r)
	core.LoadServerAssets(r)

	router.InitRouter(r)
}

func flags() {
	if home := homedir.HomeDir(); home != "" {
		env.KubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		env.KubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	env.NameSpace = *flag.String("namespace", "default", "namespace to use")
	env.Host = *flag.String("host", "127.0.0.1", "host to use")

	flag.Parse()

}

func environmentSetup() {
	envLocation := ".env"
	if env.UNIX {
		envLocation = "/etc/nxdemand/.env"
	}

	if !isFlagPassed("kubeconfig") && os.Getenv("NXDEMAND_KUBECONFIG") != "" {
		config := os.Getenv("NXDEMAND_KUBECONFIG")
		env.KubeConfig = &config
	}

	if !isFlagPassed("namespace") && os.Getenv("NXDEMAND_NAMESPACE") != "" {
		env.NameSpace = os.Getenv("NXDEMAND_NAMESPACE")
	}

	if !isFlagPassed("host") && os.Getenv("NXDEMAND_HOST") != "" {
		env.Host = os.Getenv("NXDEMAND_HOST")
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
