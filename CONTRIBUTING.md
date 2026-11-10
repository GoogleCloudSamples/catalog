# How to Contribute

We would love to accept your patches and contributions to this project.

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a
[Contributor License Agreement](https://cla.developers.google.com/about) (CLA).
You (or your employer) retain the copyright to your contribution; this simply
gives us permission to use and redistribute your contributions as part of the
project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our Community Guidelines

This project follows [Google's Open Source Community
Guidelines](https://opensource.google/conduct/).

## Contribution process

### Code Reviews

All submissions, including submissions by project members, require review. We
use [GitHub pull requests](https://docs.github.com/articles/about-pull-requests)
for this purpose.

### Before you open a pull request

Guide pages are in `content/docs/effective-samples/`, and code examples are in
`content/docs/code/<language>/`.

Preview your change locally and run the repository checks:

```shell
# Preview the site, including draft pages, at
# http://localhost:1313/catalog/
hugo server --buildDrafts

# Validate shortcodes, links, code snippets, and agent output templates
go -C _agents/skills/catalog-site/scripts run .
```

The site requires [Hugo Extended](https://gohugo.io/installation/) 0.146 or
later, [Go](https://go.dev/doc/install), and Git. The
[Hextra](https://imfing.github.io/hextra) theme is a Hugo module, so Hugo
downloads it on first run.

`--buildDrafts` includes pages marked `draft: true` in your local preview. The
production site doesn't build drafts, so set `draft: false` in the front matter
of any page that should be published.
