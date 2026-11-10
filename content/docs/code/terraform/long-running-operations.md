The Google provider waits for most long-running operations to complete
before Terraform continues. Some resources aren't usable as soon as their
operation completes. When a newly enabled API or an IAM change needs time to
propagate, and requests fail with a 4xx error such as `403` until it does,
gate downstream resources on a `time_sleep` resource. When a resource exists
but isn't serving yet, and connections are refused or return `5xx` until it
is, gate them on a `data "http"` probe with a `retry` block.
