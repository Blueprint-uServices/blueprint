# Leaf Application - Wiring Spec

`main.go` defines the wiring spec for the Leaf application.  It instantiates the leaf and nonleaf services, applies some modifiers, and deploys them in separate processes with GRPC for communication

From this directory, run
```
go run main.go
```

Doing so will print out the IR of the application.