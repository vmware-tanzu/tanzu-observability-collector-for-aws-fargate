# Fargate Collector

It collects the container stats from AWS Fargate.

## Web UI
It exposes the API endpoint for debugging purpose, It will be used to check what metrics are getting collected. To enable the web UI user needs to send a port parameter `-port <any-port>` while starting the container and same port should be mapped for the container as well.

## Supported Storage Driver
* ### Wavefront
    Wavefront storage driver sends data to any wavefront cluster using the below methods
    * #### Direct ingestion 
        It sends the data directly to the Wavefront service. This is the simplest way to get up and running quickly. To enable this method user needs to send the below parameters while starting the container
            `-storage_driver wavefront`
            `-storage_driver_options "storage_driver_wf_cluster_url=https://<your_cluster>.wavefront.com, storage_driver_wf_cluster_api_token=<your_cluster_api>"`
        NOTE - API token must have direct ingestion permission.
        
    * #### Wavefront proxy
        It sends the data to the Wavefront proxy, which then forwards the data to the Wavefront service. This is the recommended choice for a large-scale deployment that needs resilience to internet outages, control over data queuing and filtering, and more. To enable this method user needs to send the below parameter while starting the container
            `-storage_driver wavefront`
            `-storage_driver_options "storage_driver_wf_proxy_host=<proxy_host_IP>"`
    
    ### Other Options    
    * #### Metric Prefix
        Default metrics prefix is `aws.fargate.`, to override it user needs to send below storage_driver_options
        `-storage_driver_options "storage_driver_wf_metric_prefix=<metrics_prefix>"`
    * #### Metric Flush Interval
        Default metric flush innterval is `5 Seconds`, to override it user needs to send below storage_driver_options
        `-storage_driver_options "metric_flush_interval=<nnumber of seconds>"`