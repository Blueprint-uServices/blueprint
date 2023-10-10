# Test

Contains wiring specs to test plugins produce the correct IR

Tests generally contain a workflow spec and a wiring spec, which is why they reside here instead of with plugin logic (for now)

Tests for wiring specs live in the 'wiring' folder.  Workflow specs used by tests live in workflow.

```
cd wiring
go test
```