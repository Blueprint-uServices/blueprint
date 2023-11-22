# Sockshop Microservices

This is a Blueprint re-implementation / translation of the [SockShop microservices demo](https://microservices-demo.github.io).

For the most part, the application directly re-uses the original SockShop code (for services that were written in Go) or does a mostly-direct translation of code (for services that were not written in Go).  Some aspects of the application (such as HTTP URLs) were tweaked from the original implementation, but in terms of APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow] contains service implementations
* [tests] has tests of the workflow
* [wiring] configures the application's topology and deployment