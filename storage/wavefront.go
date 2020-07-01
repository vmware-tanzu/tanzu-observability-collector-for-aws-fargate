package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	. "github.com/wavefronthq/wavefront-fargate-collector/backend"
	wavefrontSenders "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func Wavefront(userInput map[string]string, wg *sync.WaitGroup) {
	var sender wavefrontSenders.Sender
	var err error
	var metricPrefix string

	proxyHost := userInput["storage_driver_wf_proxy_host"]
	clusterURL := userInput["storage_driver_wf_cluster_url"]

	if proxyHost == "" && clusterURL == "" {
		log.Fatal("Please supply either proxy IP or wavefront cluster URL")
	}

	metricFlushInterval, err := strconv.Atoi(userInput["metric_flush_interval"])
	if err != nil {
		fmt.Println("Setting metrics flush interval to 60 seconds, as it is not supplied or supplied value is invalid")
		metricFlushInterval = 60 // Set default metric flush interval to 60 second if user is not intended to change
	}

	metricPrefix = userInput["storage_driver_wf_metric_prefix"]
	if metricPrefix == "" {
		metricPrefix = "aws.fargate."
	} else {
		metricPrefix = userInput["storage_driver_wf_metric_prefix"]
	}

	Debug(userInput, fmt.Sprintf("metric prefix is: %s", metricPrefix))

	if proxyHost != "" {
		proxyCfg := &wavefrontSenders.ProxyConfiguration{
			Host:                 proxyHost, // Proxy host IP or domain name
			MetricsPort:          2878,
			DistributionPort:     40000,
			TracingPort:          50000,
			FlushIntervalSeconds: metricFlushInterval,
		}

		sender, err = wavefrontSenders.NewProxySender(proxyCfg)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		wfClusterAPIToken := userInput["storage_driver_wf_cluster_api_token"]
		if wfClusterAPIToken == "" {
			log.Fatal("Please supply wavefront cluster API token")
		}

		directCfg := &wavefrontSenders.DirectConfiguration{
			Server: clusterURL,        // your Wavefront instance URL
			Token:  wfClusterAPIToken, // API token with direct ingestion permission

			// Optional configuration properties. Default values should suffice for most use cases.
			// override the defaults only if you wish to set higher values.

			// max batch of data sent per flush interval. defaults to 10,000.
			// recommended not to exceed 40,000.
			BatchSize: 10000,

			// size of internal buffer beyond which received data is dropped.
			// helps with handling brief increases in data and buffering on errors.
			// separate buffers are maintained per data type (metrics, spans and distributions)
			// defaults to 50,000. higher values could use more memory.
			MaxBufferSize: 50000,

			// interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
			// together with batch size controls the max theoretical throughput of the sender.
			FlushIntervalSeconds: metricFlushInterval,
		}

		sender, err = wavefrontSenders.NewDirectSender(directCfg)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	for range time.Tick(time.Duration(metricFlushInterval) * time.Second) {
		Debug(userInput, "metric collection loop starts")
		metrics, err := GetMetrics()
		if err != nil {
			log.Println(err.Error())
		} else {
			if metrics == nil {
				log.Println("Data not found")
			}
			hostName, _ := os.Hostname()
			Debug(userInput, fmt.Sprintf("hostname is: %s", hostName))
			for _, item := range metrics {
				Debug(userInput, fmt.Sprintf("sending metric: %s, %s, 0, %s, %#v", metricPrefix+item.Name, item.Value, hostName, item.Tags))
				sender.SendMetric(metricPrefix+item.Name, item.Value, 0, hostName, item.Tags)
			}
		}

		Debug(userInput, "metric collection loop ends")
	}
	Debug(userInput, "metric collection ends")
	sender.Close()
	wg.Done() // Specify the waitgroup about the completion of a goroutine
	return
}
