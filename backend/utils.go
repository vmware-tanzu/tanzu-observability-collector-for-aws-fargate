package backend

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// APIEndpoint to get task stats
const APIEndpoint = "http://169.254.170.2/v2/"

type Metric struct {
	Name  string            `json:"name"`
	Tags  map[string]string `json:"tags"`
	Value float64           `json:"value"`
}

func merge(m1, m2 map[string]string) {
	for k, v := range m1 {
		m2[k] = v
	}
}

func HasKey(key string, dict map[string]string) bool {
	if _, ok := dict[key]; ok {
		return ok
	}
	return false
}

func updateTagsWithClusterInfo(clusterArn string, tags map[string]string) {
	data := strings.Split(clusterArn, ":")
	tags["accountId"] = data[4]
	tags["Region"] = data[3]
	tags["clusterName"] = strings.Split(data[5], "/")[1]
}

func updateTagsWithTaskID(taskArn string, tags map[string]string) {
	data := strings.Split(taskArn, ":")
	tags["taskId"] = strings.Split(data[5], "/")[1]
}

func extractMetric(obj interface{}, key string, metrics map[string]float64) map[string]float64 {
	switch obj.(type) {
	case bool:
		if obj.(bool) {
			metrics[key] = 1
		} else {
			metrics[key] = 0
		}
		key = ""
	case float64:
		metrics[key] = obj.(float64)
		key = ""
	case map[string]interface{}:
		for k, v := range obj.(map[string]interface{}) {
			if key != "" {
				k = key + "." + k
			}
			extractMetric(v, k, metrics)
		}
	}

	return metrics
}

func callAPI(route string) (map[string]interface{}, error) {
	response, err := http.Get(APIEndpoint + route)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(response.Body)
	var obj map[string]interface{}
	var _ = json.Unmarshal([]byte(data), &obj)
	return obj, nil
}

func getContainerStats() (map[string]map[string]float64, error) {
	payLoad, err := callAPI("stats")
	if err != nil {
		return nil, err
	}
	containerStats := make(map[string]map[string]float64)

	for k, data := range payLoad {
		metrics := make(map[string]float64)
		extractMetric(data, "", metrics)
		containerStats[k] = metrics
	}
	return containerStats, nil
}

func getContainerMetadata() (map[string]map[string]string, error) {
	payLoad, err := callAPI("metadata")
	if err != nil {
		return nil, err
	}

	_clusterArn, isClusterFound := payLoad["Cluster"]
	_containers, isContainersFound := payLoad["Containers"]
	_taskArn, isTaskArnFound := payLoad["TaskARN"]

	if !isClusterFound || !isContainersFound || !isTaskArnFound || _clusterArn == nil || _taskArn == nil || _containers == nil {
		return nil, nil
	}

	containers := _containers.([]interface{})
	clusterArn := _clusterArn.(string)
	taskArn := _taskArn.(string)

	globalTags := make(map[string]string)

	updateTagsWithClusterInfo(clusterArn, globalTags)
	updateTagsWithTaskID(taskArn, globalTags)

	// add Family
	globalTags["family"] = payLoad["Family"].(string)
	// add Revision
	globalTags["revision"] = payLoad["Revision"].(string)

	containerTags := make(map[string]map[string]string)

	for _, container := range containers {
		cTags := make(map[string]string)
		container := container.(map[string]interface{})
		containerID := container["DockerId"].(string)
		cTags["name"] = container["Name"].(string)
		cTags["id"] = container["DockerId"].(string)
		cTags["dockerName"] = container["DockerName"].(string)
		cTags["type"] = container["Type"].(string)

		// merge globalTags to cTags
		merge(globalTags, cTags)

		containerTags[containerID] = cTags
	}

	return containerTags, nil
}

func GetMetrics() ([]Metric, error) {
	metrics := []Metric{}

	containerMetadata, err := getContainerMetadata()

	if err != nil || containerMetadata == nil {
		return nil, err
	}

	containersStats, err := getContainerStats()
	if err != nil || containersStats == nil {
		return nil, err
	}

	for cDockerID, stats := range containersStats {
		mdata, isFound := containerMetadata[cDockerID]
		if isFound {
			for key, value := range stats {
				metrics = append(metrics, Metric{key, mdata, value})
			}
		}
	}
	return metrics, nil
}
