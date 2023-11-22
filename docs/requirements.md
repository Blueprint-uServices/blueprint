# Requirements

## Blueprint pre-requisites

Blueprint requires golang 1.20 or higher.

## Plugin requirements

### Plugin compile-time requirements

Some plugins have additional requirements necessary to compile applications that use those plugins.  These requirements are optional; they are only needed if the plugin is used.  Examples include:
 - [gRPC plugin](../plugins/grpc) requires that the protobuf and gRPC compiler are installed
 - [Thrift plugin](../plugins/thrift/) requires that Thrift is installed

Plugins are expected to document their requirements.

### Plugin runtime requirements

Some plugins do not have compilation-time requirements but do have runtime requirements.  These requirements are optional; they are only needed if the plugin is used.  Examples include:
 - [docker-compose plugin](../plugins/dockerdeployment/) requires that Docker is installed in order to run docker containers

Plugins are expected to document their requirements.