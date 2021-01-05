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

package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	handler "github.com/vmware-tanzu/tanzu-observability-collector-for-aws-fargate/backend"
	"github.com/vmware-tanzu/tanzu-observability-collector-for-aws-fargate/storage"
)

var (
	version string
	commit  string
)

func main() {
	log.Printf("Starting collector version: %s commit tip: %s\n", version, commit)

	// All parameters are optional
	storageDriver := flag.String("storage_driver", "", "Storage driver to send data to")
	storageDriverOptStr := flag.String("storage_driver_options", "", `Storage driver options e.g "key1=value1, key2=value2"`)
	port := flag.Int("port", 0, "Port to listen")
	debug := flag.Bool("debug", false, "Set true to enable debug mode")

	flag.Parse()

	// Waitgroup to keep track of goroutines
	var wg sync.WaitGroup

	// Map to support the list of storage driver
	funcs := map[string]func(map[string]string, *sync.WaitGroup){"wavefront": storage.Wavefront}

	// Do not use the storage driver if user is not intended
	if *storageDriver == "" {
		message := "Storage driver is not supplied"
		if *debug == true {
			log.Println(message)
		} else {
			log.Fatal(message)
		}
	} else if *storageDriver != "" {
		// Check is supplied storage driver is supported
		_, has := funcs[*storageDriver]
		if !has {
			log.Fatal("Supplied storage driver is not supported")
		}

		// Process the storageType specific inputs
		storageDriverOpt := map[string]string{}
		if *storageDriverOptStr != "" {
			for _, item := range strings.Split(strings.TrimSpace(*storageDriverOptStr), " ") {
				kwargs := strings.Split(item, "=")
				storageDriverOpt[strings.TrimSpace(kwargs[0])] = strings.TrimSpace(kwargs[1])
			}
		}

		wg.Add(1) // create a waitgroup entry for a goroutine

		// Call the storage driver
		go funcs[*storageDriver](storageDriverOpt, &wg)
	}

	// Start the server if port is supplied
	if *port != 0 {
		http.Handle("/", http.FileServer(http.Dir("./static")))
		http.HandleFunc("/metrics", handler.MetricsHandler)
		http.HandleFunc("/stats", handler.StatsHandler)
		http.HandleFunc("/metadata", handler.MetadataHandler)

		log.Printf("Server is listening on port " + strconv.Itoa(*port))
		http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	} else {
		log.Fatal("Listener configurationn is not supplied")
	}

	if *storageDriver != "" {
		wg.Wait() // Blocking all to ensure main goroutine waits until all the go routines completed
	}
}
