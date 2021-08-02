
---

title: "rate_limit_client_config.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit_client_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit_client_config.proto


## Table of Contents
  - [RateLimitClientConfigSpec](#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigSpec)
  - [RateLimitClientConfigStatus](#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus)

  - [RateLimitClientConfigStatus.State](#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus.State)






<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigSpec"></a>

### RateLimitClientConfigSpec
RateLimitClientConfig contains the client configuration for the rate limit Actions that determine how Envoy composes the descriptors that are sent to the rate limit server to check whether a request should be rate-limited


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rateLimits | [ratelimit.networking.mesh.gloo.solo.io.RateLimitClient]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient" >}}) |  | The RateLimitClient specifies the ratelimit Actions which the client (Envoy) will use to compose the descriptors that will be sent to the server to make a rate limiting decision. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus"></a>

### RateLimitClientConfigStatus
The current status of the `RateLimitClientConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the RateLimitClientConfig metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | state | [networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.rate_limit_client_config#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus.State" >}}) |  | The current state of the RateLimitClientConfig. |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus.State"></a>

### RateLimitClientConfigStatus.State
Possible states of a RateLimitClientConfig resource reflected in the status by Gloo Mesh while processing a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Resources are in a Pending state before they have been processed by Gloo Mesh. |
| ACCEPTED | 1 | Resources are in a Accepted state when they are valid and have been applied successfully to the Gloo Mesh configuration. |
| INVALID | 2 | Resources are in an Invalid state when they contain incorrect configuration parameters, such as missing required values or invalid resource references. An invalid state can also result when a resource's configuration is valid but conflicts with another resource which was accepted in an earlier point in time. |
| FAILED | 3 | Resources are in a Failed state when they contain correct configuration parameters, but the server encountered an error trying to synchronize the system to the desired state. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

