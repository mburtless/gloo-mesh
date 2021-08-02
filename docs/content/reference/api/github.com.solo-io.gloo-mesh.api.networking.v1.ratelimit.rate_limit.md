
---

title: "rate_limit.proto"

---

## Package : `ratelimit.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit.proto


## Table of Contents
  - [GatewayRateLimit](#ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit)
  - [RateLimitClient](#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient)
  - [RawRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RawRateLimit)
  - [RouteRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit)







<a name="ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit"></a>

### GatewayRateLimit
Configure the Rate-Limit Filter on a Gateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitServerRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The ratelimit service to ask about ratelimit decisions. If not provided, defaults to solo.io rate-limiter server. |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout for the ratelimit service to respond. Defaults to 100ms |
  | denyOnFail | bool |  | Defaults to false |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RateLimitClient"></a>

### RateLimitClient
The RateLimitClient specifies either a simplified, abstracted rate limiting model that allows configuring the ratelimit Actions directly (raw). The corresponding server config should be set in the RateLimitConfig.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw | [ratelimit.networking.mesh.gloo.solo.io.RawRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RawRateLimit" >}}) |  | Configure the actions and/or set actions that determine how Envoy composes the descriptors |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RawRateLimit"></a>

### RawRateLimit
Use this field if you want to inline the Envoy rate limits. Note that this does not configure the rate limit server. If you are running Gloo Mesh, you need to specify the server configuration via the appropriate field in the Gloo Mesh `RateLimitConfig` resource. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rateLimits | [][ratelimit.api.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitActions" >}}) | repeated | Actions specify how the client (Envoy) will compose the descriptors that will be sent to the server to make a rate limiting decision. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit"></a>

### RouteRateLimit
Rate limit configuration for a Route or TrafficPolicy. Configures rate limits for individual HTTP routes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitServerConfigSelector | [core.skv2.solo.io.ObjectSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector" >}}) |  | Labels to the RateLimitServerConfig ref sent to the ratelimit server |
  | raw | [ratelimit.networking.mesh.gloo.solo.io.RawRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RawRateLimit" >}}) |  | Configure the actions and/or set actions that determine how Envoy composes the descriptors |
  | ratelimitClientConfigRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the RateLimitClientConfig that configures the rate limiting model |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

