// Package runtime contains interfaces and implementations that are used by compiled Blueprint applications at runtime
//
// During compilation, Blueprint's compiler will take code from this module and include it into the compiled output.
// This is primarily implemented by the [golang/goparser] and [workflowspec] plugins
//
// [runtime/core/backend] defines the backend interfaces used by workflow specs.  A workflow spec implementation might want to import that package and use the interfaces defined there.
//
// [plugins] defines implementations that are used by plugins and automatically compiled into the application.  These should not need to be directly reference by workflow specs.
//
// [golang/goparser]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/goparser
// [workflowspec]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/workflowspec
// [runtime/core/backend]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/core/backend
// [plugins]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins
package runtime
