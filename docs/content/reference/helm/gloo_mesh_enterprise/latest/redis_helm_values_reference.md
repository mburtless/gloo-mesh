
---
title: "Redis"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|redisDashboard|struct|{"image":{"repository":"redis","registry":"docker.io","pullPolicy":"IfNotPresent"},"env":[{"name":"MASTER","value":"true"}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"sidecars":{},"floatingUserId":false,"runAsUser":10101,"serviceType":"ClusterIP","ports":{"redis":6379},"enabled":true}|Configuration for the redisDashboard deployment.|
|redisDashboard|struct|{"image":{"repository":"redis","registry":"docker.io","pullPolicy":"IfNotPresent"},"env":[{"name":"MASTER","value":"true"}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}}||
|redisDashboard.image|struct|{"repository":"redis","registry":"docker.io","pullPolicy":"IfNotPresent"}|Specify the container image|
|redisDashboard.image.tag|string| |Tag for the container.|
|redisDashboard.image.repository|string|redis|Image name (repository).|
|redisDashboard.image.registry|string|docker.io|Image registry.|
|redisDashboard.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|redisDashboard.image.pullSecret|string| |Image pull secret.|
|redisDashboard.Env[]|slice|[{"name":"MASTER","value":"true"}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|redisDashboard.resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|redisDashboard.resources.limits|map[string, struct]| ||
|redisDashboard.resources.limits.<MAP_KEY>|struct| ||
|redisDashboard.resources.limits.<MAP_KEY>|string| ||
|redisDashboard.resources.requests|map[string, struct]| ||
|redisDashboard.resources.requests.<MAP_KEY>|struct| ||
|redisDashboard.resources.requests.<MAP_KEY>|string| ||
|redisDashboard.resources.requests.cpu|struct|"125m"||
|redisDashboard.resources.requests.cpu|string|DecimalSI||
|redisDashboard.resources.requests.memory|struct|"256Mi"||
|redisDashboard.resources.requests.memory|string|BinarySI||
|redisDashboard.sidecars|map[string, struct]| |Configuration for the deployed containers.|
|redisDashboard.sidecars.<MAP_KEY>|struct| |Configuration for the deployed containers.|
|redisDashboard.sidecars.<MAP_KEY>.image|struct| |Specify the container image|
|redisDashboard.sidecars.<MAP_KEY>.image.tag|string| |Tag for the container.|
|redisDashboard.sidecars.<MAP_KEY>.image.repository|string| |Image name (repository).|
|redisDashboard.sidecars.<MAP_KEY>.image.registry|string| |Image registry.|
|redisDashboard.sidecars.<MAP_KEY>.image.pullPolicy|string| |Image pull policy.|
|redisDashboard.sidecars.<MAP_KEY>.image.pullSecret|string| |Image pull secret.|
|redisDashboard.sidecars.<MAP_KEY>.Env[]|slice| |Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|redisDashboard.sidecars.<MAP_KEY>.resources|struct| |Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|redisDashboard.sidecars.<MAP_KEY>.resources.limits|map[string, struct]| ||
|redisDashboard.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|struct| ||
|redisDashboard.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|string| ||
|redisDashboard.sidecars.<MAP_KEY>.resources.requests|map[string, struct]| ||
|redisDashboard.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|struct| ||
|redisDashboard.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|string| ||
|redisDashboard.floatingUserId|bool|false|Allow the pod to be assigned a dynamic user ID.|
|redisDashboard.runAsUser|uint32|10101|Static user ID to run the containers as. Unused if floatingUserId is 'true'.|
|redisDashboard.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|redisDashboard.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|redisDashboard.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|redisDashboard.ports.redis|uint32|6379|Specify service ports as a map from port name to port number.|
|redisDashboard.DeploymentOverrides|invalid| |Provide arbitrary overrides for the component's [deployment template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)|
|redisDashboard.ServiceOverrides|invalid| |Provide arbitrary overrides for the component's [service template](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/).|
|redisDashboard.enabled|bool|true|Enables or disables creation of the operator deployment/service|
