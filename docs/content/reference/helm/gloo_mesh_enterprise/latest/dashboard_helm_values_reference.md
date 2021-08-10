
---
title: "Dashboard"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|licenseKey|string| |Gloo Mesh Enterprise license key|
|forwardingRelayAddress|string|enterprise-networking-admin.gloo-mesh.svc.cluster.local:11100|Address to access the enterprise networking admin service at|
|metricsAddress|string|enterprise-networking.gloo-mesh.svc.cluster.local:9900|Address of the server to read metrics from|
|relayClientAuthority|string|enterprise-networking|SNI name used to connect to relay forwarding server|
|settingsName|string|settings|Name of the dashboard settings object to use|
|auth|struct|{"enabled":false,"backend":"","oidc":{"clientId":"","clientSecret":"","clientSecretRef":{},"issuerUrl":"","appUrl":"","session":{"backend":"","redis":{"host":""}}}}|Authentication configuration|
|auth.enabled|bool|false|Require authentication to access the dashboard|
|auth.backend|string| |Authentication backend to use. Supports: oidc|
|auth.oidc|struct|{"clientId":"","clientSecret":"","clientSecretRef":{},"issuerUrl":"","appUrl":"","session":{"backend":"","redis":{"host":""}}}|Settings for the OpenID Connect backend. Only used when backend is set to 'oidc'.|
|auth.oidc.clientId|string| |OIDC client ID|
|auth.oidc.clientSecret|string| |Plaintext OIDC client secret. Will be base64 encoded and stored in a secret with the reference below.|
|auth.oidc.clientSecretRef|struct|{}|Reference that a secret containing the client secret will be stored at|
|auth.oidc.clientSecretRef.name|string| ||
|auth.oidc.clientSecretRef.namespace|string| ||
|auth.oidc.issuerUrl|string| |OIDC Issuer |
|auth.oidc.appUrl|string| |URL users will use to access the dashboard|
|auth.oidc.session|struct|{"backend":"","redis":{"host":""}}|Session storage configuration. If omitted a cookie will be used.|
|auth.oidc.session.backend|string| |Session backend to use. Supports: cookie, redis|
|auth.oidc.session.redis|struct|{"host":""}|Settings for the Redis backend. Only used when backend is set to 'redis'.|
|auth.oidc.session.redis.host|string| |Host a Redis instance is accessible at. Set to 'redis.gloo-mesh.svc.cluster.local:6379' to use the included Redis deployment.|
|redis|struct|{"enabled":false}|Redis instance configuration, optionally used for auth session storage|
|redis.enabled|bool|false|Deploy a Redis instance for authentication|
|dashboard|struct|{"image":{"repository":"gloo-mesh-apiserver","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"sidecars":{"console":{"image":{"repository":"gloo-mesh-ui","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":null,"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}},"envoy":{"image":{"repository":"gloo-mesh-envoy","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"ENVOY_UID","value":"0"}],"resources":{"requests":{"cpu":"500m","memory":"256Mi"}}}},"floatingUserId":false,"runAsUser":10101,"serviceType":"ClusterIP","ports":{"console":8090,"grpc":10101,"healthcheck":8081}}|Configuration for the dashboard deployment.|
|dashboard|struct|{"image":{"repository":"gloo-mesh-apiserver","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}}||
|dashboard.image|struct|{"repository":"gloo-mesh-apiserver","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|dashboard.image.tag|string| |Tag for the container.|
|dashboard.image.repository|string|gloo-mesh-apiserver|Image name (repository).|
|dashboard.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|dashboard.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|dashboard.image.pullSecret|string| |Image pull secret.|
|dashboard.Env[]|slice|[{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|dashboard.resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|dashboard.resources.limits|map[string, struct]| ||
|dashboard.resources.limits.<MAP_KEY>|struct| ||
|dashboard.resources.limits.<MAP_KEY>|string| ||
|dashboard.resources.requests|map[string, struct]| ||
|dashboard.resources.requests.<MAP_KEY>|struct| ||
|dashboard.resources.requests.<MAP_KEY>|string| ||
|dashboard.resources.requests.cpu|struct|"125m"||
|dashboard.resources.requests.cpu|string|DecimalSI||
|dashboard.resources.requests.memory|struct|"256Mi"||
|dashboard.resources.requests.memory|string|BinarySI||
|dashboard.sidecars|map[string, struct]| |Configuration for the deployed containers.|
|dashboard.sidecars.<MAP_KEY>|struct| |Configuration for the deployed containers.|
|dashboard.sidecars.<MAP_KEY>.image|struct| |Specify the container image|
|dashboard.sidecars.<MAP_KEY>.image.tag|string| |Tag for the container.|
|dashboard.sidecars.<MAP_KEY>.image.repository|string| |Image name (repository).|
|dashboard.sidecars.<MAP_KEY>.image.registry|string| |Image registry.|
|dashboard.sidecars.<MAP_KEY>.image.pullPolicy|string| |Image pull policy.|
|dashboard.sidecars.<MAP_KEY>.image.pullSecret|string| |Image pull secret.|
|dashboard.sidecars.<MAP_KEY>.Env[]|slice| |Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|dashboard.sidecars.<MAP_KEY>.resources|struct| |Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|dashboard.sidecars.<MAP_KEY>.resources.limits|map[string, struct]| ||
|dashboard.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|struct| ||
|dashboard.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|string| ||
|dashboard.sidecars.<MAP_KEY>.resources.requests|map[string, struct]| ||
|dashboard.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|struct| ||
|dashboard.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|string| ||
|dashboard.sidecars.console|struct|{"image":{"repository":"gloo-mesh-ui","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":null,"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}}|Configuration for the deployed containers.|
|dashboard.sidecars.console.image|struct|{"repository":"gloo-mesh-ui","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|dashboard.sidecars.console.image.tag|string| |Tag for the container.|
|dashboard.sidecars.console.image.repository|string|gloo-mesh-ui|Image name (repository).|
|dashboard.sidecars.console.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|dashboard.sidecars.console.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|dashboard.sidecars.console.image.pullSecret|string| |Image pull secret.|
|dashboard.sidecars.console.Env[]|slice|null|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|dashboard.sidecars.console.resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|dashboard.sidecars.console.resources.limits|map[string, struct]| ||
|dashboard.sidecars.console.resources.limits.<MAP_KEY>|struct| ||
|dashboard.sidecars.console.resources.limits.<MAP_KEY>|string| ||
|dashboard.sidecars.console.resources.requests|map[string, struct]| ||
|dashboard.sidecars.console.resources.requests.<MAP_KEY>|struct| ||
|dashboard.sidecars.console.resources.requests.<MAP_KEY>|string| ||
|dashboard.sidecars.console.resources.requests.cpu|struct|"125m"||
|dashboard.sidecars.console.resources.requests.cpu|string|DecimalSI||
|dashboard.sidecars.console.resources.requests.memory|struct|"256Mi"||
|dashboard.sidecars.console.resources.requests.memory|string|BinarySI||
|dashboard.sidecars.envoy|struct|{"image":{"repository":"gloo-mesh-envoy","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"ENVOY_UID","value":"0"}],"resources":{"requests":{"cpu":"500m","memory":"256Mi"}}}|Configuration for the deployed containers.|
|dashboard.sidecars.envoy.image|struct|{"repository":"gloo-mesh-envoy","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|dashboard.sidecars.envoy.image.tag|string| |Tag for the container.|
|dashboard.sidecars.envoy.image.repository|string|gloo-mesh-envoy|Image name (repository).|
|dashboard.sidecars.envoy.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|dashboard.sidecars.envoy.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|dashboard.sidecars.envoy.image.pullSecret|string| |Image pull secret.|
|dashboard.sidecars.envoy.Env[]|slice|[{"name":"ENVOY_UID","value":"0"}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|dashboard.sidecars.envoy.resources|struct|{"requests":{"cpu":"500m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|dashboard.sidecars.envoy.resources.limits|map[string, struct]| ||
|dashboard.sidecars.envoy.resources.limits.<MAP_KEY>|struct| ||
|dashboard.sidecars.envoy.resources.limits.<MAP_KEY>|string| ||
|dashboard.sidecars.envoy.resources.requests|map[string, struct]| ||
|dashboard.sidecars.envoy.resources.requests.<MAP_KEY>|struct| ||
|dashboard.sidecars.envoy.resources.requests.<MAP_KEY>|string| ||
|dashboard.sidecars.envoy.resources.requests.cpu|struct|"500m"||
|dashboard.sidecars.envoy.resources.requests.cpu|string|DecimalSI||
|dashboard.sidecars.envoy.resources.requests.memory|struct|"256Mi"||
|dashboard.sidecars.envoy.resources.requests.memory|string|BinarySI||
|dashboard.floatingUserId|bool|false|Allow the pod to be assigned a dynamic user ID.|
|dashboard.runAsUser|uint32|10101|Static user ID to run the containers as. Unused if floatingUserId is 'true'.|
|dashboard.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|dashboard.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|dashboard.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|dashboard.ports.console|uint32|8090|Specify service ports as a map from port name to port number.|
|dashboard.ports.grpc|uint32|10101|Specify service ports as a map from port name to port number.|
|dashboard.ports.healthcheck|uint32|8081|Specify service ports as a map from port name to port number.|
|dashboard.DeploymentOverrides|invalid| |Provide arbitrary overrides for the component's [deployment template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)|
|dashboard.ServiceOverrides|invalid| |Provide arbitrary overrides for the component's [service template](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/).|
