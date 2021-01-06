/*
 * Copyright 2019-2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/vmware-tanzu/tanzu-observability-collector-for-aws-fargate/backend"
	wavefrontSenders "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func Wavefront(userInput map[string]string, wg *sync.WaitGroup) {
	var sender wavefrontSenders.Sender
	var err error
	var metricPrefix string
	var metricFlushInterval, proxyMetricPort int

	proxyHost := userInput["storage_driver_wf_proxy_host"]
	clusterURL := userInput["storage_driver_wf_cluster_url"]

	if proxyHost == "" && clusterURL == "" {
		log.Fatal("Please supply either proxy IP or wavefront cluster URL")
	}

	metricFlushIntervalUserInput := userInput["metric_flush_interval"]
	if metricFlushIntervalUserInput == "" {
		metricFlushInterval = 60 // Set default metric flush interval to 60 second if user is not intended to change
	} else {
		metricFlushInterval, err = strconv.Atoi(metricFlushIntervalUserInput)
		if err != nil {
			log.Fatal("Supplied value for metrics flush interval is invalid, must be an integer")
		}
	}

	metricPrefix = userInput["storage_driver_wf_metric_prefix"]
	if metricPrefix == "" {
		metricPrefix = "aws.fargate."
	} else {
		metricPrefix = userInput["storage_driver_wf_metric_prefix"]
	}

	Debug(userInput, fmt.Sprintf("metric prefix is: %s", metricPrefix))

	if proxyHost != "" {
		if strings.Contains(proxyHost, ":") {
			log.Fatal("Supplied value for proxy host IP is invalid, should not contain colon")
		}

		proxyMetricPortUserInput := userInput["storage_driver_wf_metric_port"]
		if proxyMetricPortUserInput == "" {
			proxyMetricPort = 2878 // Set default metric port to 2878 if user is not intended to change
		} else {
			proxyMetricPort, err = strconv.Atoi(proxyMetricPortUserInput)
			if err != nil {
				log.Fatal("Supplied value for metric port is invalid, must be a valid port number")
			}
		}

		Debug(userInput, fmt.Sprintf("Using proxy host: %s", proxyHost))

		proxyCfg := &wavefrontSenders.ProxyConfiguration{
			Host:                 proxyHost, // Proxy host IP or domain name
			MetricsPort:          proxyMetricPort,
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
				Debug(userInput, fmt.Sprintf("sending metric: %s, %f, 0, %s, %v", metricPrefix+item.Name, item.Value, hostName, item.Tags))
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
