---
date: '2026-01-29T22:09:43-08:00'
draft: false
title: 'Metadata'
description: 'Planned rules for sample metadata used in governance, inventory, and testing.'
weight: 4
---

Sample metadata is structured, machine-readable information co-located with sample code in version control that defines the **intent, scope, and implementations of a sample use case**.

Rather than treating individual code files or [region tags](author-samples.md#region-tags) as disconnected documentation fragments, metadata establishes each sample use case as a governed technical asset that facilitates sample governance, inventory management, discovery, and automated validation.

## Why Every Use Case Needs Metadata {#why-metadata}

Without co-located metadata, sample code and the descriptive context surrounding it (titles, descriptions, product mappings, and supported languages) live in separate systems. Requiring metadata for every sample use case solves four core challenges:

* **Single Source of Truth (Preventing Data Drift):** Keeping titles, descriptions, and narratives alongside the code in GitHub ensures that documentation and the Samples Browser stay synchronized with the actual implementation as code evolves.
* **Cross-Language Consistency:** A single use-case definition groups all language implementations (such as Python, Node.js, Java, Go, C#, Ruby, PHP, and Terraform) of the same logical task. This makes it possible to verify that every language teaches the same lesson and implements the expected region tags.
* **Discoverability & Coverage Analysis:** Metadata connects samples to the Google Cloud products, services, and API operations they demonstrate. Developers, technical writers, and product teams can discover samples by use case or API operation—even for multi-file patterns that do not map 1:1 to a single region tag—and identify coverage gaps across languages.
* **Preserving Intent for Maintainers & AI:** Capturing the *Developer Narrative* and key lessons records *why* the sample exists and *what* it is meant to teach. Maintainers can safely update libraries and idioms without losing the original teaching objective, and AI systems gain verified intent paired with working code.

## What Metadata Captures {#what-metadata-captures}

A sample metadata definition describes a logical use case and links it to the concrete code implementations that fulfill it:

* **Identity & Summary:** A unique identifier, human-readable title, and concise description used across the Samples Browser, search, and catalog tooling.
* **Objective & Developer Narrative:** The developer problem being solved, the teaching goal written for maintainers, the target Google Cloud products, and the core lessons demonstrated.
* **Use-Case & API Context:** For [API Operation Samples](api-samples.md), the service and method being demonstrated (for example, `secretmanager.googleapis.com` and `GetSecret`), enabling coverage tracking across the API surface.
* **Region Tags:** The canonical list of documentation [region tags](author-samples.md#region-tags) associated with the use case, linking embedded documentation snippets back to the sample definition.
* **Language Implementations:** The set of language and runtime implementations that fulfill the use case in the repository.

## Guidelines {#guidelines}

### One Metadata Definition per Use Case {#one-per-use-case}

Define a single metadata record for each logical sample use case, and list all language or runtime implementations of that use case within it. Do not create separate, disconnected metadata definitions for each programming language.

***Why?** Grouping implementations under one use case guarantees a consistent developer experience across languages and enables automated checks to verify that all expected languages implement the same scenario.*

### Co-locate Metadata with Code {#co-locate-metadata}

Store metadata files in version control alongside the sample code they describe (for example, in a `.samples/` directory within the service folder), and update metadata in the same pull request whenever adding a new language implementation, changing region tags, or updating a sample's scope.

***Why?** Co-locating metadata with code eliminates manual synchronization steps and ensures review and validation happen as part of the standard code review workflow.*

### Capture the Developer Narrative {#developer-narrative}

Include a clear **Developer Narrative** in the sample's objective that explains the user journey and why a developer would perform the operation, rather than merely restating the API method or function name.

***Why?** The title and description introduce the sample to readers, while the narrative guides future maintainers when refactoring code, upgrading client libraries, or reviewing whether a change preserves the lesson.*

### Keep Code and Metadata Aligned {#alignment}

Every implementation and region tag declared in metadata must correspond to actual sample code in the repository, and every sample region tag in code must be accounted for in metadata:

* **No Missing Implementations:** If metadata declares an implementation for a language, a corresponding sample file in that language containing the expected [region tag](author-samples.md#region-tags) must exist.
* **No Orphaned Snippets:** Every sample `[START ...]` region tag in code must be claimed by a metadata definition so no sample goes untracked.

***Why?** Automated repository validators rely on this bidirectional mapping to catch missing language coverage, mismatched region tags, and untracked snippets before changes are merged.*

### Preserve Human Refinements {#preserve-refinements}

While initial metadata files may be bootstrapped from API definitions or code generation tooling, human refinements to titles, descriptions, and developer narratives are authoritative. Automated updates to implementation lists or technical fields must preserve human-authored prose.

***Why?** API schemas can describe method signatures, but human authors provide the real-world context and problem framing that make a sample effective for learning.*
