
---

title: "status.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for status.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## status.proto


## Table of Contents
  - [AppliedIngressGateway](#common.mesh.gloo.solo.io.AppliedIngressGateway)

  - [ApprovalState](#common.mesh.gloo.solo.io.ApprovalState)






<a name="common.mesh.gloo.solo.io.AppliedIngressGateway"></a>

### AppliedIngressGateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The Destination on the mesh that acts as an ingress gateway for the mesh. |
  | externalAddresses | []string | repeated | The externally accessible address(es) for this ingress gateway Destination. |
  | port | uint32 |  | The port on the ingress gateway Destination designated for receiving cross cluster traffic. |
  | externalPort | uint32 |  | The external facing port on the ingress gateway Destination designated for receiving cross cluster traffic. May differ from the destination_port if the Kubernetes Service is of type NodePort. |
  




 <!-- end messages -->


<a name="common.mesh.gloo.solo.io.ApprovalState"></a>

### ApprovalState
State of a Policy resource reflected in the status by Gloo Mesh while processing a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Resources are in a Pending state before they have been processed by Gloo Mesh. |
| ACCEPTED | 1 | Resources are in a Accepted state when they are valid and have been applied successfully to the Gloo Mesh configuration. |
| INVALID | 2 | Resources are in an Invalid state when they contain incorrect configuration parameters, such as missing required values or invalid resource references. An invalid state can also result when a resource's configuration is valid but conflicts with another resource which was accepted in an earlier point in time. |
| FAILED | 3 | Resources are in a Failed state when they contain correct configuration parameters, but the server encountered an error trying to synchronize the system to the desired state. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

