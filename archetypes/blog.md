---
date: '{{ .Date }}'
draft: true
title: '{{ now.Format "January 2006" }} and <Featured Topic>'
description: "Monthly release notes & <one-line summary of the featured sample engineering or integration topic>."
tags:
  - releasenotes
  # Add topic tag(s), e.g., integrationtips, samplestyle, metadata, testing
---

## <Featured Topic: A Sample Engineering Principle or Integration Tip>

<!--
Hook: What recurring problem did sample authors or users run into this month?
Walk through the pattern or trade-off and link to 1-2 real samples added or
updated this month that demonstrate the recommended approach.
-->

## Posts About Samples Engineering

<!-- Optional: Remove this section if there are no external posts or talks this month. -->
* [<Post or Talk Title>](https://example.com) by <Author>

## Release Notes

<!--
Generate metrics for the target month with:
  go -C _agents/skills/catalog-site/scripts run . metrics --month YYYY-MM --format markdown
-->

This month we added **<N> new samples** and updated **<M> existing samples** across **<L> languages** in the Samples Catalog.

### New Sample Coverage

* **<Product / Service Area>** (`<languages>`): <1-sentence summary of the developer scenarios covered, linking to representative samples>.

### Sample Updates & Modernization

* **<Product / Service Area>** (`<languages>`): <1-sentence summary of what was modernized or improved>.
