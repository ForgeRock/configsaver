# Config Saver - Saves Forgeops Platform Configuration

This is a POC / work in progress

The config saver is a client/server protocol enabling a client sidecar to receive
and save product configuration from the configsaver server.

The protocol is defined in [configsaver.proto](proto/configsaver.proto), and consists of
two gRPC calls:
* GetConfig - gets the full product configuration from the server. A tar ball with the
 full configuration is returned.
* UpdateConfig   - updates the product configuration on the server. The update is
  a tarball of the full or partial configuraition changes to be saved by the server.


The server currently performs a git clone of an upstream repo (default, forgeops). When deployed, the server repo
should be saved to PVC running in the namespace of the deployment. This provides persistence
of configuration changes, until the changes are pushed to an upstream repo. Client sidecars in AM, IDM
and IG will contact the server via gRPC.

The client is "dumb" in that it does not use git to track configuration changes. The client
looks at the timestamps of the files, and sends any changed files to the config saver. It is
up to the server to decide how and where to save the files. It is possible there are no
configuration changes (for example, a timestamp changed, but the contents did not), in which case the server may ignore the
updates. The intent of a "dumb client / smart server" design is to allow for future changes in how files are saved by the server.


## TODO:

* Create docker images for the server and the client
* Create a K8S deployment and sidecars
* Implement creating upstream tracking branches (e.g. autosave). Currently the server exits if the upstream branch does not exist.
* Implement push to upstream branch
* More flexible options for configuring the server. For example
   the paths to the product configurations under the repo are hard coded. This should be configurable.
* The server could perform some configuration validation. For example, ensuring json is well formed.
* The server could perform replacement of hard coded values with commons expressions.

## Notes

* gRPC limits payload size to 4MB. The tar files are much smaller than this limit.
   The tar bytes can be compressed if space saving is required in the future.
