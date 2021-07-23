
---

title: "istio_installation.proto"

---

## Package : `admin.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for istio_installation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## istio_installation.proto


## Table of Contents
  - [IstioInstallationSpec](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationSpec)
  - [IstioInstallationStatus](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus)
  - [IstioInstallationStatus.IstioOperatorStatus](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus)
  - [IstioInstallationStatus.IstioOperatorStatusesEntry](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatusesEntry)

  - [IstioInstallationStatus.IstioOperatorStatus.State](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus.State)
  - [IstioInstallationStatus.State](#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.State)






<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationSpec"></a>

### IstioInstallationSpec
The IstioInstallation API and it's associated features are undergoing development so this API is not currently supported.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusterNames | []string | repeated | The clusters where the IstioOperators should be installed. |
  | istioOperatorSelector | [core.skv2.solo.io.ObjectSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector" >}}) |  | Selector for the IstioOperator CRs that should be installed on the managed clusters. |
  





<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus"></a>

### IstioInstallationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the IstioInstallation metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.admin.v1alpha1.istio_installation#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.State" >}}) |  | The current state of the IstioOperator. |
  | message | string |  | A human readable message about the current state of the IstioInstallation. |
  | istioOperatorStatuses | [][admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatusesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.admin.v1alpha1.istio_installation#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatusesEntry" >}}) | repeated | The status of each IstioOperator that should be installed by Gloo Mesh, where the key is the concatenation of the IstioOperator's name, namespace, and cluster and the value is the operator's status. |
  





<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus"></a>

### IstioInstallationStatus.IstioOperatorStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the IstioOperator metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.admin.v1alpha1.istio_installation#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus.State" >}}) |  | The current state of the IstioOperator. |
  | message | string |  | A human readable message about the current state of the IstioOperator. |
  | revision | string |  | The revision tag for the associated Istio components. |
  





<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatusesEntry"></a>

### IstioInstallationStatus.IstioOperatorStatusesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.admin.v1alpha1.istio_installation#admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus" >}}) |  |  |
  




 <!-- end messages -->


<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.IstioOperatorStatus.State"></a>

### IstioInstallationStatus.IstioOperatorStatus.State
The state of a IstioOperator installation.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Waiting for resources to be installed. |
| INSTALLING | 1 | In the process of installing Istio resources on to the managed cluster. |
| HEALTHY | 2 | All Istio components were installed successfully and they are healthy. |
| ERROR | 3 | This Istio installation is in an error state. |



<a name="admin.enterprise.mesh.gloo.solo.io.IstioInstallationStatus.State"></a>

### IstioInstallationStatus.State
The state of the IstioInstallation.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Waiting for the Istio installation to be processed. |
| ACCEPTED | 1 | Finished processing the Istio installation successfully. |
| FAILED | 2 | Failed while processing the Istio installation parameters. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

