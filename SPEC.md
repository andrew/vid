# VID: content-addressed vulnerability identifiers

Version 1, draft, April 2026.

A VID identifies a vulnerability by hashing where it lives. There is no issuing authority. Two people who find the same vulnerability compute the same VID without talking to each other.

## Inputs

### purl

The [package URL](https://github.com/package-url/purl-spec) of the affected package: a standard string naming a package in a registry, such as `pkg:npm/lodash` or `pkg:gem/rails`. For code not published to a registry, use a `pkg:github/owner/repo` or `pkg:generic/name` purl. A repository purl may also be used for published code when the intent is to scope the finding to the source so that forks and repackagings of the same file are covered; one finding may carry both a registry-scoped and a repository-scoped VID.

The purl MUST be normalised as follows before hashing:

- no `@version` component
- no `?qualifiers` component
- no `#subpath` component
- type and namespace/name case-folded per the type-specific rules in the purl spec (for example, npm names are lowercased; Go module paths are not)
- percent-encoding normalised per the purl spec

These restate the canonical form defined by the purl spec and its per-type rules. A reference normaliser will be named in a later draft because existing purl libraries do not all produce byte-identical output for the same input.

### gitoid

The SHA-256 gitoid of the file in which the vulnerability lives, expressed as 64 lowercase hex characters. The SHA-256 gitoid of a file is

    sha256( "blob " <bytelen> 0x00 <bytes> )

where `<bytelen>` is the file's length in bytes as ASCII decimal and `<bytes>` is the file's content. This is the construction used by git's SHA-256 object format and by [OmniBOR](https://omnibor.io). It can be computed from the file alone; neither git nor a checkout is required.

For a registry purl (`pkg:npm`, `pkg:pypi`, `pkg:gem`, ...) the reference bytes are the file as it appears in the published package archive. The source repository at the corresponding tag is equivalent whenever the package publishes its source unchanged, which is common in Go, Python, Ruby, Rust, and PHP and uncommon in JavaScript, where a build step often sits between the repository and the published tarball. Where repository and archive differ, VIDs computed from each will differ, and tooling that needs consumer-side matching should hash the archive.

For a repository purl (`pkg:github`, `pkg:generic`) the reference bytes are the file as it appears in the checkout being analysed.

### Choosing the file

Use the file containing the call to the dangerous operation. If the vulnerability is a missing check rather than a dangerous call, use the file where the check would be added.

    SQL injection         file containing the query construction or execution call
    command injection     file containing the exec/spawn/system call
    XSS                   file containing the unescaped write to the response or DOM
    unsafe deserialise    file containing the call that deserialises untrusted bytes
    missing auth check    file containing the handler that should perform the check

If a vulnerability has more than one sink, mint one VID per sink and link them as aliases in the surrounding record. Record other involved files (taint sources, helpers, intermediate transforms) as metadata alongside the VID.

A vulnerability that is a property of the package as a whole, with no file that would be changed to fix it, is out of scope for v1.

## Construction

Concatenate with a single `\n` (0x0A) separator and no trailing newline:

    purl 0x0A gitoid

Hash the UTF-8 bytes with SHA-256. Take the first 15 bytes. Encode as lowercase base32 (RFC 4648 alphabet `a`-`z` `2`-`7`, no padding), giving 24 characters. Insert a hyphen every four characters and prefix with `VID-1-`:

    VID-1-xxxx-xxxx-xxxx-xxxx-xxxx-xxxx

The `1` is this spec version. Any change to inputs, normalisation, hash, or encoding increments it.

## Test vector

    file bytes      68 65 6c 6c 6f 20 77 6f 72 6c 64 0a    ("hello world\n", 12 bytes)
    gitoid          0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d
    purl            pkg:generic/example

    vid preimage    "pkg:generic/example\n0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d"
    sha256          e16606c09d7bcf3d33d1d4dea777e2de1e470223f58587593f010627eaf4de2b
    first 15 bytes  e16606c09d7bcf3d33d1d4dea777e2

    VID-1-4fta-nqe5-ppht-2m6r-2tpk-o57c

A second vector exercising purl case normalisation. The npm type lowercases the name, so `pkg:npm/Lodash` MUST normalise to `pkg:npm/lodash` before hashing:

    purl input      pkg:npm/Lodash
    purl normal     pkg:npm/lodash
    gitoid          0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d
    sha256          3067697d3e4fe14ef37d9012a9f71e2631543c3755a62ca763efa52133656153

    VID-1-gbtw-s7j6-j7qu-5435-sajk-t5y6

## Properties

The gitoid pins a VID to exact file content. Editing the file produces a different VID. Renaming or moving the file does not, because the gitoid depends only on bytes. The same file appearing unchanged across several released versions yields one VID, which is why the purl carries no version.

The 120-bit truncation gives second-preimage resistance of roughly 2^120. Accidental collision is not a practical concern at any plausible number of identifiers.

A VID never expires. If a later release reintroduces the same vulnerable bytes after a fix, the original VID matches the regression. The surrounding record tracks which versions are fixed.

Anyone can compute a VID for any file in any package whether or not a vulnerability is present. The identifier names the location of a claim and says nothing about whether the claim is true.

Publishing a VID does not reveal its inputs, but the preimage space is small: someone who can guess a plausible package and file can hash them and check for a match. Given a claimed preimage, anyone can recompute the VID and verify it.

File selection is the one place this scheme depends on shared convention. Two researchers who choose different files for the same underlying vulnerability will compute different VIDs. This is accepted: the VIDs are linked as aliases in the surrounding record, the same way advisory databases already link CVE, GHSA, and vendor identifiers for one issue.

## Status

Draft. Before this is stable: a reference purl normaliser must be named and pinned; more test vectors are needed covering each purl type's case rules; a reference CLI should exist; and a record format for what travels alongside a VID (CWE, severity, affected versions, aliases) should be sketched, with the OSV schema's `aliases` field as the likely target.
