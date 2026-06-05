# VID

A way to identify a piece of code that someone is claiming is security-relevant. The scheme deliberately identifies code rather than vulnerabilities, because independent parties can reliably agree on bytes long before they can reliably agree on severity, exploitability, CWE classification, affected packages, or even whether a finding is real.

The identifier is a hash of the bytes being discussed, computed independently by anyone with access to the same code, so any two parties looking at the same code arrive at the same identifier without coordinating and without an issuing authority. Exchanging identifiers between parties reveals nothing about the underlying code, only whether the parties have overlapped. Everything else (which package the code came from, who is affected, what the weakness is, even whether the claim is correct) travels alongside as metadata.

The target is open source libraries pulled in as dependencies by other projects. The most direct application is scanner interoperability, where several scanners running on the same code emit VIDs and downstream consumers correlate findings across vendors without inventing a vendor-by-vendor mapping. The same primitive supports researcher coordination, maintainer dedup, and consumer-side matching against installed code, and the construction itself works on anything with a file and a sink line.

Findings that span several functions fragment into one VID per function, and that is where this convergence story is weakest.

- [SPEC.md](SPEC.md): the construction
- [examples/](examples/): real advisory fixtures used as the test corpus
- [FAQ.md](FAQ.md): common questions

## Install

    go install github.com/andrew/VID/cmd/vid@latest

The CLI takes one or more `file:line` arguments and prints the VID to stdout:

    $ vid examples/GHSA-vh95-rmgr-6w4m/vuln.js:73
    VID-yssz-i3ln-4ob6-z2fx-3rpt-jve3

A VID is the literal string `VID-` followed by six groups of four lowercase base32 characters separated by hyphens, 33 characters in total. Multiple sinks combine into one multi-sink VID, order-independent:

    $ vid examples/CVE-2014-0160/_vuln_dtls.c:1455 examples/CVE-2014-0160/_vuln_tls.c:2554
    VID-yqay-acdk-fxsp-43nl-47bv-lylk

Setting `VID_DEBUG=1` prints the preimage and per-sink mode, language, and OID to stderr:

    $ VID_DEBUG=1 vid examples/GHSA-vh95-rmgr-6w4m/vuln.js:73
    VID-yssz-i3ln-4ob6-z2fx-3rpt-jve3
    preimage: fc9cbcca1f66640a92003ac2eb5a8a1665a15d6eed5af2124dce8359d5466dd0
    sink examples/GHSA-vh95-rmgr-6w4m/vuln.js:73  mode=function  lang=javascript  oid=fc9cbcca1f66640a92003ac2eb5a8a1665a15d6eed5af2124dce8359d5466dd0

## The problem

Finding vulnerabilities in open source used to be slow, skilled work. A researcher might spend a week on one library and come away with one good bug, and the systems built around that pace (report it privately to the maintainer, wait for a fix, request a CVE number, publish an advisory) assumed findings were rare and each one got individual attention.

That assumption no longer holds, because automated tools, particularly ones built on large language models, can audit a codebase in minutes and flag real weaknesses at a rate no human reviewer could match. Anyone running such a tool against open source for a few days will accumulate more findings than they could disclose in a year, and popular libraries are being scanned by many such people at once.

