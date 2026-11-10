`getName()` waits for the initial RPC that starts the operation, after which
`peekMetadata()` returns the initial metadata without blocking. If the initial
RPC fails, `getName()` throws `ExecutionException` wrapping the underlying
`ApiException` (such as `AlreadyExistsException`). If the operation fails on
the server, `operation.get(...)` throws `ExecutionException` in the same way.
