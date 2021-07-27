
---

title: "certificate_verification.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for certificate_verification.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## certificate_verification.proto


## Table of Contents
  - [CertificateVerificationSpec](#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec)
  - [CertificateVerificationStatus](#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus)

  - [CertificateVerificationSpec.VerificationAction](#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec.VerificationAction)
  - [CertificateVerificationStatus.State](#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus.State)






<a name="networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec"></a>

### CertificateVerificationSpec
CertificateVerification is the resource by which a user can verify the traffic during a VirtualMesh certificate rotation.<br>To do this, a user would create a CertificateVerification containing: 1. The step being verified 2. The action which the user would like to kick off 3. The VirtualMesh being rotated<br> An example of a Verification for a Virtual Mesh which has just added a new root successfully would be: ```yaml apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1 kind: CertificateVerification metadata: name: successful-verification namespace: gloo-mesh spec: action: CONTINUE virtualMesh: name: my-virtual-mesh namespace: gloo-mesh step: ADDING_NEW_ROOT ```<br>An example of a Verification  for a Virtual Mesh which has failed to propagate the new intermediate would be the following: In addition this example does not specify a namespace for the virtualMesh because it is in the same namesapce as the CertificateVerification. ```yaml apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1 kind: CertificateVerification metadata: name: successful-verification namespace: gloo-mesh spec: action: ROLLBACK virtualMesh: name: my-virtual-mesh step: PROPAGATING_NEW_INTERMEDIATE ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| step | [certificates.mesh.gloo.solo.io.CertificateRotationState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.ca_options#certificates.mesh.gloo.solo.io.CertificateRotationState" >}}) |  | The rotation state to verify using this CertificateVerification. This must be an active state 1. ADDING_NEW_ROOT 2. PROPAGATING_NEW_INTERMEDIATE 3. DELETING_OLD_ROOT 4. PREVIOUS_CA |
  | action | [networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec.VerificationAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.certificate_verification#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec.VerificationAction" >}}) |  | The action which this verification will kick off |
  | virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The VirtualMesh being rotated which this resource should apply to. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus"></a>

### CertificateVerificationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the CertificateVerification metadata. If the `observedGeneration` does not match `metadata.generation`, the issuer has not processed the most recent version of this request. |
  | errors | []string | repeated | Any error observed which prevented the CertificateVerification from being processed. If the error is empty, the request has been processed successfully |
  | state | [networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.certificate_verification#networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus.State" >}}) |  | The current state of the CertificateVerification resource as reported by the rotation verifier. |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.CertificateVerificationSpec.VerificationAction"></a>

### CertificateVerificationSpec.VerificationAction
The actions available when verifying

| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTINUE | 0 | Default action. This will continue the rotation. This option should only be used if the traffic has been verified to be healthy across the VirtualMesh |
| ROLLBACK | 1 | This action will move the rotation back to the previous active state. This should be used when the traffic is unhealthy as a result of a rotation step, and you need to return to the previous good state. |



<a name="networking.enterprise.mesh.gloo.solo.io.CertificateVerificationStatus.State"></a>

### CertificateVerificationStatus.State
Possible states in which a CertificateVerification can exist.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | The CertificateVerification has yet to be picked up by the translator. |
| VERIFIED | 1 | The CertificateVerification has been used to verify a rotation step. |
| INVALID | 2 | The CertificateVerification is invalid. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

