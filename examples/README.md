# Examples

This directory contains some example Blueprint applications, primarily for use in getting started with Blueprint.  The full range of Blueprint applications are listed at [TODO]().

- [leaf](leaf) is a two-service application where one service calls the other
- [sockshop](sockshop) re-implements the SockShop microservices benchmark
- [dsb_hotel](dsb_hotel) re-implements the hotel-reservation application from the DeathStarBench microservices benchmark
- [dsb_sn](dsb_sn) re-implements the social-network application from the DeathStarBench microservices benchmark
- [train_ticket](train_ticket) re-implements the TrainTicket microservices benchmark

### Related documentation
- [../docs/manual/workflow.md](docs/manual/workflow.md) describes how to write a Blueprint application
- [../docs/manual/wiring.md](docs/manual/wiring.md) describes creating a wiring spec for an application, which will make use of plugins
- [../docs/manual/plugins.md](docs/manual/plugins.md) gives an overview of the available plugins and their functionality

### Other modules in this repository
- [../blueprint](../blueprint) implements Blueprint's compiler as well as the [WiringSpec API](../blueprint/pkg/wiring) used by Blueprint applications.
- [../plugins](../plugins) contains a range of plugins that implement most of Blueprint's features and functionality
- [../runtime](../runtime) contains runtime code that is used by Blueprint applications and by plugins