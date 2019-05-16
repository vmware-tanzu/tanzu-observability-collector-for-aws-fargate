package backend

import (
	"encoding/json"
	"io"
	"net/http"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := GetMetrics()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(metrics)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get(APIEndpoint + "metadata")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(w, response.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get(APIEndpoint + "stats")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(w, response.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
