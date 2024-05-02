module github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workload

go 1.21

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

replace github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow => ../workflow

require github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow v0.0.0

require github.com/blueprint-uservices/blueprint/runtime v0.0.0