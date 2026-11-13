---
name: catalog-site
description: >-
  Guides authoring, updating, and validating documentation, blog posts, Hextra
  shortcodes, and multi-language code snippet examples in the Cloud Samples
  Catalog (catalog) Hugo repository. Use this skill whenever working in the
  catalog repository, editing guidelines under content/docs/, writing posts
  under content/blog/, updating code examples under content/docs/code/, or
  preparing small-batch PRs.
---

# Cloud Samples Catalog (`catalog`) Repository Skill

## Repository Architecture

The site is built with **Hugo** and the **[Hextra](https://imfing.github.io/hextra)** theme (`github.com/imfing/hextra` in [go.mod](../../../go.mod)).

| Path | Purpose |
| :--- | :--- |
| [hugo.yaml](../../../hugo.yaml) | Site configuration, top-level navigation menu (`About`, `Blog`, `Docs`, `Explore`, `Search`, `Source`), and Hextra settings (`page.tabs.sync: true`). |
| [content/_index.md](../../../content/_index.md) | Landing page with Hextra `cards` shortcodes. |
| [content/about.md](../../../content/about.md) | About page for the Cloud Samples Catalog. |
| [content/blog/](../../../content/blog) | Sample engineering articles and monthly release notes (`YYYY-MM-DD.md`). |
| [archetypes/blog.md](../../../archetypes/blog.md) | Base Hugo archetype template for monthly update and release notes blog posts. |
| [content/docs/effective-samples/](../../../content/docs/effective-samples) | The Samples Style Guide (`_index.md`, `author-samples.md`, `api-samples.md`, `metadata.md`). |
| [content/docs/code/](../../../content/docs/code) | Language-specific code examples included by `guideline_examples.html`. |
| [layouts/_shortcodes/guideline_examples.html](../../../layouts/_shortcodes/guideline_examples.html) | Generates synced language tabs for a guideline slug (`{{% guideline_examples "<slug>" %}}`). |
| [layouts/_shortcodes/include_code.html](../../../layouts/_shortcodes/include_code.html) | Reads a file from the repo root and renders it inside a fenced Markdown code block (`{{% include_code "<path>" "<lang>" %}}`). |
| [layouts/_partials/samples/languages.html](../../../layouts/_partials/samples/languages.html) | The one list of guideline example languages (name, fence language, extension, folder), shared by the HTML tabs and the agent/LLM text outputs. |
| [layouts/_partials/llms/](../../../layouts/_partials/llms) | Plain-Markdown rendering for the agent/LLM outputs (`<page>.md`, `llms-full*.txt`, `llms-rules.txt`). See [Agent and LLM outputs](#agent-and-llm-outputs). |

---

## Writing Guideline Text

### Creating a New Guideline

Work through these steps in order:

1. **Confirm it's new.** Search [content/docs/effective-samples/](../../../content/docs/effective-samples) (`author-samples.md`, `api-samples.md`, `metadata.md`, and `_index.md`) for the topic. If an existing guideline covers part of it, extend that guideline instead.
2. **Choose the page.** Use `author-samples.md` when the rule applies across sample types, `api-samples.md` only when it applies exclusively to API operation samples (see **Guide-Wide Over Type-Specific** under Conventions), `metadata.md` for use-case metadata and inventory governance, and `_index.md` for guide-wide principles, scope, or review practices.
3. **Choose the section.** In `author-samples.md`, use the section that names the concern: Embeddable, Structure, Code, Respect, or Maintainable & Testable. In `api-samples.md`, put requirements under Guidelines and optional practices under Recommendations.
4. **Name it.** Write a Title Case heading that names the practice, with an explicit kebab-case anchor, such as `### Protect Sensitive Data {#sensitive-data}`. Reviewers link to the anchor from PR comments, so don't change it after publishing. Use the anchor as the example slug.
5. **Draft from the template** below.
6. **Add examples when code shows the rule better than prose.** Follow [Writing Guideline Code Examples](#writing-guideline-code-examples). Skip examples for rules about wording or process, such as Respectful Words.
7. **Verify API claims in library source.** Before a rule, note, or example states how a client library behaves (such as whether a client has a close or dispose method), confirm it in the library's source code. Documentation summaries and memory aren't enough.
8. **Update the page's `description`** front matter if it lists the page's topics.
9. **Validate and preview** as described in [Branching, Small-Batch PRs & Verification Workflow](#branching-small-batch-prs-verification-workflow).

#### Template

```markdown
### Guideline Name {#slug}

**Applies to:** {{< icon "scissors" >}}Code Snippets

One imperative sentence stating the rule.

* **Case or sample type:** what to do.
* **Another case:** what to do.

One paragraph for exceptions.

{{% guideline_examples "slug" %}}

{{< callout type="info" >}}
**Recommendation for <subset of samples>**

Optional advice for that subset.
{{< /callout >}}

***Why?** The developer benefit.*
```

Include only the parts the guideline needs, in this order. Add the **Applies to** line only when the guideline is narrower than its section's default (in `author-samples.md`, `Structure` defaults to `Code Snippets`, while `Embeddable`, `Code`, `Respect`, and `Maintainable & Testable` default to all samples). End with the Why? and put nothing after it.

### Conventions

When writing or revising a guideline in `content/docs/effective-samples/`:

1. **Rule First**: Open with the rule as one imperative sentence. Put definitions, exceptions, and lists after it.
2. **Checkable Rules**: Replace soft limits ("completely", "as closely as possible", "where possible") with a test a reviewer can apply, such as "a value that works unchanged in any developer's environment".
3. **Don't Restate Other Guidelines**: Link to another guideline only when it gives a deeper, more specific treatment of the same aspect of a sample (for example, the Assert step links to Process the result). Don't link to related guidelines just to note that they exist.
4. **Keep General Statements General**: The guideline's intro should hold for any sample. Put specifics, such as API requests, in the steps or details that follow.
5. **Describe the Sample, Not the Reader**: Write "to show how the sample gets from setup to result", not "so readers can see...".
6. **One List of Examples**: Don't list examples in the prose and again in a bulleted list.
7. **Neutral Placeholders in Prose**: Illustrate with placeholders such as `request.config.options.mode`. A real API in prose that doesn't match the code examples confuses readers.
8. **Rules Defined Here Aren't External Patterns**: Don't present a structure the guide defines, such as Arrange, Act, Assert, as an outside pattern; the heading names it and the guideline defines it.
9. **Consolidate Scattered Exceptions**: When exceptions about one topic pile up across unrelated guidelines, move them into a dedicated guideline (for example, Protect Sensitive Data).
10. **Guide-Wide Over Type-Specific**: If a rule applies to all samples, state it once in `author-samples.md` and remove type-specific copies from `api-samples.md`.
11. **State Contradictory Language Idioms in Prose**: When a language's idiom departs from the headline rule (such as Go writing to `io.Writer` or returning wrapped errors rather than printing to `stdout`/`stderr`, or Terraform declaring `output` blocks), state that carve-out briefly in the guideline prose as well as in `<slug>.md`. `llms-rules.txt` omits both code blocks and companion `<slug>.md` notes, so rules-only readers still need the exception.
12. **Why? States the Benefit**: Explain the developer benefit, which can be more than one (such as learning and customization). Don't restate the rule.
13. **Cross-Page Links**: Link to another page by its source file, relative to the current file, for example `author-samples.md#minimal-arguments` or `../../about.md`. Hextra resolves these to the page's URL, and they also work when browsing the repository. Link to generated site files, such as `/llms.txt`, with root-relative paths.
14. **No Hard Wrapping**: Write each paragraph and bullet on one line, even when nearby text is wrapped.

---

## Writing Guideline Code Examples

Each example in `content/docs/code/<folder>/` teaches one guideline. Keep only the code that guideline is about: drop setup, fake API calls, and boilerplate that don't make the point, and don't add code (such as region tags, imports, client setup, or error handling) just to satisfy other guidelines. For the code that is shown, follow these rules:

1. **Sample Functions**: When the guideline is about function-level structure (such as arguments, client lifecycle, or printing results), wrap the example in a well-named sample function with standard parameter documentation. Otherwise, a fragment is fine. Don't include CLI argument parsers.
2. **Follow the Guide, Except for Focus**: Code shown in an example follows the guide's rules for its sample type, except where departing from a rule makes the example clearer and more focused on the guideline it illustrates (for example, skipping client cleanup in an example about building a request). Apply rules for one sample type, such as [API Operation Samples](../../../content/docs/effective-samples/api-samples.md), only to examples of that type, unless the example itself demonstrates that construct (for example, when an example builds an API request, never set an optional field to its API default).
3. **No Internal Author Notes**: Never leave `# NOTE FOR SAMPLE AUTHOR` or `TODO` meta-comments inside `content/docs/code/` files—everything in those files is rendered directly to readers.
4. **No License Headers**: Do not add copyright or license headers to `content/docs/code/` files. They are fragments rendered inline in the guide, not published sample source files, so the [License Headers](../../../content/docs/effective-samples/author-samples.md#license-headers) guideline does not apply to them.
5. **Stay Syntactically Complete**: Trimmed examples must still be valid where the guideline depends on it. For example, Go `region-tags.go` imports `fmt` and uses it, because unused imports don't compile.
6. **Comments for This Guideline Only**: Don't add comments that belong to another guideline (such as client reuse outside `clients.*`), that list alternative values or defaults, or that narrate intent (such as "harmless example value"). Put explanations of the example itself in the companion `<slug>.md` note, not in code comments, because agents copy comments from examples into the code they write.
7. **Current Client Surfaces**: PHP uses `V2\Client\*` clients with request objects. Go imports `cloud.google.com/go/<api>/apiv<n>/<api>pb`, not `google.golang.org/genproto`.
8. **Consistent Across Languages**: Within one example, use the same values and output labels in every language. The values are illustrative, not required. Use `example-` prefixes for placeholder values (such as `example-project-id`), not `my-`, and choose a location the API has capacity in rather than defaulting to `us-central1`.
9. **Format Go with `gofmt`**: For fragments without a `package` line, check with a temporary `package x` header.
10. **Terraform**: Add a `.tf` example when HCL can show the rule (for example, `sensitive = true`); use a companion `.md` for notes.

---

## How Guideline Code Examples Render

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
- **Companion `<slug>.md` Note Callouts**: `guideline_examples.html` renders an optional `content/docs/code/<folder>/<slug>.md` file inside a Hextra `{{< callout type="important" icon="annotation" >}}` above the code block (or standalone when no `<slug>.<ext>` file exists, such as `terraform/errors.md`). Keep `content/docs/code/_index.md` configured with `build: { list: never, render: never }` and `cascade.build: { list: never, render: never }` so companion `.md` files are never rendered as standalone pages.
- **Handling "Not applicable"**: If a guideline has neither a code snippet (`<slug>.<ext>`) nor a companion note (`<slug>.md`) for a language (for example, `imports` or `clients` in Terraform), **omit both files entirely**. Do not create a `# N/A` stub file; when neither file exists, `guideline_examples.html` automatically renders `{{< callout type="info">}}Not applicable{{< /callout >}}`.
- **Known highlighting issue**: Hugo's highlighter (Chroma) does not color a comment that follows a top-level `}` in Java and Terraform, so a final `[END ...]` line renders as plain tokens. It is cosmetic; leave the code as is. A trailing newline or blank line does not fix it.
- **Unique tab ids**: `guideline_examples.html` calls Hextra's `shortcodes/tabs` partial directly with an id built from the slug and the call's ordinal (for example, `tabs-region-tags-3`). Don't switch it back to rendering a `{{< tabs >}}` string through `.Page.RenderString`: Hextra numbers tab sets by `.Ordinal`, which restarts at 0 in every `RenderString` call, so every example set would share ids and screen readers would announce the wrong panel labels.

### Agent and LLM outputs

Every page's Markdown version (`<page>.md`) and the `llms-full*.txt` files must teach exactly what the HTML page teaches, as plain Markdown, with each language's examples only in the files for that language.

When you:

- **Edit content with existing shortcodes**: nothing extra. For hand-written tabs, use the exact language names in `languages.html` (`Node.js`, `C#`).
- **Use a Hextra shortcode not yet used in `content/`**: add `layouts/_partials/llms/<name>.txt` that renders it as plain Markdown, plus one-line `layouts/_shortcodes/<name>.md` and `<name>.txt` wrappers. Without them, its HTML leaks into the agent outputs.
- **Edit `layouts/_partials/llms/`**: pipe every printed value through `safeHTML`. Hugo parses partials as HTML and escapes them otherwise.
- **Add a language**: update [languages.html](../../../layouts/_partials/samples/languages.html) and `report.Languages()` in [report.go](scripts/internal/report/report.go), add the `llms<folder>` output format, `outputs.home` entry, and `params.page.contextMenu.links` entry in [hugo.yaml](../../../hugo.yaml), copy an existing language's three one-line templates (`layouts/llms<folder>.txt`, `layouts/_shortcodes/guideline_examples.llms<folder>.txt`, and `layouts/_shortcodes/tab.llms<folder>.txt`), create `content/docs/code/<folder>/` examples as applicable, and update hand-written `{{< tabs >}}` sets and `Code Style` in [author-samples.md](../../../content/docs/effective-samples/author-samples.md).
- **Add a full-text template**: name it `layouts/llms<name>.txt`, never `home.llms<name>.txt`, which would replace `/llms.txt`.

The `templates`, `languages`, and `shortcodes` validators enforce the layout and registration rules above.

---

## Writing Blog Posts

Posts in [content/blog/](../../../content/blog) (`YYYY-MM-DD.md`) cover stories, sample engineering patterns, and monthly catalog dispatches:

- **Front matter & publishing**: Include `date`, `title`, `description`, and topic `tags` (such as `samplestyle`, `integrationtips`, `metadata`, `testing`, or `releasenotes`). Keep `draft: true` while drafting; set `draft: false` only when ready to publish on the production site.
- **Links & inline code**: Link to guide pages via relative `.md` source paths (such as `../docs/effective-samples/author-samples.md#minimal-arguments`), and ensure inline sample code follows the Effective Samples Guide.
- **Standalone topical posts**: Articles unpacking a design principle, style change, or integration pattern use free-form headings and real sample links without release rollups or scorecards.
- **Monthly catalog release posts**: One recurring post type, scaffolded from [archetypes/blog.md](../../../archetypes/blog.md) and tagged `releasenotes` plus topic tag(s):
  1. Collect raw monthly activity across the sample repositories (authenticate via `GITHUB_TOKEN` or `gh auth` to avoid API rate limits):
     ```bash
     go -C _agents/skills/catalog-site/scripts run . metrics --month YYYY-MM --format markdown
     ```
  2. Open the post with [release_metrics.html](../../../layouts/_shortcodes/release_metrics.html) (`{{< release_metrics month="..." created="..." updated="..." languages="..." lang_list="..." highlights="..." >}}`).
  3. Lead with a technical spotlight on a notable cross-catalog theme, then curate `New Sample Coverage` and `Sample Updates & Modernization` into grouped product narratives with representative sample links rather than pasting raw commit subjects verbatim.

---

## Branching, Small-Batch PRs & Verification Workflow

1. **Always Work on a Branch**:
   - Create a focused branch off `main` for each small batch of related changes (`feat/<topic>`, `fix/<topic>`, `docs/<topic>`).
   - Sync first: `git pull --rebase origin main` before creating each branch.
   - To update a branch that is under review, merge `main` into it rather than rebasing. PRs are squash-merged, so branch history doesn't matter, and merging avoids a force-push.
   - Never commit directly to `main`.
2. **Validate Before Committing**:
   - CI deploys Hugo on `main` without running the Go validator, so always run the repository checks locally before committing:
     ```bash
     # Run all 5 domain validators (shortcodes, links, snippets, templates, languages)
     go -C _agents/skills/catalog-site/scripts run .

     # Or target a single validation domain
     go -C _agents/skills/catalog-site/scripts run . [shortcodes|links|snippets|templates|languages]

     # Run unit tests whenever modifying _agents/skills/catalog-site/scripts/
     go -C _agents/skills/catalog-site/scripts test ./...
     ```
   - **Validator coverage bounds**: `shortcodes`, `links`, `templates`, and `languages` verify shortcode file mappings/case, relative `.md` links and `{#anchor}` targets, `llms` template rules, and language registration. `snippets` checks `"NOTE FOR/TO SAMPLE AUTHOR"` comments, region tag format/matching, single-line double quotes in C#/Java/Go/Terraform, and full syntax **only for Ruby** (`ruby -c`)—verify `TODO` absence, formatting (`gofmt`), and syntax for other languages directly.
   - If `hugo` is available, run `hugo --buildDrafts` to check drafts and `hugo` (without `--buildDrafts`) to confirm production build visibility (`draft: false` on pages meant to publish).
   - Preview with `hugo server --buildDrafts --disableFastRender`. Edits to `content/docs/code/` files don't re-render the guide page, because the shortcode reads them directly; restart the server or touch the guide page to see them.
3. **Commit & PR Conventions**:
   - Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`) with a clear one-line summary and a detailed description of the problem and solution.
   - Push the branch to `origin` and open a focused pull request for review.
