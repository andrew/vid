# FAQ

Questions that came up during review. See [README.md](README.md) for the rationale and [SPEC.md](SPEC.md) for the construction.

**Is this trying to replace CVE?**
No. VID is a pre-CVE coordination layer. A VID exists from the moment of finding; once a CVE is assigned the VID record carries it as an alias and the CVE is what everyone cites publicly.

**A whitespace change to the file produces a new VID. Isn't that too fragile?**
It is precise rather than fragile. The identifier says "these exact bytes in this package," and if the bytes change the claim needs re-checking. The old and new VIDs are linked as aliases in the surrounding record, so nothing is lost; you gain certainty about which code a given VID actually refers to.

**Why isn't the CWE part of the hash?**
Because two researchers looking at the same bug often pick different CWEs, and any subjective input breaks the property that independent finders converge on the same identifier. CWE travels as metadata alongside the VID.

**Why is there no version in the purl?**
The file hash already pins the content. If the same vulnerable file ships unchanged in five releases, one VID covers all five. Affected version ranges belong in the surrounding record.

**How is this different from OmniBOR?**
OmniBOR's gitoid identifies a file's bytes wherever they appear. VID uses that same hash as one input but adds the package name, because the unit of disclosure and triage is the package, and the same bytes vendored into fifty packages means fifty maintainers. VID also names a vulnerability location specifically; a gitoid on its own names an artefact.

**Can someone brute-force a published VID to find out what it refers to?**
Often yes, for popular packages. The preimage is a package name plus a file hash, and someone who suspects which package is involved can hash each of its files and check. A VID is a commitment that binds you to a finding; it does not hide the finding from a determined guesser.

**What about JavaScript packages with a build step?**
The repository file and the published file have different bytes and therefore different VIDs. The published archive is the reference for consumer-side matching. A researcher scanning the repository computes one VID; tooling can map it to the published-archive VID, or the record carries both as aliases.

**Two researchers pick different sink files for the same bug. Now what?**
They have two VIDs for one issue. The VIDs are linked as aliases in the surrounding record, the same way a CVE and a GHSA for one issue are linked today. In practice automated scanners that converge on the same bug tend to converge on the same sink file, so this is less common than it sounds.

**What about a vulnerability that's a property of the whole package, with no single file to point at?**
Out of scope for v1. An earlier draft used the package manifest as a fallback, but that collides unrelated package-level findings onto one VID, which is worse than admitting the scheme doesn't fit.

**Doesn't GitLab have a patent on vulnerability fingerprinting?**
US 11,868,482 covers hashing a scope, an offset, and a classifier. VID hashes a package name and a content hash, with no location offset and no classifier. They are different constructions. This is an observation about the published claims and is not legal advice.

**Where do VID records live?**
Wherever you put them. There is no required database. The expectation is that existing advisory databases will carry VIDs as one more alias, and that researchers and consumers will hold their own sets locally.
