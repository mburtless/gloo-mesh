
---

title: "rate_limit_server_config.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit_server_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit_server_config.proto


## Table of Contents
  - [RateLimitServerConfigSpec](#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec)
  - [RateLimitServerConfigSpec.Raw](#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec.Raw)
  - [RateLimitServerConfigStatus](#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus)

  - [RateLimitServerConfigStatus.State](#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus.State)






<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec"></a>

### RateLimitServerConfigSpec
A `RateLimitConfig` describes the ratelimit server policy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw | [networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec.Raw]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.rate_limit_server_config#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec.Raw" >}}) |  | Define a policy using the raw configuration format used by the server and the client (Envoy). |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigSpec.Raw"></a>

### RateLimitServerConfigSpec.Raw
This object allows users to specify rate limit policies using the raw configuration formats used by the server and the client (Envoy). When using this configuration type, it is up to the user to ensure that server and client configurations match to implement the desired behavior. The server (and the client libraries that are shipped with it) will ensure that there are no collisions between raw configurations defined on separate `RateLimitConfig` resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptors | [][ratelimit.api.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Descriptor" >}}) | repeated | The descriptors that will be applied to the server. {{/* Note: validation of this field disabled because it slows down cue tremendously*/}} |
  | setDescriptors | [][ratelimit.api.solo.io.SetDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.SetDescriptor" >}}) | repeated | The set descriptors that will be applied to the server. {{/* Note: validation of this field disabled because it slows down cue tremendously*/}} |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus"></a>

### RateLimitServerConfigStatus
The current status of the `RateLimitServerConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the RateLimitServerConfig metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | state | [networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.rate_limit_server_config#networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus.State" >}}) |  | The current state of the RateLimitServerConfig. |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitServerConfigStatus.State"></a>

### RateLimitServerConfigStatus.State
Possible states of a RateLimitServerConfig resource reflected in the status by Gloo Mesh while processing a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Resources are in a Pending state before they have been processed by Gloo Mesh. |
| ACCEPTED | 1 | Resources are in a Accepted state when they are valid and have been applied successfully to the Gloo Mesh configuration. |
| REJECTED | 2 | Resources are in an Invalid state when they contain incorrect configuration parameters, such as missing required values or invalid resource references. An invalid state can also result when a resource's configuration is valid but conflicts with another resource which was accepted in an earlier point in time. |
| FAILED | 3 | Resources are in a Failed state when they contain correct configuration parameters, but the server encountered an error trying to synchronize the system to the desired state. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

