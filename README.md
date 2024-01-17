<p align="center">
  <img src="https://blueprint-uservices.github.io/assets/img/blueprint%20logo%204 small.png" width=200/>
</p>

Blueprint is an extensible compiler for microservice applications, along with a collection of off-the-shelf microservice benchmark applications.

Using Blueprint, you can:
 * Run a number of off-the-shelf microservice applications and use them for your benchmarking, prototyping, and research experimentation.
 * Change and reconfigure those off-the-shelf applications, so that they use different frameworks, infrastructure, and deployment choices.
 * Easily develop your own microservice applications and leverage Blueprint's built-in features that decouple your application's logic from (Blueprint-generated) boilerplate.
 * Prototype and develop your own microservice *infrastructure*, and integrate and evaluate your infrastructure with all of the existing applications.

Blueprint is particularly aimed at **prototyping and experimentation** use cases.  It is intended for use by anybody, but particularly researchers and practitioners wanting to experiment with microservice applications.  Its central goal is to reduce the amount of effort involved when changing and re-compiling the infrastructure choices of a microservice application.  

## Status

Blueprint is currently pre-release with most initial features implemented.  However, some features may not be implemented yet, some applications may be incomplete, and some aspects of the user guide may be empty.  Please join the [blueprint-uservices](https://blueprint-uservices.slack.com/) Slack if you have any questions, requests, or comments.

## Getting Started

See the 📖[Getting Started](docs/manual/gettingstarted.md) page of the User Manual to get started compiling and running your first Blueprint application.

## Resources

📘[User Manual](docs/manual)

📚[Applications](examples) - off-the-shelf applications that come packaged with Blueprint\
📝[Wiring Spec Plugins](plugins) - plugins that come packaged with Blueprint that can be used in wiring specs to modify applications\
📓[Workflow Spec Backends](runtime/core) - backend interfaces that can be used in workflow specs when developing applications

### API Documentation on go.dev

🚀[Blueprint Compiler](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/blueprint)\
🚀[Blueprint Plugins](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/plugins)\
🚀[Blueprint Runtime Components](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/runtime)

## Project

If you anticipate making use of Blueprint for your research project, we recommend familiarizing yourself with the SOSP 2023 publication below, which outlines and demonstrates some motivating use cases for Blueprint.

📄[Blueprint: A Toolchain for Highly-Reconfigurable Microservices](https://blueprint-uservices.github.io/assets/pdf/anand2023blueprint.pdf)\
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Vaastav Anand, Deepak Garg, Antoine Kaufmann, Jonathan Mace\
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*29th ACM Symposium on Operating Systems Principles (SOSP '23)*

If you are looking for the artifact that was evaluated for the SOSP paper, it can be found [here](https://gitlab.mpi-sws.org/cld/blueprint), though it is not maintained.  Instead we recommend using the Blueprint implementation contained in this repository.

### Mailing List & Slack

 * Slack: [blueprint-uservices](https://blueprint-uservices.slack.com/)
 * Google Group (mailing list): [blueprint-uservices](https://groups.google.com/g/blueprint-uservices)

### Contributors

We are a team of researchers:
 * [Vaastav Anand](https://vaastavanand.com/), PhD student at the Max Planck Institute for Software Systems (MPI-SWS)
 * [Jonathan Mace](https://www.microsoft.com/en-us/research/people/jonathanmace/), Researcher at Microsoft Research and Adjunct Faculty at the Max Planck Institute for Software Systems (MPI-SWS)

If you are interested in contributing, please contact us on Slack!
