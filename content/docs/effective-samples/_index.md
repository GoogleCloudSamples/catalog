---
date: '2026-01-29T21:45:54-08:00'
draft: false
title: 'Effective Samples'
description: 'Principles, scope, and review practices of the Samples Style Guide.'
---

This document defines standards for effective code samples. It encourages recommended practices for authoring code intended to teach others. By consistently applying these practices, the collection of samples as a whole becomes more approachable.

All samples must meet these goals:
  * **Teach through code**: Show users how to use the product and explain why we recommend an approach.
  * **Easy to Use**: Products and samples are trustworthy and delightful when they are easy to use. Sample should be runnable with minimal effort.

## Principles

These principles guide decisions in how to design and implement every sample.

  1. **Clarity:** The code's purpose is easy to understand. Clearly explain any required knowledge.
  2. **Community Standard:** The code should be written using idiomatic, community-driven coding practices.
  3. **Consistency:** The code has the same implementation from language-to-language and recurring patterns between use cases.
  4. **Simplicity:** The code accomplishes its goal in the most concise way possible.
  5. **Portability:** The code should be able to run locally or in the Cloud.
  6. **Maintainability:** Write code that is easy to maintain.

{{< callout >}}**Applying the Principles: *Community Standard vs. Consistency:*** Samples should use preferred code patterns that are idiomatic to the particular language (e.g., promises, futures, static API calls from singletons, builders, etc.). This may result in code that is different from language-to-language. Prefer what is idiomatic to the language. Generally, do things “the way the community does it.”
{{< /callout >}}

{{< callout >}}
**Applying the Principles: *Clarity vs. Simplicity:*** Samples should clearly explain what developers need to know to understand the lesson they teach. Once the lesson has been made explicit, the sample should be as simple as possible.
{{< /callout >}}

## Scope

This guide applies to Google Cloud samples. New samples and rewrites follow
the current version of this guide. A rewrite changes what a sample
demonstrates, such as its API operation or use case, or most of its code.

For smaller changes, such as bug fixes, dependency updates, and edits to a
few lines, bring the code you change into line with this guide, unless
that makes the sample inconsistent.

## Using this Guide

* **As a *Sample Designer***, use this guide to shape how you scope samples, and which details to note as required for Samples Authors.
* **As a *Sample Author***, use this guide to shape how you implement code to optimize for teaching effectiveness.
* **As a *Sample Reviewer***, use this guide to cite sample practice expectations for new samples and proposed changes to samples. Reviewers and authors share responsibility for how samples meet these guidelines.
* **As a *Coding Agent***, start at [`llms.txt`](/llms.txt). It indexes the Markdown version of each page and the full guide, including a version for each language.

### How to make review comments

Feedback on samples PRs should reference the style guide wherever possible. Leave a short description of what’s needed to comply with the style guide, and then link to the relevant guideline.

![Screenshot of a GitHub review comment with the text "This should handle case sensitivity issues to be compatible with Windows. See http://googlecloudplatform.github.io/samples-style-guide/#portability."](example-review-comment.png)
