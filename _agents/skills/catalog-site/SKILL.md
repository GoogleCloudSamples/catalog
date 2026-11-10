---
name: catalog-site
description: >-
  Guides authoring, updating, and validating documentation, blog posts, Hextra
  shortcodes, and multi-language code snippet examples in the Cloud Samples
  Catalog (catalog) Hugo repository. Use this skill whenever working in the
  catalog repository, editing guidelines under content/docs/, updating code
  examples under content/docs/code/, or preparing small-batch PRs.
---

# Cloud Samples Catalog (`catalog`) Repository Skill

Use this skill when making content, code example, shortcode, or configuration changes in the Cloud Samples Catalog (`catalog`) repository.

## 1. Repository Architecture

The site is built with **Hugo** and the **[Hextra](https://imfing.github.io/hextra)** theme (`github.com/imfing/hextra` in [go.mod](../../../go.mod)).

| Path | Purpose |
| :--- | :--- |
| [hugo.yaml](../../../hugo.yaml) | Site configuration, top-level navigation menu (`About`, `Blog`, `Docs`, `Explore`, `Search`, `Source`), and Hextra settings (`page.tabs.sync: true`). |
| [content/_index.md](../../../content/_index.md) | Landing page with Hextra `cards` shortcodes. |
| [content/about.md](../../../content/about.md) | About page for the Cloud Samples Catalog. |
| [content/blog/](../../../content/blog) | Monthly release notes and sample engineering articles (`YYYY-MM-DD.md`). |
| [archetypes/blog.md](../../../archetypes/blog.md) | Base Hugo archetype template for monthly update and release notes blog posts. |
| [content/docs/effective-samples/](../../../content/docs/effective-samples) | The Samples Style Guide (`_index.md`, `author-samples.md`, `api-samples.md`, `metadata.md`). |
| [content/docs/code/](../../../content/docs/code) | Language-specific code examples included by `guideline_examples.html`. |
| [layouts/_shortcodes/guideline_examples.html](../../../layouts/_shortcodes/guideline_examples.html) | Generates synced language tabs for a guideline slug (`{{% guideline_examples "<slug>" %}}`). |
| [layouts/_shortcodes/include_code.html](../../../layouts/_shortcodes/include_code.html) | Reads a file from the repo root and renders it inside a fenced Markdown code block (`{{% include_code "<path>" "<lang>" %}}`). |
| [layouts/_partials/samples/languages.html](../../../layouts/_partials/samples/languages.html) | The one list of guideline example languages (name, fence language, extension, folder), shared by the HTML tabs and the agent/LLM text outputs. |
| [layouts/_partials/llms/](../../../layouts/_partials/llms) | Plain-Markdown rendering for the agent/LLM outputs (`<page>.md`, `llms-full*.txt`, `llms-rules.txt`). See [Agent and LLM outputs](#agent-and-llm-outputs). |

---

## 2. How Multi-Language Guideline Examples Work

When a documentation page calls `{{% guideline_examples "<slug>" %}}`, [guideline_examples.html](../../../layouts/_shortcodes/guideline_examples.html) iterates through the supported languages, defined once in [languages.html](../../../layouts/_partials/samples/languages.html), in this exact order:

1. **Python**: `content/docs/code/python/<slug>.py` (`python`)
2. **Node.js**: `content/docs/code/node/<slug>.js` (`javascript`)
3. **Java**: `content/docs/code/java/<slug>.java` (`java`)
4. **Go**: `content/docs/code/go/<slug>.go` (`go`)
5. **C#**: `content/docs/code/csharp/<slug>.cs` (`csharp`)
6. **Ruby**: `content/docs/code/ruby/<slug>.rb` (`ruby`)
7. **PHP**: `content/docs/code/php/<slug>.php` (`php`)
8. **Terraform**: `content/docs/code/terraform/<slug>.tf` (`terraform`)

### Critical Shortcode Rules
- **Always call with `%` delimiters**: Use `{{% guideline_examples "<slug>" %}}` and `{{% include_code "<path>" "<lang>" %}}` so Hugo processes the returned Markdown through `.Page.RenderString`.
- **Exact lowercase filenames**: `os.FileExists` is case-sensitive on Linux. The filename in `content/docs/code/<folder>/` **must** match `<slug>.<ext>` (or `<slug>.md` for companion notes) in exact lowercase (for example, `imports.cs` and `imports.java`, never `Imports.cs` or `Imports.java`).
- **Companion `<slug>.md` Note Callouts**: Language-specific author guidance that belongs outside the copy-paste code snippet should be placed in an optional `content/docs/code/<folder>/<slug>.md` file. `guideline_examples.html` renders `<slug>.md` inside a Hextra `{{< callout type="important" icon="annotation" >}}` above the code block (or standalone when no `<slug>.<ext>` file exists, such as `terraform/errors.md`). Keep `content/docs/code/_index.md` configured with `build: { list: never, render: never }` and `cascade.build: { list: never, render: never }` so companion `.md` files are never rendered as standalone pages.
- **Handling "Not applicable"**: If a guideline has neither a code snippet (`<slug>.<ext>`) nor a companion note (`<slug>.md`) for a language (for example, `imports` or `clients` in Terraform), **omit both files entirely**. Do not create a `# N/A` stub file; when neither file exists, `guideline_examples.html` automatically renders `{{< callout type="info">}}Not applicable{{< /callout >}}`.
- **Known highlighting issue**: Hugo's highlighter (Chroma) does not color a comment that follows a top-level `}` in Java and Terraform, so a final `[END ...]` line renders as plain tokens. It is cosmetic; leave the code as is. A trailing newline or blank line does not fix it.
- **Unique tab ids**: `guideline_examples.html` calls Hextra's `shortcodes/tabs` partial directly with an id built from the slug and the call's ordinal (for example, `tabs-region-tags-3`). Don't switch it back to rendering a `{{< tabs >}}` string through `.Page.RenderString`: Hextra numbers tab sets by `.Ordinal`, which restarts at 0 in every `RenderString` call, so every example set would share ids and screen readers would announce the wrong panel labels.

### Agent and LLM outputs

Every page's Markdown version (`<page>.md`) and the `llms-full*.txt` files must
teach exactly what the HTML page teaches, as plain Markdown, with each
language's examples only in the files for that language.

When you:

- **Edit content with existing shortcodes**: nothing extra. For hand-written
  tabs, use the exact language names in `languages.html` (`Node.js`, `C#`).
- **Use a Hextra shortcode not yet used in `content/`**: add
  `layouts/_partials/llms/<name>.txt` that renders it as plain Markdown, plus
  one-line `layouts/_shortcodes/<name>.md` and `<name>.txt` wrappers. Without
  them, its HTML leaks into the agent outputs.
- **Edit `layouts/_partials/llms/`**: pipe every printed value through
  `safeHTML`. Hugo parses partials as HTML and escapes them otherwise.
- **Add a language**: update `languages.html` and `report.Languages()`, add
  the `llms<folder>` output format and its `outputs.home` entry in
  `hugo.yaml`, and copy an existing language's three one-line templates
  (`layouts/llms<folder>.txt`,
  `layouts/_shortcodes/guideline_examples.llms<folder>.txt`, and
  `layouts/_shortcodes/tab.llms<folder>.txt`).
- **Add a full-text template**: name it `layouts/llms<name>.txt`, never
  `home.llms<name>.txt`, which would replace `/llms.txt`.

Verify with the default `go -C _agents/skills/catalog-site/scripts run .`,
which includes the `templates`, `languages`, and `shortcodes` checks for these rules.

---

## 3. Style Guide Standards for `content/docs/code/` Examples

Examples in `content/docs/code/<folder>/` are often stripped down to focus on a
single guideline, so don't expand an example (such as adding region tags,
imports, client setup, or error handling) just to satisfy other guidelines. For
the code that *is* included, follow the applicable guidelines below:

1. **Copy-Paste-Callable**: Package snippets inside a well-named function/method with standard parameter documentation. Do not include CLI argument parsers.
2. **Region Tags**:
   - Format: `snake_case` matching `[a-z][a-z0-9_]*[a-z0-9]`, starting with the product prefix.
   - Include `import` / `using` / `require` statements **inside** the `[START ...]` and `[END ...]` tags.
   - Exclude maintenance boilerplate (`package` in Go, `'use strict';` in Node.js) from region tags.
3. **Minimal Arguments**:
   - Limit function arguments to environment-specific values: `project_id`, `location_id`, `resource_id`, and `local_file_path` (in the language's idiomatic casing, e.g., `projectId` or `project_id`).
   - Use `example-` prefixes for placeholder values (e.g., `"example-project-id"`), not `my-`.
4. **Inline Behavioral Parameters**: Hard-code behavioral literals (such as `target_language_code="es"`, `filter="supported_by=RISK_ANALYSIS"`) directly in the request construction rather than passing them as function parameters.
5. **Process the Result (`stdout`)**: Avoid returning values from sample functions. Print a concise summary of the response to `stdout`.
6. **Error Handling & Logging (`stderr`)**:
   - Catch specific Google Cloud / API exceptions rather than silently ignoring errors.
   - Log errors to `stderr` prefixed with `"error:"` and include troubleshooting context (Project, Location, Resource).
7. **No Internal Author Notes**: Never leave `# NOTE FOR SAMPLE AUTHOR` or `TODO` meta-comments inside `content/docs/code/` files—everything in those files is rendered directly to readers.
8. **No License Headers**: Do not add copyright or license headers to `content/docs/code/` files. They are fragments rendered inline in the guide, not published sample source files, so the [License Headers](../../../content/docs/effective-samples/author-samples.md#license-headers) guideline does not apply to them.
9. **Convey the Guideline, Not a Program**: Keep only the code the guideline is about. Drop setup, fake API calls, and boilerplate that don't make the point.
10. **Stay Syntactically Complete**: Trimmed examples must still be valid where the guideline depends on it. For example, Go `region-tags.go` imports `fmt` and uses it, because unused imports don't compile.
11. **Don't Contradict Applicable Guidelines**: Keep examples stripped down to
    the guideline being taught (see item 9), without adding code just to cover
    other guidelines. For the code that is shown, don't violate other guidelines
    that apply to that sample type—and don't apply sample-type-specific rules
    (such as [API Operation Samples](../../../content/docs/effective-samples/api-samples.md))
    to examples for other sample types or general guidelines unless the example
    itself demonstrates that construct (for example, when an example builds an
    API request, never set an optional field to its API default).
12. **Comments for This Guideline Only**: Don't add comments that belong to another guideline (such as client reuse outside `clients.*`), that list alternative values or defaults, or that narrate intent (such as "harmless example value").
13. **Current Client Surfaces**: PHP uses `V2\Client\*` clients with request objects. Go imports `cloud.google.com/go/<api>/apiv<n>/<api>pb`, not `google.golang.org/genproto`.
14. **Consistent Across Languages**: Within one example, use the same values and output labels in every language. The values are illustrative, not required. For example, use `example-project-id` for a project ID, and choose a location the API has capacity in rather than defaulting to `us-central1`.
15. **Format Go with `gofmt`**: For fragments without a `package` line, check with a temporary `package x` header.
16. **Terraform**: Add a `.tf` example when HCL can show the rule (for example, `sensitive = true`); use a companion `.md` for notes.

---

## 4. Branching, Small-Batch PRs & Verification Workflow

1. **Always Work on a Branch**:
   - Create a focused branch off `main` for each small batch of related changes (`feat/<topic>`, `fix/<topic>`, `docs/<topic>`).
   - Sync first: `git pull --rebase origin main` before creating each branch.
   - To update a branch that is under review, merge `main` into it rather than rebasing. PRs are squash-merged, so branch history doesn't matter, and merging avoids a force-push.
   - Never commit directly to `main`.
2. **Validate Before Committing**:
   - Run the Go/Cobra repository validator CLI from the repository root to verify shortcode file mappings, case sensitivity, internal Markdown links/anchors, leaked author notes, and code snippet syntax:
     ```bash
     # Run all domain validators (shortcodes, links, snippets)
     go -C _agents/skills/catalog-site/scripts run .

     # Or target a specific validation domain
     go -C _agents/skills/catalog-site/scripts run . shortcodes
     go -C _agents/skills/catalog-site/scripts run . links
     go -C _agents/skills/catalog-site/scripts run . snippets
     go -C _agents/skills/catalog-site/scripts run . templates
     go -C _agents/skills/catalog-site/scripts run . languages

     # Collect monthly created vs. updated sample metrics across the 8 sample repos
     go -C _agents/skills/catalog-site/scripts run . metrics --month 2026-09 --format markdown
     ```
   - The `templates` check reads the layouts that render the agent and LLM text outputs, without building the site. It fails on a `layouts/home.llms*.txt` layout (it would replace `/llms.txt`), on a printed value in `layouts/_partials/llms/` without `safeHTML`, and on a shortcode used in `content/` without `.md` and `.txt` overrides. The `languages` check fails when `report.Languages()` drifts from `languages.html`, or a language is missing its output format, `outputs.home` entry, or templates.
   - If `hugo` is available in the environment, run `hugo --buildDrafts` to verify the full site build.
   - Preview with `hugo server --buildDrafts --disableFastRender`. Edits to `content/docs/code/` files don't re-render the guide page, because the shortcode reads them directly; restart the server or touch the guide page to see them.
3. **Commit & PR Conventions**:
   - Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`) with a clear one-line summary and a detailed description of the problem and solution.
   - Push the branch to `origin` and open a focused pull request for review.

---

## 5. Authoring Guidelines

When writing or revising a guideline in `content/docs/effective-samples/`:

1. **Rule First**: Open with the rule as one imperative sentence. Put definitions, exceptions, and lists after it.
2. **Checkable Rules**: Replace soft limits ("completely", "as closely as possible", "where possible") with a test a reviewer can apply, such as "a value that works unchanged in any developer's environment".
3. **Don't Restate Other Guidelines**: Link to another guideline only when it gives a deeper, more specific treatment of the same aspect of a sample (for example, the Assert step links to Process the result). Don't link to related guidelines just to note that they exist.
4. **Keep General Statements General**: The guideline's intro should hold for any sample. Put specifics, such as API requests, in the steps or details that follow.
5. **Describe the Sample, Not the Reader**: Write "to show how the sample gets from setup to result", not "so readers can see...".
6. **One Set of Examples**: Don't list examples in the prose and again in a bulleted list.
7. **Neutral Placeholders in Prose**: Illustrate with placeholders such as `request.config.options.mode`. A real API in prose that doesn't match the code examples confuses readers.
8. **Rules Defined Here Aren't External Patterns**: Don't present a structure the guide defines, such as Arrange, Act, Assert, as an outside pattern; the heading names it and the guideline defines it.
9. **Consolidate Scattered Exceptions**: When exceptions about one topic pile up across unrelated guidelines, move them into a dedicated guideline (for example, Protect Sensitive Data).
10. **Guide-Wide Over Type-Specific**: If a rule applies to all samples, state it once in `author-samples.md` and remove type-specific copies from `api-samples.md`.
11. **Why? States the Benefit**: Explain the developer benefit, which can be more than one (such as learning and customization). Don't restate the rule.
12. **Cross-Page Links**: Link to another page by its source file, relative to the current file, for example `author-samples.md#minimal-arguments` or `../../about.md`. Hextra resolves these to the page's URL, and they also work when browsing the repository. Link to generated site files, such as `/llms.txt`, with root-relative paths.
