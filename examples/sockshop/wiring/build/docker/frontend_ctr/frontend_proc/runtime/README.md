# Runtime

This module contains interfaces and implementations that are used by Blueprint applications at runtime, in the compiled applications.  During compilation, Blueprint's compiler will take code from this module and include it into the compiled output.

[runtime/core/backend](core/backend) defines the backend interfaces used by workflow specs.  A workflow spec implementation might want to import that package and use the interfaces defined there.

[plugins](plugins) defines implementations that are used by plugins and automatically compiled into the application.  These should not need to be directly reference by workflow specs.