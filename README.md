# statusProxy 
statusProxy is a small service intended to proxy external webhook incoming via Cloud Run to a VM on Cloud Compute for the purposes of utilizing load-balancing and SSL certs - whilst being able to standardise incoming without introducing general bugs

## Settings
|envar|usage|
|-|-|
|`PORT`|Port for the proxy server. Defaults to 8080|
|`PROXY_TO`|The fully qualified URL for the backend server|

## Rationale
Few reasons

1. Cloud Run has https by default and backend is on Cloud Compute which requires acquisition of SSL certs and running a load-balancer which raises costs significantly.
1. **StatusSentry** need not be overcomplicated by parsing, authenticating and standardizing incoming requests
1. Allows for custom response calls needed by various services to be abstracted from **StatusSentry**.
1. Allows for using the proxy as a custom global load balancer and proper scaling when required