When several of those tools run against the same library they flag largely the same bugs, so a maintainer ends up with five email threads about one problem from five well-meaning researchers. After a few rounds of this the maintainer stops reading security reports; maintainers of widely-used projects are already [writing about the influx](https://daniel.haxx.se/blog/2026/04/22/high-quality-chaos/).

Writing a clear report, finding the right contact, and following up takes far longer than finding the bug did, so reporting itself becomes the bottleneck and a backlog accumulates faster than any individual can responsibly work through. Most of those findings never reach the maintainer or the users.

Users of the library are blind until publication: a company that depends on the library will eventually see a CVE in their scanner, months after the bug was first found, and only then start working out whether the vulnerable code is actually reachable from their application. Usually it is not, but proving that under time pressure is its own cost.

The CVE system itself is strained, with identifier assignment going through a small number of authorities, enrichment of published records backlogged, and the funding behind the central database publicly unstable. Building more process on top of it is not appealing.

## What a VID is

A VID is computed rather than assigned: a tool locates the function being pointed at, hashes its bytes, and the resulting hash, encoded into a short string, is the identifier. It looks like this:

    VID-r7m4-3xqj-6p2k-3wfn-5tza-h6vc

Because the input is bytes, anyone with the same function bytes computes the same VID, so two researchers analysing the same code converge on the same string without coordinating, and a scanner emits the VID alongside its findings the same way it would emit a line number. If the sink is at module scope or in a language with no bundled grammar, the construction falls back to hashing the whole file and the result is still a `VID-...` string with the same shape.

The hash is over literal bytes, without normalisation, so any edit inside the function (whitespace, comments, a reformatter pass, a renamed local, CRLF instead of LF, git's autocrlf conversion on a cross-platform checkout) changes the VID. Two byte-identical copies of a function converge across packages, files, and forks; copies that differ only by formatting do not.

This is the load-bearing empirical bet of the whole scheme: that registry tarballs, git checkouts, and distro repackagings of the same release routinely produce byte-identical function bytes in practice, and that reformat churn between versions is rarer than the parser-version churn an AST-based scheme would inherit.

The alternatives are hashing a canonicalised tree-sitter AST, or light normalisation such as LF-only line endings and trimmed trailing whitespace. Both resist some reformatting at the cost of either pinning every VID to a specific grammar version or carrying their own normalisation logic that every consumer must reproduce. The scheme picks grammar-stability over reformat-resilience.

Ecosystems where the published archive is a build of the source (npm with bundlers, Python with compiled wheels, Java with jars) introduce a related complication: the registry function bytes and the repository function bytes diverge, and a VID computed from the source produces a different identifier than a VID computed from the published artefact. One logical finding can therefore carry two VIDs depending on which artefact a researcher or scanner looked at. The two are linked as aliases in whatever record carries them, and consumer-side matching against installed code is performed against the installed bytes (which match the registry artefact).

Beyond the bytes themselves, nothing enters the hash: not the package name (its purl, like `pkg:npm/lodash`), not the file path, not the line number, not the CWE, severity, affected version range, or data flow analysis. All of those travel with the VID as metadata in whatever record is kept, because the set of purls that cover a given piece of code (registry copies, repo copies, forks, vendored copies, private mirrors, downstream redistributions across distros) is open and unknowable, and two researchers cannot be expected to agree on which one to pick.

A VID is to a vulnerability report what a git blob hash is to a source file: it names the bytes being discussed, and the interpretation of those bytes belongs to other layers. The byte-hash itself, `sha256("blob " <len> 0x00 <bytes>)`, is git's blob hash construction, also standardised by [OmniBOR](https://omnibor.io) for identifying software artefacts without an authority. In VID the construction is applied to a tree-sitter-extracted function range rather than a whole file, with an encoding chosen for short readable identifiers. The novel part is the convention of using the result as a coordination key across tools, researchers, and consumers, rather than the hash itself; [SWHID](https://www.swhid.org) identifies source artefacts the same way for archival, and SARIF fingerprints and similar per-tool hashes dedup findings within a single scanner's output but are not designed to match across tools or people.

Which function counts as the sink is the judgment call the scheme still depends on, and it matters more than it looks. A taint flow that runs from `parseInput` through `buildQuery` to `db.exec` gives three valid sinks, and analysts naturally anchor on different points in the chain and produce three different VIDs for what is conceptually the same finding. Other ambiguous cases include a wrapper and the function it wraps, one of several near-identical template renderers, or a callsite and the helper it calls.

Where two researchers pick the same function their VIDs match. Where they pick differently the resulting VIDs are linked as aliases, the way advisory databases link CVE and GHSA, with the difference that the link can live in whatever record a consumer chooses to keep rather than requiring a central authority.

The function range itself depends on a specific tree-sitter grammar producing the same boundaries for each function, which is the most stable thing a parser produces but is not authority-free, and pinning a grammar means a grammar bump is also a spec bump.

A report from a researcher to a maintainer carries the VID at the top:

    VID:      VID-r7m4-3xqj-6p2k-3wfn-5tza-h6vc
    Package:  pkg:npm/example
    File:     lib/query.js
    Line:     42
    CWE:      CWE-89

    User-supplied `req.params.id` reaches `db.query()` without parameterisation...

The first line is computed and the rest is the researcher's analysis, and a maintainer can dedup against earlier reports on the first line alone.

## What it lets you do

Two researchers with overlapping private finding sets have no good way to learn that overlap without sharing the underlying details first. With VIDs each side hashes their candidate functions locally and exchanges only the resulting identifiers, and a match indicates convergence on the same code while the rest of each list is what is uniquely held. Every existing mechanism requires revealing at least the file path or the package being analysed before overlap can be established.

When several automated scanners run on overlapping code by different people and flag overlapping functions as security-relevant, matching VIDs across those flags are independent confirmation that the scanners pointed at the same code, which carries more weight than any single scanner's hit rate. The scanners can still disagree about severity, CWE, exploitability, or whether the finding is real at all, and a matching VID does not collapse those disagreements; it only establishes that the same code is under discussion. Scanner vendors are also spared inventing vendor-specific identifier schemes, and downstream consumers correlate findings across vendors without a vendor-by-vendor mapping.

A maintainer triaging incoming reports sees at a glance which carry duplicate VIDs and which are distinct, so five reports about the same bug carrying the same VID are visibly four duplicates without any infrastructure beyond the convention that reports include the identifier.

Publishing the VID alone (commit message, signed log entry, public timestamp service) binds a researcher to having found something at time T without revealing what they found, and when the eventual CVE publishes the inputs the timestamp is verifiable. The VID is not a cryptographic secret, since someone with a corpus of candidate functions can hash each and check matches, but the bytes are not revealed by the act of timestamping alone.

A consumer with a set of VIDs from a trusted researcher or scanner output can hash every function in their installed dependency tree and check for matches against the set, and where a match lands, that is the function to examine for reachability from the application's own entry points. Most matches turn out to be unreachable, and the "we're fine" answer is available weeks or months before the corresponding CVE publishes.

## What it is not

VID is a pre-CVE coordination layer, not a replacement for CVE or a publication format. A VID exists from the moment of finding and aliases to a CVE once one is assigned, publishing a vulnerability remains the province of CVE and GHSA, and VID is used upstream of that in the private and pre-disclosure phase.

A published VID is a soft commitment rather than a sealed envelope. The 120-bit truncation makes brute-forcing arbitrary preimages infeasible, but for any specific library at a specific version the candidate corpus is small and fully public, and anyone who guesses the package can hash every function in it in seconds and find the match. Timestamping still works because the timestamp survives the eventual disclosure, but secrecy is weak against anyone willing to scan the obvious suspects.

Some classes of vulnerability fit the model badly: a bug that consists of an absent check or call has no canonical bytes to point at, and researchers describing the same missing-check finding may pick different enclosing functions and converge less reliably. Bugs in configuration files, build descriptors, algorithm choice, or default values stretch the model further, and the construction is sharpest for vulnerabilities with a definite sink such as a dangerous call, a tainted write, or a misused API.

Adversarial use is out of scope. The model assumes participants acting in good faith. Attackers trying to mint colliding identifiers or precompute VIDs over public code to claim credit are not defended against, and the model also does not address coordinated mis-attribution between parties exchanging lists.

The same bytes in multiple packages produce one VID, and the package mapping lives in the surrounding record rather than in the identifier itself. When to tell the maintainer, when to publish, and what to share with whom remain human decisions outside the spec's scope, and advisory databases are welcome to carry VIDs as aliases (the scheme is more useful if they do), but the construction depends on no database existing.

A VID identifies the code being claimed about, with the truth or importance of the claim sitting at a different layer. A low-priority finding has a VID as readily as a critical one, and a VID can equally be computed for a false positive; the identifier asserts that someone is making a claim about this code, and the truth of the claim is established the same way it always was. Deduplication reduces the count of redundant reports without reducing the count of unimportant ones, but a maintainer is still materially better off when four visible duplicates collapse to one at the top of the inbox even with the total noise volume unchanged.

## Status

The construction is in [SPEC.md](SPEC.md). The repository carries a Go reference implementation (CLI and library) and a test corpus of 22 real advisories under [examples/](examples/), each with a vulnerable file (or files), a fixed file (or files), and a perturbation file used to assert that edits outside the sink function leave the VID unchanged.

Coverage is two advisories each across JavaScript, TypeScript, Ruby, Python, Go, Rust, PHP, Java, C, and C++, plus two multi-sink advisories (Heartbleed and runc CVE-2024-21626) where the patch lands in two parallel code paths.

The corpus validates boundary stability across grammar versions and out-of-function invariance against perturbation. The cross-artefact byte-identity claim — that registry tarballs, git checkouts, and distro repackagings of the same release routinely produce byte-identical function bytes — is asserted by the construction but is not measured by the corpus.

Multi-sink is the place this convergence story is weakest. The moment a vulnerability spans more than one function the parties have to agree on the set of sinks before they converge, which puts them back in the coordination problem the single-function case escapes. The library combines multiple sinks by sorting their OIDs, deduplicating, and joining with `\n` in the preimage, and the two multi-sink fixtures verify that the combined VID is stable across edits outside any of the sink functions.

## License

MIT. See [LICENSE](LICENSE).
