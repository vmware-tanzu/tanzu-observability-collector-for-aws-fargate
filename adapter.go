package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	handler "github.com/wavefronthq/wavefront-fargate-collector/backend"
	"github.com/wavefronthq/wavefront-fargate-collector/storage"
)

func main() {
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
		}else {
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
			for _, item := range strings.Split(*storageDriverOptStr, " ") {
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
