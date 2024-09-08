Here, i have defined two GET endpoints, one that will be exposed and the other for internal
Service to Service calls to `track`endpoint.

To support a concurrency of around 12k requests per second, we are using async writes to the files
using 100 go routines and channels with size of 120. Therefore, the service can support around 12K requests per second 
and for the calls to become blocking, the channels need to get filled.

Here, i have used a map to maintain the unique Ids received every minute and reset it every minute.

We can track the current size of each channel using grafana metrics along with measuring the latency.
