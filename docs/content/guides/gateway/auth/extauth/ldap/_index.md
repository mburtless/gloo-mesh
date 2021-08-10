---
title: LDAP
weight: 60
description: Authenticate and authorize requests using LDAP.
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

The _Lightweight Directory Access Protocol_, commonly referred to as LDAP, is an open protocol used to store and retrieve 
hierarchically structured data over a network. It has been widely adopted by enterprises to centrally store and secure 
organizational information. A common use case for LDAP is to maintain information about members of an organization, 
assign them to specific user groups, and give each of them access to resources based on their group memberships.

In this guide we will deploy a simple LDAP server to your Kubernetes cluster and see how you can use it together with 
Gloo Mesh to authenticate users and control access to a target service based on the user's group memberships.

{{% notice note %}}
We recommend that you check out [**this excellent tutorial**](https://www.digitalocean.com/community/tutorials/understanding-the-ldap-protocol-data-hierarchy-and-entry-components) 
by Digital Ocean to familiarize yourself with the basic concepts and components of an LDAP server; although it is not 
strictly necessary, it will help you better understand this guide.
{{% /notice %}}


## Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`). Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

#### Creating a Virtual Gateway
Now let's configure Gloo Mesh Gateway to route requests to the destination we just created. To do that, we define a simple Virtual
Gateway to match all requests that:

- contain a `Host` header with value `www.example.com` and
- have a path that starts with `/ratings`

{{< readfile file="guides/gateway/auth/extauth/basic_auth/test-no-auth-vg.yaml" markdown="true">}}

Let's send a request that matches the above route to the Gloo Mesh gateway and make sure it works:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

The above command should return:

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

#### Deploy an LDAP server
We also need to deploy an LDAP server to your cluster and configure it with a simple set of users and groups. This 
information will be used to determine which requests can access the destination we just defined. 

We have prepared a [**shell script**](setup-ldap.sh) that takes care of setting up the necessary resources. It creates:

1. a `configmap` with the LDAP server bootstrap configuration
2. a `deployment` running OpenLDAP
3. a `service` fronting the deployment
 
The script accepts an optional string argument, which determines the namespace in which the resources will be created 
(`default` if not provided). After you have downloaded the script to your working directory, you can run the following 
commands to execute it:

```shell
chmod +x setup-ldap.sh
./setup-ldap.sh    

No namespace provided, using default namespace
Creating configmap with LDAP server bootstrap config...
configmap/ldap created
Creating LDAP service and deployment...
deployment.apps/ldap created
service/ldap created
```

{{% expand "The details of the script are beyond the scope of this guide; if you are interested, you can inspect them by clicking on this paragraph." %}}
```bash
{{< readfile file="guides/gateway/auth/extauth/ldap/setup-ldap.sh" >}}
```
{{% /expand %}}

To understand the user configuration, it is worth looking at the last two data entries in the config map:

```text
03_people.ldif: |
  # Create a parent 'people' entry
  dn: ou=people,dc=solo,dc=io
  objectClass: organizationalUnit
  ou: people
  description: All solo.io people

  # Add 'marco'
  dn: uid=marco,ou=people,dc=solo,dc=io
  objectClass: inetOrgPerson
  cn: Marco Schmidt
  sn: Schmidt
  uid: marco
  userPassword: marcopwd
  mail: marco.schmidt@solo.io

  # Add 'rick'
  dn: uid=rick,ou=people,dc=solo,dc=io
  objectClass: inetOrgPerson
  cn: Rick Ducott
  sn: Ducott
  uid: rick
  userPassword: rickpwd
  mail: rick.ducott@solo.io

  # Add 'scottc'
  dn: uid=scottc,ou=people,dc=solo,dc=io
  objectClass: inetOrgPerson
  cn: Scott Cranton
  sn: Cranton
  uid: scottc
  userPassword: scottcpwd
  mail: scott.cranton@solo.io
04_groups.ldif: |+
  # Create top level 'group' entry
  dn: ou=groups,dc=solo,dc=io
  objectClass: organizationalUnit
  ou: groups
  description: Generic parent entry for groups

  # Create the 'developers' entry under 'groups'
  dn: cn=developers,ou=groups,dc=solo,dc=io
  objectClass: groupOfNames
  cn: developers
  description: Developers group
  member: uid=marco,ou=people,dc=solo,dc=io
  member: uid=rick,ou=people,dc=solo,dc=io
  member: uid=scottc,ou=people,dc=solo,dc=io

  # Create the 'sales' entry under 'groups'
  dn: cn=sales,ou=groups,dc=solo,dc=io
  objectClass: groupOfNames
  cn: sales
  description: Sales group
  member: uid=scottc,ou=people,dc=solo,dc=io

  # Create the 'managers' entry under 'groups'
  dn: cn=managers,ou=groups,dc=solo,dc=io
  objectClass: groupOfNames
  cn: managers
  description: Managers group
  member: uid=rick,ou=people,dc=solo,dc=io
```

We can see that the root of the LDAP directory hierarchy is the `dc=solo,dc=io` entry, which has two child entries:

- `ou=groups,dc=solo,dc=io` is the parent entry for user groups in the organization. It contains three groups:
    - cn=`developers`,ou=groups,dc=solo,dc=io
    - cn=`sales`,ou=groups,dc=solo,dc=io
    - cn=`managers`,ou=groups,dc=solo,dc=io
    
- `ou=people,dc=solo,dc=io` is the parent entry for people in the organization and in turn has the following entries:
   - uid=`marco`,ou=people,dc=solo,dc=io
   - uid=`rick`,ou=people,dc=solo,dc=io
   - uid=`scott`,ou=people,dc=solo,dc=io
   
The user credentials and memberships are summarized in the following table:

|  username |   password   | member of developers | member of sales | member of managers |
|-----------|--------------|----------------------|-----------------|--------------------|
| marco     | marcopwd     | Y                    |  N              |   N               |
| rick      | rickpwd      | Y                    |  N              |   Y               |
| scott     | scottpwd     | Y                    |  Y              |   N               |

To test that the LDAP server has been correctly deployed, let's port-forward the corresponding deployment:

```bash
kubectl port-forward deployment/ldap 8088:389
```

In a different terminal instance, run the following command (you must have `ldapsearch` installed):

```bash
ldapsearch -H ldap://localhost:8088 -D "cn=admin,dc=solo,dc=io" -w "solopwd" -b "dc=solo,dc=io" -LLL dn
```

You should see the following output, listing the **distinguished names (DNs)** of all entries located in the subtree 
rooted at `dc=solo,dc=io`:

```text
dn: dc=solo,dc=io

dn: cn=admin,dc=solo,dc=io

dn: ou=people,dc=solo,dc=io

dn: uid=marco,ou=people,dc=solo,dc=io

dn: uid=rick,ou=people,dc=solo,dc=io

dn: uid=scottc,ou=people,dc=solo,dc=io

dn: ou=groups,dc=solo,dc=io

dn: cn=developers,ou=groups,dc=solo,dc=io

dn: cn=sales,ou=groups,dc=solo,dc=io

dn: cn=managers,ou=groups,dc=solo,dc=io
```

### Secure the Virtual Gateway

Now that we have all the necessary components in place, let use the LDAP server to secure the Virtual Gateway we created 
earlier.

#### LDAP auth flow
Before updating our Virtual Gateway, it is important to understand how Gloo Mesh interacts with the LDAP server. Let's first 
look at the `Ldap` configuration:

- `address`: this is the address of the LDAP server that Gloo Mesh will query when a request matches the Virtual Gateway.
- `userDnTemplate`: this is a template string that Gloo Mesh uses to build the DNs of the user entry that 
   needs to be authenticated and authorized. It must contains a single occurrence of the “%s” placeholder.
- `membershipAttributeName`: case-insensitive name of the attribute that contains the names of the groups an entry is 
   member of. Defaults to `memberOf` if not provided.
- `allowedGroups`: DNs of the user groups that are allowed to access the secured destination.

To better understand how this configuration is used, let's go over the steps that Gloo Mesh performs when it detects a 
request that needs to be authenticated with LDAP:

1. Look for a [Basic Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) header on the request 
   and extract the username and credentials
2. If the header is not present, return a `401` response
3. Try to perform a [BIND](https://ldap.com/the-ldap-bind-operation/) operation with the LDAP server. To do this, Gloo Mesh 
   needs to know the DN of the user entry. It will build it by substituting the name of the user (extracted from the 
   basic auth header) for the `%s` placeholder in the `userDnTemplate`. It is important to note that 
   [special characters](https://ldapwiki.com/wiki/DN%20Escape%20Values) will be removed from the username before performing 
   the bind operation; this is done to prevent injection attacks.
4. If the operation fails, it means that the user is unknown or their credentials are incorrect; return a `401` response
5. Issue a search operation for the user entry (with a [`base` scope](https://ldapwiki.com/wiki/BaseObject)) and look 
   for an attribute with a name equal to `membershipAttributeName` on the user entry.
6. Check if one of the values for the attribute matches one of the `allowedGroups`; if so, allow the request, otherwise return a `403` response.

#### Create an LDAP AuthConfig
Now that we have a good understanding of how Gloo Mesh interacts with the LDAP server we can create an `AuthConfig` CRD with 
our LDAP configuration:

{{< highlight yaml "hl_lines=9-12" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: ldap
  namespace: bookinfo
spec:
  configs:
  - ldap:
      address: "ldap://ldap.default.svc.cluster.local:389" # Substitute your namespace for `default` here
      userDnTemplate: "uid=%s,ou=people,dc=solo,dc=io"
      allowedGroups:
      - "cn=managers,ou=groups,dc=solo,dc=io"
{{< /highlight >}}

We can see that:

- the configuration points to the Kubernetes DNS name and port of our LDAP service (`ldap.default.svc.cluster.local:389` if 
  you deployed it to the `default` namespace);
- Gloo Mesh will look for user entries with DNs in the format `uid=<USERNAME_FROM_HEADER>,ou=people,dc=solo,dc=io`, which, 
  if you recall, is the format of the user entry DNs we bootstrapped our server with;
- only members of the `cn=managers,ou=groups,dc=solo,dc=io` group can access the destination.

#### Update the Virtual Gateway
Once the `AuthConfig` containing the LDAP configuration has been created, we can use it to secure our Virtual Gateway 
by applying the following:

{{< readfile file="guides/gateway/auth/extauth/ldap/test-ldap-auth-vg.yaml" markdown="true">}}


This configures the Virtual Gateway to authenticate all requests to `/echo` using using the configuration stored in the 
`AuthConfig` CRD named `ldap` in the `gloo-system` namespace.

Let's verify that our Virtual Gateway behaves as expected. The basic auth header requires credentials to be encoded, 
so here are the `base64`-encoded credentials for some test users:

| username | password | basic auth header                         | comments                                    |
|----------|----------|-------------------------------------------|---------------------------------------------|
| marco    | marcopwd | Authorization: Basic bWFyY286bWFyY29wd2Q= | Member of "developers" group                |
| rick     | rickpwd  | Authorization: Basic cmljazpyaWNrcHdk     | Member of "developers" and "managers" group |
| john     | doe      | Authorization: Basic am9objpkb2U=         | Unknown user                                |

##### No auth header
To start with, let's send a request without any header:

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

returns

{{< highlight bash "hl_lines=9" >}}
*   Trying 192.168.99.100...
* TCP_NODELAY set
* Connected to 192.168.99.100 (192.168.99.100) port 31940 (#0)
> GET /echo HTTP/1.1
> Host: 192.168.99.100:31940
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 401 Unauthorized
< date: Tue, 10 Sep 2019 17:14:39 GMT
< server: envoy
< content-length: 0
<
* Connection #0 to host 192.168.99.100 left intact
{{< /highlight >}}

We can see that Gloo Mesh returned a `401` response.

##### Unknown user
Now let's try the unknown user, which will produce the same result:

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: Basic am9objpkb2U="
```

returns

{{< highlight bash "hl_lines=10" >}}
*   Trying 192.168.99.100...
* TCP_NODELAY set
* Connected to 192.168.99.100 (192.168.99.100) port 31940 (#0)
> GET /echo HTTP/1.1
> Host: 192.168.99.100:31940
> User-Agent: curl/7.54.0
> Accept: */*
> Authorization: Basic am9objpkb2U=
>
< HTTP/1.1 401 Unauthorized
< date: Tue, 10 Sep 2019 17:25:21 GMT
< server: envoy
< content-length: 0
<
* Connection #0 to host 192.168.99.100 left intact
{{< /highlight >}}

##### Developer user
If we try to authenticate as a user that belongs to the "developers" group, Gloo Mesh will return a `403` response, 
indicating that the user was successfully authenticated, but lacks the permissions to access the resource.

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: Basic bWFyY286bWFyY29wd2Q="
```

returns

{{< highlight bash "hl_lines=10" >}}
*   Trying 192.168.99.100...
* TCP_NODELAY set
* Connected to 192.168.99.100 (192.168.99.100) port 31940 (#0)
> GET /echo HTTP/1.1
> Host: 192.168.99.100:31940
> User-Agent: curl/7.54.0
> Accept: */*
> Authorization: Basic bWFyY286bWFyY29wd2Q=
>
< HTTP/1.1 403 Forbidden
< date: Tue, 10 Sep 2019 17:29:12 GMT
< server: envoy
< content-length: 0
<
* Connection #0 to host 192.168.99.100 left intact
{{< /highlight >}}

##### Manager user
Finally, if we provide a user that belongs to the "managers" group, we will be able to access the destination.

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: Basic cmljazpyaWNrcHdk"
```

returns

{{< highlight bash "hl_lines=10 19" >}}
*   Trying 192.168.99.100...
* TCP_NODELAY set
* Connected to 192.168.99.100 (192.168.99.100) port 31940 (#0)
> GET /echo HTTP/1.1
> Host: 192.168.99.100:31940
> User-Agent: curl/7.54.0
> Accept: */*
> Authorization: Basic cmljazpyaWNrcHdk
>
< HTTP/1.1 200 OK
< x-app-name: http-echo
< x-app-version: 0.2.3
< date: Tue, 10 Sep 2019 17:30:12 GMT
< content-length: 15
< content-type: text/plain; charset=utf-8
< x-envoy-upstream-service-time: 0
< server: envoy
<
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
* Connection #0 to host 192.168.99.100 left intact
{{< /highlight >}}

### Summary 
In this tutorial we have shown how Gloo Mesh can integrate with LDAP to authenticate incoming requests and authorize them based 
on the group memberships of the user associated with the request credentials.

To clean up the resources we created, you can run the following commands:

```bash
kubectl delete configmap ldap
kubectl delete deployment ldap
kubectl delete service ldap
```
