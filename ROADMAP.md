# Roadmap

Rough order of work. Items near the top block items further down.

## Near term

- Reference purl normaliser. Pin one implementation by name and version, or write a small standalone one, so independent tools produce byte-identical canonical purls. This is the first blocker for interoperability.
- Reference CLI. `vid <purl> <file>` as a standalone binary, and `scrutineer vid` as the same thing inside scrutineer. The Ruby script in `examples/` is a placeholder.
- Test vectors per purl type. At minimum npm, pypi, gem, golang, cargo, maven, github, generic, each exercising that type's case and encoding rules.
- Record format. A minimal JSON shape for what travels alongside a VID: CWE, severity, location, affected version ranges, aliases (other VIDs, CVE, GHSA), status. The OSV schema is the obvious starting point; a VID record is likely an OSV record with the VID as `id` and a `database_specific` block carrying the preimage.
- Stamp VIDs on scrutineer findings and include the VID line in the `disclose` skill's report drafts.

## Later

- Consumer-side matching in scrutineer. Take an application's SBOM, compute gitoids over the installed tree, match against held VIDs, run reachability per match, emit OpenVEX.
- A comparison channel for researchers. Some way for two parties with private VID lists to learn their overlap. The minimal version is an encrypted bulletin board synced by a vouched group; the design is sketched but unbuilt.
- FAQ expansion as human feedback surfaces recurring questions.
- Advisory database adoption. Get OSV, GHSA, and similar to accept VID strings in their `aliases` fields. This is a conversation more than an implementation.
- A GitHub Action that computes gitoids over a repository at a given commit and reports any matches against a supplied VID set.

## Out of scope for now

- Whole-package vulnerabilities with no fix file.
- A path-based companion identifier stable across file edits.
- Any central registry, authority, or required database.
