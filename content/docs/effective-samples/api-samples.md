---
date: '2026-01-29T22:09:43-08:00'
draft: false
title: 'Authoring Samples: API Operations'
description: 'Extra rules for API operation samples: running without edits, single use case, importability, and errors by verb.'
weight: 2
---

This section of the guide presents additional practices for a type of **{{< icon "scissors" >}} Code Snippet**. 

### Defining API Samples {#defining-api-samples}

**{{< icon "color-swatch" >}}API Sample:** A single piece of code showing the end-to-end steps of using an API operation. These are foundational building blocks and developers benefit from maximizing consistency.

{{< callout >}}API Samples benefit from more automation so there is additional capacity to add guidelines.
{{< /callout >}}

### Guidelines {#guidelines}

#### Single Use Case {#single-use-case}

Show a single use case in a single sample. Set only the request fields
this use case needs, and don't set optional fields to their default
values.

***Why?** Each sample answers a single question as clearly as possible. Other samples can address unrealized use cases, and reference documentation will address API operations that remain undocumented in samples.*

#### Support Importing Sample Code {#support-importing-sample-code}

The sample should be subject to import for use cases such as unit testing or invoking the sample from another package. This may include exporting symbols. Prefer exporting symbols outside region tags.

#### Correct Syntax within Region Tags {#correct-syntax-within-region-tags}

The code inside [region tags](author-samples.md#region-tags)
must be syntactically correct and complete on its own. Start the region
just before the first import statement and end it after the last line of
the sample function, including imports, client setup, and the function
body.

Go `package` statements are placed outside region tags and are an
exception to this guideline.

***Why?** The code presented in documentation should run correctly when
copied out of the text.*

#### Bounded Wait for Long-running Operations {#bounded-wait}

Always set the wait for a
[long-running operation](author-samples.md#long-running-operations)
to 90 seconds using the client library's timeout mechanism, set at or
immediately before the wait call, even when the operation takes longer than
90 seconds to complete.

* **At the wait call:** Comment that the call polls for up to 90
  seconds (stating the time unit when the syntax does not make it
  explicit).
* **On timeout:** Return without an error and do not cancel the
  operation. Comment that the operation continues running on the
  server, and print the operation name with a note that it can be used
  to check progress later.

See the [Long-running Operations](author-samples.md#long-running-operations)
examples, which use a 90-second bound.

***Why?** Many long-running operations take several minutes or hours,
which is too long for an API Sample or its tests to block. A fixed
90-second wait lets fast operations finish, demonstrates polling, and
keeps every API Sample testable, while the timeout branch shows how to
hand off the operation name when the work outlasts the wait.*

### Recommendations {#recommendations}

#### Classify Errors with Library Support {#classify-errors-with-library-support}

When possible, use a library that helps identify library error types.

```javascript
import {status} from '@grpc/grpc-js';

// ...

catch (err) {
  if (err.code === status.NOT_FOUND) {
    console.error(
      `error: Data policy '${dataPolicyId}' not found in location '${location}' for project '${projectId}'.`,
    );
  }
}
```

***Why?** The library helps developers find the contracts associated with error type classification. Without this it can be challenging to find consistent guidance.*

#### Specific Error Handling by Verb {#specific-error-handling-by-verb}

This recommendation extends [Error Handling](author-samples.md#errors) with a starting place for which errors to handle. For [long-running operations](author-samples.md#long-running-operations), these errors can be returned either when starting the operation or when the operation concludes.

* **`CREATE`** operations should catch `ALREADY_EXISTS` errors.  
* **`DELETE`, `GET`, `UPDATE`** operations should catch `NOT_FOUND` errors for the specified resource.  
* **`LIST`** operations should catch `NOT_FOUND` errors for the specified project or location.
