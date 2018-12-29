# srv-acl

The ACL service works as an entrypoint for our platform. Multiple instances runs behind an external google cloud load balancer to reduce any downtime.

## Behaviour
Every outgoing message from the ACL is json and uses the JSend protocol to standarise the output. How the content of the data field is structured, is up to the individual services. However, the services access through the ACL (public / endpoints) must return json valid content as this is handled in the ACL on responses. An exception to all of this, is the `/script` endpoint, as this handles traffic from user scripts which are not directly called from our platform.

## Goal
Verify incoming requests to be authenticated (JWT) and that they have the correct permission level to access their desired service. Because of this, the ACL also keeps track with service discovery and when a new service is reported by Consul with the tag "platform-endpoint"; it is automatically given an endpoint at `/api/<service-name>`.

To see the current configuration for all the endpoints and the default roles/permission levels, visit `/configuration`.

## User jolie scripts
In lack of a better terminology, this refers to the jolie scripts deployed by users through the Jolie-deployer. These are identified through the tag `user-endpoint` and their token fetched from the token tag `token:<token>`. When one of these are registerred as a service with Consul, the ACL service creates an endpoint for them at `/script/<token>`. This can be accessed by anyone, and the user themselves are responsible for authentication and restricting access.

See the repository srv-cloud-server to understand how the tags are created and used.

## Enforcing data values
As there might be a need to use auth values in the backend, and they cannot use the header fields, nor have a proper libraries to parse JWT: The ACL layer parses both body and GET query params in order to detect auth values and enforce their validity compared to the included JWT. If the JWT is missing, these values are reset with default zero values.
> NOTE! This feature can be turned off for development in the Consul KV storage: srv-acl_ACLEntry-config_enforce = false

Keys that are enforced:
 - acle_user_id: string
 - acle_user_level: int
 - acle_user_level_str: string

##### acle_user_level
Permission flags of a given user. Up to 32 flags. See permission.go for each one, and their int value.

##### acle_user_level_str
Only holds the minimum fulfilled static role (such as "usr", "dev", "adm"). This means that when the acle_user_level holds all the permission flags required by one or several static roles, the static role with highest permission flags is selected.
Note that users with a permission level below "usr" is named "nobody".