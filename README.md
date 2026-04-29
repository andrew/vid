# VID

A way to give a vulnerability an identifier the moment you find it, without asking anyone, such that anyone else who finds the same vulnerability gets the same identifier.

The target is open source libraries: code published to a package registry and pulled in as a dependency by other projects. The construction works for anything with a package URL and a file, but the use cases below assume a library with many downstream users rather than an application that is deployed and run.

See [SPEC.md](SPEC.md) for the construction. This document is the why.

## The problem

Finding vulnerabilities in open source used to be slow, skilled work. A researcher might spend a week on one library and come away with one good bug. The systems built around that pace (report it privately to the maintainer, wait for a fix, request a CVE number, publish an advisory) assumed findings were rare and each one got individual attention.

That assumption no longer holds. Automated tools, particularly ones built on large language models, can now audit a codebase in minutes and surface real weaknesses at a rate no human reviewer could match. Anyone running such a tool against open source for a few days will accumulate more findings than they could disclose in a year, and popular libraries are being scanned by many such people at once.

Several things break at that rate.

Maintainers get the same report many times. The tools converge: point five of them at the same library and they find largely the same bugs. Five researchers each send a polite, well-formatted report to a volunteer maintainer who now has five email threads about one problem. After a few rounds of this the maintainer stops reading security reports. Maintainers of widely-used projects are already [writing about the influx](https://daniel.haxx.se/blog/2026/04/22/high-quality-chaos/).

Reporting becomes the bottleneck. Writing a clear report, finding the right contact, and following up takes far longer than finding the bug did, so a backlog accumulates faster than any individual can responsibly work through it. Most of those findings never reach the maintainer or the users.

Users of the library are blind until publication. A company that depends on the library will eventually see a CVE in their scanner, months after the bug was first found, and only then start working out whether the vulnerable code is actually reachable from their application. Usually it is not, but proving that under time pressure is its own cost.

The CVE system itself is strained. Identifier assignment goes through a small number of authorities, enrichment of published records is backlogged, and the funding behind the central database has been publicly unstable. Building more process on top of it is not appealing.

## What a VID is

A VID is computed, not assigned. You take the package name (its purl, a standard format like `pkg:npm/lodash`) and a hash of the exact content of the file where the bug lives (the sink file, in the spec's terms), combine them in a fixed way, and hash the result. Out comes a short string like `VID-1-4fta-nqe5-ppht-2m6r-2tpk-o57c`.

Because the inputs describe where the vulnerability is, anyone looking at the same bug computes the same string regardless of who they are or when they look. There is nobody to ask and nothing to wait for.

The file hash is the same construction git uses internally and that the [OmniBOR](https://omnibor.io) project standardised for identifying software artefacts without an authority. VID adds the package name on top, because the same vulnerable file copied into fifty packages means fifty separate conversations with fifty maintainers. Related schemes worth knowing about: [SWHID](https://www.swhid.org) identifies source artefacts the same way for archival; SARIF fingerprints and similar per-tool hashes deduplicate findings within a single scanner's output but are not designed to match across tools or people.

The weakness class (CWE), severity, the line number, the data flow, and everything else a report contains travel alongside the VID as metadata. They are not part of the identifier, because two researchers will agree on which file the bug is in far more reliably than they will agree on which CWE best describes it, and the identifier is only useful if independent finders converge on it. A report header might look like:

    VID:       VID-1-4fta-nqe5-ppht-2m6r-2tpk-o57c
    Package:   pkg:npm/example
    File:      lib/query.js
    CWE:       CWE-89
    Severity:  high
    Location:  lib/query.js:42

    User-supplied `req.params.id` reaches `db.query()` without parameterisation...

The first line is computed; everything below it is the researcher's analysis.

The file hash works the same on a git checkout, a release tarball, or an installed package, because it depends only on bytes. For packages that publish their source unchanged (most of Go, Python, Ruby, Rust, PHP) those three are the same bytes and everyone converges. For packages that ship built output, the repository file and the published file differ and so do their VIDs, meaning the same bug can legitimately carry one VID from a source scan and another from an installed-tree scan; the published archive is the reference point for consumer-side matching, and mapping between the two is a tooling concern. Choosing which file represents the vulnerability is the one judgment call left in an otherwise mechanical scheme; the spec gives a rule and examples, and accepts that occasional disagreement produces two VIDs for one issue, linked after the fact. In practice the tools surfacing findings at this volume are also the tools computing the VID, and automated scanners that converge on the same vulnerability tend to converge on the same sink file.

## What it lets you do

A maintainer receiving five reports that each carry a VID can see at a glance that four are duplicates. That alone needs no infrastructure beyond a convention that reports include the identifier. The larger reduction comes when researchers compare VID lists with each other before reporting and send one email between them, which needs a comparison channel this document does not specify.

A researcher can publish a VID the day they find something, in a commit message or a signed log entry, as a timestamped record that they found it, without revealing what it is. When the advisory is eventually published the inputs are revealed and the early timestamp is provable.

Two researchers who each have a private list of findings can compare VID lists and learn where they overlap without showing each other the underlying details. The overlap tells them where to combine effort; the remainder is what is uniquely theirs.

A user of the library who has access to a set of VIDs, from a researcher they trust or from a future shared feed, can compute file hashes over their installed dependency tree and check for matches. A match tells them which file in which package to look at, which is enough to check whether their own application ever calls into that file with untrusted input. Usually it does not, and they can record that conclusion against the VID now. When a CVE is later assigned and their scanner lights up, the analysis is already done. VID lets consumers invert the disclosure timeline: impact analysis happens before the CVE exists.

The purl decides what a VID is scoped to. A registry purl like `pkg:npm/lodash` names the bug in that one published package. A repository purl like `pkg:github/lodash/lodash` names it in the source, which covers any fork or downstream repackaging (nix, homebrew, conda) that ships the file unchanged, regardless of what registry name it appears under. The same finding can carry both, linked as aliases: registry-scoped for the conversation with that package's maintainer, repo-scoped for tracking the vulnerable file across republishings that CVE-based scanning largely misses today.

Pinning to bytes also catches regressions. If a later release of the package happens to ship the same vulnerable file again after a fix, the old VID matches it where a version-range advisory would not.

After publication a VID is one more alias alongside the CVE and any vendor advisory IDs, and tools that track aliases between identifier schemes could treat it the same way once they recognise the prefix.

Several of these uses get more valuable as more parties exchange VIDs, but none require it. A single researcher with nobody to compare against still gets a stable name for their own records, a timestamp they can prove later, and a dedup key in the maintainer's inbox.

## What it is not

It is not a replacement for CVE. VID is a pre-CVE coordination layer, not a competing namespace. A VID is meant to exist from the moment of finding and then alias to the CVE once one is assigned; the CVE remains the public name everyone cites.

It is not secret. A published VID reveals nothing on its own, but the input space is small enough that someone who suspects which package is involved can hash the candidate files and check. That is still enough for timestamping, which only needs to bind a researcher to a finding.

A VID names one file's worth of bytes in one package, so the same logical vulnerability will accumulate several VIDs over its life: one per content change to the affected file, and one per place those bytes are published, whether vendored into another package, forked and republished under a new name, or repackaged by a downstream distributor such as nix, homebrew, or conda. That spread is intentional, since each of those has its own maintainer or packager who needs their own conversation. Linking related VIDs as the same underlying issue is the job of an `aliases` list in whatever record is kept against each VID, the same way advisory databases already link CVE, GHSA, and vendor IDs today. The identifier gives precision; equivalence is layered on top.

Nothing here is a disclosure policy. When to tell the maintainer, when to publish, and who is allowed to know what remain human decisions. The identifier gives every party a stable name to use while making them.

There is no required database. The spec defines how to compute the string; what anyone stores against it, and who they share that with, is up to them. Existing advisory databases are free to carry VIDs as aliases, and the scheme is more useful if they do, but nothing depends on it.

It does not filter signal from noise. A low-priority finding has a VID as readily as a critical one. Deduplication reduces the count of redundant reports without reducing the count of unimportant ones.

It does not make findings true. A VID can be computed for a false positive as easily as for a real bug. The identifier asserts that someone is making a claim about this file in this package; whether they are right is established the same way it always was.

## Status

Draft. The construction in [SPEC.md](SPEC.md) is settling but not stable. The first thing a reference implementation has to pin is purl normalisation, since a one-byte difference in the canonical package string produces a different identifier and existing purl libraries do not all agree. After that: a CLI, more test vectors covering each ecosystem's case rules, and a worked treatment of packages whose published files differ from their repository source.
