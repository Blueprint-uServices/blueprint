# Blueprint

Blueprint is an extensible compiler for microservice applications, along with a collection of off-the-shelf microservice benchmark applications.

Using Blueprint, you can:
 * Run a number of off-the-shelf microservice applications and use them for your benchmarking, prototyping, and research experimentation.
 * Change and reconfigure those off-the-shelf applications, so that they use different frameworks, infrastructure, and deployment choices.
 * Easily develop your own microservice applications and leverage Blueprint's built-in features that decouple your application's logic from (Blueprint-generated) boilerplate.
 * Prototype and develop your own microservice *infrastructure*, and integrate and evaluate your infrastructure with all of the existing applications.

Blueprint is particularly aimed at **prototyping and experimentation** use cases.  It is intended for use by anybody, but particularly researchers and practitioners wanting to experiment with microservice applications.  Its central goal is to reduce the amount of effort involved when changing and re-compiling the infrastructure choices of a microservice application.  

## Getting Started

To get started compiling and running your first Blueprint application, see the 📖[Getting Started](docs/manual/gettingstarted.md) page of the User Manual.

## Resources

📘[User Manual](docs/manual) - the main source of Blueprint documentation\
📚[Applications](examples) - off-the-shelf applications that come packaged with Blueprint\
📝[Wiring Spec Plugins](plugins) - plugins that come packaged with Blueprint that can be used in wiring specs
📓[Workflow Spec Backends](runtime/core) - backend interfaces that can be used in workflow specs\

### API Documentation on go.dev

🚀[Blueprint Compiler](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/blueprint)\
🚀[Blueprint Plugins](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/plugins)\
🚀[Blueprint Runtime Components](https://pkg.go.dev/github.com/blueprint-uservices/blueprint/runtime)

## Project

If you anticipate making use of Blueprint for your research project, we recommend familiarizing yourself with the SOSP 2023 publication below, which outlines and demonstrates some motivating use cases for Blueprint.

📄[Blueprint: A Toolchain for Highly-Reconfigurable Microservices](https://blueprint-uservices.github.io/assets/pdf/anand2023blueprint.pdf)\
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Vaastav Anand, Deepak Garg, Antoine Kaufmann, Jonathan Mace\
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*29th ACM Symposium on Operating Systems Principles (SOSP '23)*

### Mailing List & Slack

 * Slack: [blueprint-uservices](https://blueprint-uservices.slack.com/)
 * Google Group (mailing list): [blueprint-uservices](https://groups.google.com/g/blueprint-uservices)

### Contributors

We are a team of researchers:
 * [Vaastav Anand](https://vaastavanand.com/), PhD student at the Max Planck Institute for Software Systems (MPI-SWS)
 * [Jonathan Mace](https://www.microsoft.com/en-us/research/people/jonathanmace/), Researcher at Microsoft Research and Adjunct Faculty at the Max Planck Institute for Software Systems (MPI-SWS)

If you are interested in contributing, please contact us on Slack!
