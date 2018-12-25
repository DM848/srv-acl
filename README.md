# srv-acl

The ACL service works as an entrypoint for our platform. Multiple instances runs behind an external google cloud load balancer to reduce any downtime.

## Goal
Verify incoming requests to be authenticated (JWT) and that they have the correct permission level to access their desired service. Because of this, the ACL also keeps track with service discovery and when a new service is reported by Consul with the tag "platform-endpoint"; it is automatically given an endpoint at `/api/<service-name>`.

To see the current configuration for all the endpoints and the default roles/permission levels, visit `/configuration`.

## User jolie scripts
In lack of a better terminology, this refers to the jolie scripts deployed by users through the Jolie-deployer. These are identified through the tag `user-endpoint` and their token fetched from the token tag `token:<token>`. When one of these are registerred as a service with Consul, the ACL service creates an endpoint for them at `/script/<token>`. This can be accessed by anyone, and the user themselves are responsible for authentication and restricting access.

See the repository srv-cloud-server to understand how the tags are created and used.
