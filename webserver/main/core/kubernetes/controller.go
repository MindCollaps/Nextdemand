package kubernetes

import (
	"NextDemand/main/web/env"
	"context"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"math/rand"
	"os"
)

var ClientSet *dynamic.DynamicClient

func Init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ClientSet = clientset
	DeleteAllRunning()
}

func Test() {
	fmt.Println("Testing mode")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Test failed - but thats normal since we dont have a kubernetes cluster running in testing mode")
		}
	}()
	SpawnNewNextcloudDeployment("test")
}

func GetRandomId() (string, error) {
	uid := uuid.New()

	count := 0
	for {
		jobRes := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
		data, _ := ClientSet.Resource(jobRes).Namespace(env.NameSpace).Get(context.TODO(), "nextcloud-job-"+uid.String(), metav1.GetOptions{})
		if data == nil {
			return uid.String(), nil
		}
		count++
		if count > 10 {
			return "", fmt.Errorf("Failed to generate unique id")
		}
		uid = uuid.New()
	}
}

func SpawnNewNextcloudDeployment(instanceId string) (string, error) {
	fmt.Println("Spawning new Nextcloud deployment with instanceId:" + instanceId)
	var yamlData, err = os.ReadFile("./nextcloud.yml")
	if env.Testing {
		yamlData, err = os.ReadFile("../kubernetes/nextcloud.yml")

	}
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return "", err
	}

	// Define the deployment structure
	deployment := &unstructured.Unstructured{}

	// Unmarshal YAML data into the deployment structure
	if err := yaml.Unmarshal(yamlData, &deployment.Object); err != nil {
		fmt.Println("Error unmarshaling YAML:", err)
		return "", err
	}

	//Job deployment
	//Change dep name to include instanceId
	deployment.Object["dep"].(map[string]interface{})["metadata"].(map[string]interface{})["name"] = "nextcloud-dep-" + instanceId

	//Change metadata label instanceId
	deployment.Object["dep"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId

	//Change metadata label instanceId in spec
	deployment.Object["dep"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId

	//Change spec selector instanceId
	deployment.Object["dep"].(map[string]interface{})["spec"].(map[string]interface{})["selector"].(map[string]interface{})["matchLabels"].(map[string]interface{})["instanceId"] = instanceId

	//Change default password
	password := generateRandomPassword(10)
	deployment.Object["dep"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})[1].(map[string]interface{})["value"] = password

	//Change the domain
	deployment.Object["dep"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})[2].(map[string]interface{})["value"] = instanceId + "." + env.Host

	//Service deployment
	//Change service name to include instanceId
	deployment.Object["service"].(map[string]interface{})["metadata"].(map[string]interface{})["name"] = "nextcloud-service-" + instanceId
	//Change metadata label instanceId
	deployment.Object["service"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId
	//Change spec selector instanceId
	deployment.Object["service"].(map[string]interface{})["spec"].(map[string]interface{})["selector"].(map[string]interface{})["instanceId"] = instanceId

	//Ingress deployment
	//Change ingress name to include instanceId
	deployment.Object["ingress"].(map[string]interface{})["metadata"].(map[string]interface{})["name"] = "nextcloud-ingress-" + instanceId
	//Change metadata label instanceId
	deployment.Object["ingress"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId
	//Change spec rules host to include instanceId
	deployment.Object["ingress"].(map[string]interface{})["spec"].(map[string]interface{})["routes"].([]interface{})[0].(map[string]interface{})["match"] = "Host(`" + instanceId + "." + env.Host + "`)"
	//Change spec service name to include instanceId
	deployment.Object["ingress"].(map[string]interface{})["spec"].(map[string]interface{})["routes"].([]interface{})[0].(map[string]interface{})["services"].([]interface{})[0].(map[string]interface{})["name"] = "nextcloud-service-" + instanceId

	depRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	dep := unstructured.Unstructured{Object: deployment.Object["dep"].(map[string]interface{})}

	//Output dep as json formatted string
	json, _ := dep.MarshalJSON()
	fmt.Println(string(json))

	serviceRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	service := unstructured.Unstructured{Object: deployment.Object["service"].(map[string]interface{})}

	json, _ = service.MarshalJSON()
	fmt.Println(string(json))

	ingressRes := schema.GroupVersionResource{Group: "traefik.containo.us", Version: "v1alpha1", Resource: "ingressroutes"}
	ingress := unstructured.Unstructured{Object: deployment.Object["ingress"].(map[string]interface{})}

	json, _ = ingress.MarshalJSON()
	fmt.Println(string(json))

	_, err = ClientSet.Resource(depRes).Namespace(env.NameSpace).Create(context.TODO(), &dep, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating dep:", err)
		return "", err
	}

	fmt.Println("Created dep")

	_, err = ClientSet.Resource(serviceRes).Namespace(env.NameSpace).Create(context.TODO(), &service, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating service:", err)
		return "", err
	}

	fmt.Println("Created service")

	_, err = ClientSet.Resource(ingressRes).Namespace(env.NameSpace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating ingress:", err)
		return "", err
	}

	fmt.Println("Created ingress")

	return password, nil
}

func DeleteAllRunning() {
	depRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "deployments"}
	serviceRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	//ingressRes := schema.GroupVersionResource{Group: "traefik.containo.us", Version: "v1alpha1", Resource: "IngressRoute"}
	pod := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	err := ClientSet.Resource(depRes).Namespace(env.NameSpace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "app=nextcloud",
	})
	if err != nil {
		fmt.Println("Error deleting jobs:", err)
	}

	err = ClientSet.Resource(serviceRes).Namespace(env.NameSpace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "app=nextcloud",
	})
	if err != nil {
		fmt.Println("Error deleting services:", err)
	}

	/*
		err = ClientSet.Resource(ingressRes).Namespace(env.NameSpace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
			LabelSelector: "app=nextcloud",
		})
		if err != nil {
			fmt.Println("Error deleting ingresses:", err)
		}
	*/

	err = ClientSet.Resource(pod).Namespace(env.NameSpace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "app=nextcloud",
	})
	if err != nil {
		fmt.Println("Error deleting pods:", err)
	}
}

func DeleteInstance(instanceId string) {
	depRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	serviceRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	//ingressRes := schema.GroupVersionResource{Group: "traefik.containo.us", Version: "v1alpha1", Resource: "IngressRoute"}

	err := ClientSet.Resource(depRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-dep-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting deployment nextcloud-dep-"+instanceId+": ", err)
	} else {
		fmt.Println("Deleted deployment nextcloud-dep-" + instanceId)

	}

	err = ClientSet.Resource(serviceRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-service-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting service nextcloud-service-"+instanceId+": ", err)
	} else {
		fmt.Println("Deleted service nextcloud-service-" + instanceId)
	}

	/*err = ClientSet.Resource(ingressRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-ingress-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting nextcloud-ingress-"+instanceId+": ", err)
	} else {
		fmt.Println("Deleted ingress nextcloud-ingress-" + instanceId)
	}
	*/
}

func generateRandomPassword(size int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%&")
	b := make([]rune, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
