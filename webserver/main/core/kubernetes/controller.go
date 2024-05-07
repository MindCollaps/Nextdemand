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
		_, err := ClientSet.Resource(jobRes).Namespace(env.NameSpace).Get(context.TODO(), "nextcloud-job-"+uid.String(), metav1.GetOptions{})
		if err == nil {
			return uid.String(), nil
		}
		count++
		if count > 10 {
			return "", fmt.Errorf("Failed to generate unique id")
		}
	}
}

func SpawnNewNextcloudDeployment(instanceId string) {
	fmt.Println("Spawning new Nextcloud deployment with instanceId:" + instanceId)
	var yamlData, err = os.ReadFile("./nextcloud.yml")
	if env.Testing {
		yamlData, err = os.ReadFile("../kubernetes/nextcloud.yml")

	}
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	// Define the deployment structure
	deployment := &unstructured.Unstructured{}

	// Unmarshal YAML data into the deployment structure
	if err := yaml.Unmarshal(yamlData, &deployment.Object); err != nil {
		fmt.Println("Error unmarshaling YAML:", err)
		return
	}

	//Job deployment
	//Change job name to include instanceId
	deployment.Object["job"].(map[string]interface{})["metadata"].(map[string]interface{})["name"] = "nextcloud-job-" + instanceId
	//Change metadata label instanceId
	deployment.Object["job"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId
	//Change metadata label instanceId in spec
	deployment.Object["job"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["instanceId"] = instanceId

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
	deployment.Object["ingress"].(map[string]interface{})["spec"].(map[string]interface{})["rules"].([]interface{})[0].(map[string]interface{})["host"] = instanceId + "." + env.Host
	//Change spec service name to include instanceId
	deployment.Object["ingress"].(map[string]interface{})["spec"].(map[string]interface{})["rules"].([]interface{})[0].(map[string]interface{})["http"].(map[string]interface{})["paths"].([]interface{})[0].(map[string]interface{})["backend"].(map[string]interface{})["service"].(map[string]interface{})["name"] = "nextcloud-service-" + instanceId

	jobRes := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	job := unstructured.Unstructured{Object: deployment.Object["job"].(map[string]interface{})}

	//Output job as json formatted string
	json, _ := job.MarshalJSON()
	fmt.Println(string(json))

	serviceRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	service := unstructured.Unstructured{Object: deployment.Object["service"].(map[string]interface{})}

	json, _ = service.MarshalJSON()
	fmt.Println(string(json))

	ingressRes := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	ingress := unstructured.Unstructured{Object: deployment.Object["ingress"].(map[string]interface{})}

	json, _ = ingress.MarshalJSON()
	fmt.Println(string(json))

	_, err = ClientSet.Resource(jobRes).Namespace(env.NameSpace).Create(context.TODO(), &job, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating job:", err)
	}

	fmt.Println("Created job")

	_, err = ClientSet.Resource(serviceRes).Namespace(env.NameSpace).Create(context.TODO(), &service, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating service:", err)
	}

	fmt.Println("Created service")

	_, err = ClientSet.Resource(ingressRes).Namespace(env.NameSpace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error creating ingress:", err)
	}

	fmt.Println("Created ingress")
}

func DeleteInstance(instanceId string) {
	jobRes := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	serviceRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	ingressRes := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}

	err := ClientSet.Resource(jobRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-job-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting job:", err)
	}

	err = ClientSet.Resource(serviceRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-service-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting service:", err)
	}

	err = ClientSet.Resource(ingressRes).Namespace(env.NameSpace).Delete(context.TODO(), "nextcloud-ingress-"+instanceId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error deleting ingress:", err)
	}
}
