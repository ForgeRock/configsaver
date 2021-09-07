# Config Saver - Saves Forgeops Platform Configuration

This is a POC / work in progress

The config saver is a client/server protocol enabling a client sidecar to receive
and save product configuration from the configsaver server.

The protocol is defined in [configsaver.proto](proto/configsaver.proto), and consists of
two gRPC calls:
* GetConfig - gets the full product configuration from the server. A tar ball with the
 full configuration is returned.
* UpdateConfig   - updates the product configuration on the server. The update is
  a tarball of the full or partial configuration changes to be saved by the server.


The server currently performs a git clone of an upstream repo (default, forgeops). When deployed, the server repo
should be saved to PVC running in the namespace of the deployment. This provides persistence
of configuration changes, until the changes are pushed to an upstream repo. Client sidecars in AM, IDM
and IG will contact the server via gRPC.

The client is "dumb" in that it does not use git to track configuration changes. The client
looks at the timestamps of the files, and sends any changed files to the config saver. It is
up to the server to decide how and where to save the files. It is possible there are no
configuration changes (for example, a timestamp changed, but the contents did not), in which case the server may ignore the
updates. The intent of a "dumb client / smart server" design is to allow for future evolution in how files are saved by the server.


## TODO:

* Create a K8S deployment and sidecars [WIP in forgeops branch]
* Implement creating upstream tracking branches (e.g. autosave). Currently the server exits if the upstream branch does not exist.
* Implement push to upstream branch
* The server could perform some configuration validation. For example, ensuring json is well formed.
* The server could perform replacement of hard coded values with commons expressions.
* The product map that tells the server where to find am or idm config within the cloned repo is hard coded to forgeops/docker/. Consider
 making it configurable.

## Notes

* gRPC limits payload size to 4MB. The tar files are much smaller than this limit.
   The tar bytes can be compressed if space saving is required in the future.

## Developer Notes

To test the server and client:

```bash
# Run the server. The server files are in tmp/forgeops
make serve
# In another shell window

# Runs the client in a "one shot" mode. Client downloads config and exits
#  The files are downloaded to tmp/client
make client

# runs the client in sync mode. Client will watch the directory for changes, and upload results to the server
make  client-sync

# Try to change a file in tmp/client - you should see the file being updated in tmp/forgeops.  Note forgeops is a git repo
 and you can use git commands to see changes. Try `git status` and `git log`

```

## Environment Variables

* CONFIG_REPO - The git repo to clone as the source of configuration. Default is forgeops.
* CONFIG_DIR -  working directory where the server or client stores files.
* CONFIG_SERVER - the URL for the client to  connect to the server. Default is localhost:50051
* CONFIG_PRODUCT - the product the client is configuring (am or idm). This is passed to the server
 to help it locate the configuration within the cloned repo. Defaults to `am`
* GIT_SSH_PATH - path to git ssh credentials needed to clone a repo or to push changes. This is optional.
  If not provided, the repo should be public.