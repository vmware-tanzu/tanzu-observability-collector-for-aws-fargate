# Wavefront Collector for AWS Fargate [![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

[Wavefront](https://docs.wavefront.com) is a high-performance streaming analytics platform for monitoring and optimizing your environment and applications.

The Wavefront Collector for AWS Fargate collects container stats from AWS Fargate and sends metrics to Wavefront.

## Web UI
The Collector provides an API endpoint for debugging purposes that can be used to check what metrics are being collected. To enable the web UI, pass in a port parameter `-port <any-port>` while starting the Collector and ensure the same port is mapped to the container as well.

## Wavefront Storage Driver
The Collector can send data to Wavefront using the [proxy](https://docs.wavefront.com/proxies.html) or via [direct ingestion](https://docs.wavefront.com/direct_ingestion.html).

### Direct Ingestion
Use direct ingestion to send the data directly to the Wavefront service. This is the simplest way to get up and running quickly.

To enable this method, start the container with the following options:
```
-storage_driver wavefront
-storage_driver_options "storage_driver_wf_cluster_url=https://<YOUR_CLUSTER>.wavefront.com -storage_driver_wf_cluster_api_token=<YOUR_API_TOKEN>"
```

**Note:** The API token must have direct ingestion permission.

### Wavefront Proxy
The Proxy is the recommended choice for a production deployment that needs resilience to internet outages, control over data queuing and filtering, and more.
This sends the data to the Wavefront proxy, which then forwards the data to the Wavefront service.

To enable this method, start the container with the following options:
```
-storage_driver wavefront
-storage_driver_options "storage_driver_wf_proxy_host=<proxy_host_IP>"
```

## Configuration Options
### Metric Prefix
The default metrics prefix is `aws.fargate.`. To override the prefix, modify the `storage_driver_options` as follows:
```
-storage_driver_options "storage_driver_wf_metric_prefix=<metrics_prefix>"
```

### Metrics Flush Interval
The default flush interval is 5 seconds. To override the interval, modify the `storage_driver_options` as follows:
```
-storage_driver_options "metric_flush_interval=<nnumber of seconds>"
```


[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-fargate-collector.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-fargate-collector
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-fargate-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-fargate-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-fargate-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-fargate-collector/
