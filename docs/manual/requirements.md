# Requirements

In order to compile a Blueprint application, the build machine must satisfy the pre-requisites listed here.  Note that these requirements apply to the machine on which you are compiling the application; they do not typically also apply to the machine(s) on which you run the application.

## Compiler Requirements

Blueprint requires golang 1.20 or higher.

We **highly recommend** also installing the following in order to run the Blueprint examples.  Follow the instructions under the **Prerequisites** heading for the following plugins:
 * [gRPC](../../plugins/grpc)
 * [Docker](../../plugins/docker)
 * [Kubernetes](../../plugins/kubernetes)

The above dependencies are sufficient for compiling most of the Blueprint example applications.  However, in addition to the above, some plugins might have further dependencies that you will need to install before you can use that plugin.  The plugin will document these dependencies.