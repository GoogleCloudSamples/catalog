`operation.promise()` takes no arguments, and polling does not begin until
`operation.promise()` is called. Setting
`operation.backoffSettings.totalTimeoutMillis` before calling `promise()`
bounds the wait while keeping the library's default polling intervals.
