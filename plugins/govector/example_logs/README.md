# Sample GoVector Logs

This directory contains sample GoVector logs generated from the [leaf](https://github.com/Blueprint-uServices/blueprint/tree/main/examples/leaf) application variant generated from the `govector` wiring specification.

## Individual Log Files

+ `leaf_process.goveclogger-Log.txt`: Log file generated from the process containing the leaf service instance.
+ `nonleaf_process.goveclogger-Log.txt`: Log file generated from the process containing the nonleaf service instance.
+ `shiviz_all_services.log`: [ShiViz](https://bestchai.bitbucket.io/shiviz/)-compatible log file generated using the [GoVector](https://github.com/DistributedClocks/GoVector) command line tool.

## Log Entry Format

Each log entry follows the following format:

```
$process.id {$process.vector_clock}
$log.message
```

Here is a detailed explanation for each field in the log entry:

+ `$process.id`: corresponds to the unique ID of the process. The `id` for each process is generated as `process_name.goveclogger`. For example, the `id` for a process with name `p1` would be `p1.goveclogger`.
+ `$process.vector_clock`: string representation of the vector clock timestamp at the time the log entry was generated. The vector clock is represented as a tuple of key-value pairs with each process's id as the key and its locally available logical time as the value.
+ `$log.message`: the message accompanying the log entry.
