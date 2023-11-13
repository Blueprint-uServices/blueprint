#!/bin/bash

HOME_DIR=$PWD
dirs=(
# Core packages
    "blueprint/pkg/blueprint/ioutil"
    "blueprint/pkg/blueprint/logging"
    "blueprint/pkg/coreplugins/address"
    "blueprint/pkg/coreplugins/backend"
    "blueprint/pkg/coreplugins/pointer"
    "blueprint/pkg/coreplugins/service"
    "blueprint/pkg/ir/"
    "blueprint/pkg/wiring/"
# Plugin packages
    "plugins/circuitbreaker"
    "plugins/clientpool"
    "plugins/docker"
    "plugins/dockerdeployment"
    "plugins/golang"
    "plugins/goproc"
    "plugins/grpc"
    "plugins/healthchecker"
    "plugins/http"
    "plugins/jaeger"
    "plugins/linux"
    "plugins/linuxcontainer"
    "plugins/memcached"
    "plugins/mongodb"
    "plugins/opentelemetry"
    "plugins/redis"
    "plugins/retries"
    "plugins/simplecache"
    "plugins/simplenosqldb"
    "plugins/thrift"
    "plugins/workflow"
    "plugins/workload"
    "plugins/xtrace"
    "plugins/zipkin"
# Runtime packages
    "runtime/core/backend"
    "runtime/plugins/clientpool"
    "runtime/plugins/golang"
    "runtime/plugins/jaeger"
    "runtime/plugins/memcached"
    "runtime/plugins/mongodb"
    "runtime/plugins/opentelemetry"
    "runtime/plugins/redis"
    "runtime/plugins/simplecache"
    "runtime/plugins/simplenosqldb"
    "runtime/plugins/xtrace"
    "runtime/plugins/zipkin"
)

for dir in "${dirs[@]}"
do
    cd $dir
    title=$(echo "$dir" | tr '/' '_')
    outfile=$HOME_DIR/docs/$title.md
    echo "Generating documentation for $dir"
    echo "---" > $outfile
    echo "title: $dir" >> $outfile
    echo "---" >> $outfile
    echo "# $dir" >> $outfile
    go doc -all | godoc2markdown >> $outfile
    cd $HOME_DIR
done
