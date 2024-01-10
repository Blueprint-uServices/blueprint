# Requirements

Different plugins have different pre-requisites; you only need to install the pre-requisites of a plugin if you intend to use it.  That said, we recommend installing the following:

 * golang 1.20 or higher
 * Docker
 * Kubernetes
 * gRPC

## Blueprint pre-requisites

Blueprint requires golang 1.20 or higher.

## Plugin requirements

See the respective plugin documentation in [plugins](../../plugins/) for individual plugin requirements.

### Plugin compile-time requirements

Some plugins have additional requirements necessary to compile applications that use those plugins.  These requirements are optional; they are only needed if the plugin is used.  Examples include:
 - [gRPC plugin](../plugins/grpc) requires that the protobuf and gRPC compiler are installed
 - [Thrift plugin](../plugins/thrift/) requires that Thrift is installed

Plugins are expected to document their requirements.

### Plugin runtime requirements

Some plugins do not have compilation-time requirements but do have runtime requirements.  These requirements are optional; they are only needed if the plugin is used.  Examples include:
 - [docker-compose plugin](../plugins/dockerdeployment/) requires that Docker is installed in order to run docker containers

Plugins are expected to document their requirements.