# FAQ

See [README.md](README.md) for the rationale and [SPEC.md](SPEC.md) for the construction.

**Is this trying to replace CVE?**
No. VID is a pre-CVE coordination layer. A VID exists from the moment of finding; once a CVE is assigned, the VID becomes one of the CVE's aliases and the CVE remains the public name everyone cites. VID handles the part of the workflow that happens before disclosure (dedup, private comparison, scanner interoperability, consumer matching against installed code) which CVE was never designed for and is increasingly overwhelmed by.

**What does the lifecycle of a finding with a VID look like?**
A scanner or researcher identifies a function as security-relevant and computes its VID. The VID can be logged privately, used internally for dedup, exchanged with another researcher to discover overlap without revealing the underlying code, or sent to the maintainer as part of a report. The maintainer triages incoming reports by VID, sees duplicates collapse to a single distinct entry, fixes the function (which changes its bytes, and therefore its VID), and records that the old VID is patched while the new bytes carry a new VID. If the maintainer or a CNA pursues a CVE, the VID becomes one of the CVE's aliases. Consumers running scanners against their installed dependency trees emit their own VIDs and match them against known-vulnerable VIDs, and matches identify the specific functions in the tree that need reachability analysis from the application's own entry points.

**Why does VID identify code rather than vulnerabilities?**
Because independent parties can reliably agree on bytes long before they can reliably agree on whether those bytes constitute a vulnerability. Two scanners can flag the same function and disagree about CWE, severity, exploitability, or whether the finding is real at all, while still emitting the same VID and so making their disagreement legible. Identifying "the vulnerability" directly would require an input expressing someone's judgment about severity or weakness class, and the convergence property would dissolve.

**Why is the package (purl) not part of the hash?**
Because the set of purls that cover any given piece of code is open and unknowable: registry copies, repo copies, forks, vendored copies, private mirrors, downstream redistributions across distros. Two researchers cannot be expected to agree on which one to pick, and including any subset would be arbitrary. Hashing only the bytes makes convergence depend on the code itself, and the same vulnerable function copied into many packages produces one VID matching every copy. The purl travels alongside the VID as metadata in any record kept against it.

**Why is the CWE not part of the hash?**
For the same convergence reason as the purl: two researchers looking at the same bug routinely pick different CWEs, and any subjective input breaks the property that independent finders compute the same identifier. CWE travels as metadata in the surrounding record.

**Why function bytes, and not the whole file?**
Hashing the whole file means any edit anywhere in the file (an unrelated helper added, an import reordered, a formatter pass) produces a new VID, so one bug appears as several across releases. The function is the actual vulnerable unit and stays stable when unrelated code in the same file changes. The cost is that the construction needs a parser to find the function boundary; the spec uses tree-sitter and falls back to the whole-file hash when no enclosing function is available (module-scope sinks, languages tree-sitter does not support).

**A whitespace change inside the function produces a new VID. Isn't that fragile?**
Yes, and this is the load-bearing empirical bet of the scheme: that byte-identical function bytes show up routinely across registry tarballs, git checkouts, and distro repackagings of the same release in practice, and that reformat churn between versions is rarer than the parser-version churn an AST-based scheme would inherit instead. If the test corpus shows the bet is wrong (real-world copies of the same release diverge by formatting often), the construction shifts to light normalisation (LF-only line endings, trimmed trailing whitespace) or AST-based hashing. The current default is "literal bytes, no normalisation" because that survives grammar upgrades unchanged and is the simplest construction to reimplement from scratch.

**What about JavaScript packages with a build step?**

The repository file and the published file have different bytes and therefore different VIDs. The published archive is the reference point for consumer-side matching against installed code, since that is what dependents actually run. A researcher scanning the repository computes one VID; the same finding scanned against the published archive computes another; the two are linked as aliases in whatever record carries them.

The same applies to any ecosystem where the published artefact differs from the source repository: bundled or minified npm packages, packages that compile C extensions into wheels, Java jars containing compiled `.class` files.

For ecosystems where the published artefact is the source unchanged (most Go, Ruby, pure-Python), the repository VID and the registry VID coincide *when the bytes coincide*. That is not automatic. A Go module proxy serves zip-packed source verbatim, so line endings the author committed survive; a checkout with `git autocrlf=true` on Windows can produce CRLF where the proxy has LF, and the VIDs diverge. The "source unchanged" property holds at the artefact level, not at every local checkout.

**Two researchers pick different functions for the same bug. Now what?**
They produce two different VIDs for one conceptual finding. The two are linked as aliases in any record either party keeps, and any consumer who has both records sees the link. The aliasing is local knowledge, not universal truth, because the construction does not depend on a central database equivalent to the CVE/GHSA registry: two consumers can have different views of what is aliased to what, depending on which records they have collected. In practice automated scanners that converge on the same bug tend to converge on the same sink function (they share biases), so this matters most when a human and a scanner anchor on different points in a taint chain.

**How is this different from OmniBOR, SWHID, and SARIF?**
OmniBOR's gitoid is the same hash construction VID uses internally, but OmniBOR identifies a whole file's bytes and does not extract a function range or address vulnerability coordination workflows. SWHID identifies source artefacts for archival, again at the file or commit level rather than the function level. SARIF fingerprints dedup findings within one scanner's output but are not designed to match across tools or people. VID is OmniBOR-style content addressing applied to function-extracted byte ranges, with the convention of using the result as a coordination key across tools, researchers, maintainers, and consumers.

**Can someone brute-force a published VID to find out what it refers to?**
For popular packages, yes. The preimage is a function's bytes, and someone who suspects which library the finding is in can extract every function from that library and hash each one in seconds. The 120-bit truncation makes brute-forcing the bytes from scratch infeasible, but the candidate space for any specific library at a specific version is small and fully public. A published VID is a soft commitment that binds the publisher to a finding at time T; it does not hide the finding from a determined guesser. Against a well-known target, a published VID is closer to a pointer than a sealed envelope.

**What about a vulnerability that has no specific code location?**
Out of scope. If the issue is a property of the package as a whole (an insecure default applied at install time, a missing security feature with no specific file that would change to fix it), the construction has nothing to hash that converges. Falling back to the package manifest would collide unrelated package-level findings onto one VID, which is worse than admitting the scheme does not fit. Such findings are handled at the package level by CVE and GHSA.

**Where do VID records live?**
Wherever the people using them put them. There is no required database. Existing advisory databases (OSV, GHSA, vendor records) can carry VIDs as one more alias if they choose, but the construction does not depend on any database existing. Researchers and consumers commonly hold local sets, and scanner vendors emit VIDs alongside their findings without needing a central registry.

**Doesn't GitLab have a patent on vulnerability fingerprinting?**

The relevant patent family includes US 11,868,482 ("scope and offset fingerprints") and the sibling US 12,086,271 ("smatch values of scopes").

Claim 1 of 11,868,482 recites a fingerprint generated from a scope and an offset, where the offset is computed by subtracting the sink line from the start line of the scope. The fingerprint changes as the sink moves within its enclosing scope. VID has no offset and no line arithmetic at all: it hashes the literal bytes of the enclosing function via the git blob construction, and the same finding at any line inside that function produces the same VID. That offset requirement is the cleanest claim-level distinction.

US 12,086,271 covers identifying a minimum enclosing scope by a smatch value computed from line arithmetic. The mechanism differs from VID's tree-sitter ancestor walk to the nearest function-like node, though both narrow to a smallest enclosing unit at the sink line.

This is an observation about the published independent-claim language of the two patents, not legal advice. An implementer with adoption risk should consult counsel; dependent claims and continuations are not analysed here.
