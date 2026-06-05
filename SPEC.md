# VID specification

This document describes how a VID is computed. The reference implementation in this repository is the source of truth.

## Identifier format

A VID is the string `VID-` followed by six groups of four characters separated by hyphens:

    VID-xxxx-xxxx-xxxx-xxxx-xxxx-xxxx

Each character is one of `a`-`z` or `2`-`7`, drawn from the RFC 4648 base32 lowercase alphabet without padding, and the total string length is 33 characters including the prefix and hyphens. There is no version number in the identifier; the construction is treated as fixed until the spec stabilises, and any breaking change to the inputs or encoding will be addressed by reopening the question of whether to add a version prefix at that point.

## Inputs

The construction takes one or more sinks. Each sink is a triple:

    sink = (filename, source_bytes, sink_line)

`filename` is used to detect the source language by extension. `source_bytes` is the literal byte content of the file. `sink_line` is a 1-based line number naming the line a researcher or scanner is pointing at as security-relevant. Nothing else enters the computation: the package name, file path within the package, line number value, CWE, severity, affected version range, and any analysis travel with the VID as metadata in whatever record is kept against it, but none of them affect the resulting identifier.

## Construction

For each sink, the construction produces a 64-character lowercase hex OID via one of two paths.

### Function path

If the file extension maps to a supported language, the source bytes are parsed with the corresponding tree-sitter grammar. Starting from the first non-whitespace column of the sink line, the implementation queries the parse tree for the smallest named descendant containing that point, then walks up to the nearest ancestor whose node type matches the per-language function-node table (below). If such an ancestor is found, its byte range `[start, end)` in the source is the *function bytes*, and the OID is

    sha256("blob " <bytelen> 0x00 <function_bytes>)

expressed as 64 lowercase hex characters. This is the construction git uses for blob object IDs and that OmniBOR standardises for content-addressed software artefacts.

The hash is over the literal bytes of the function range, with no normalisation: whitespace, comments, line endings, and identifier names all enter the hash as they appear in the source.

The exact value of `sink_line` within the function does not affect the OID. The line only determines which function the ancestor walk lands in; the OID is computed over the bytes of that function. Two sinks at different lines of the same function produce the same OID, which is the property the combining step (below) relies on for deduplication.

### File fallback

If the file extension is unsupported, the parser cannot produce a tree, or no ancestor matches the function-node table at the sink line, the construction falls back to hashing the whole file:

    sha256("blob " <bytelen> 0x00 <file_bytes>)

The encoding is the same 64-character lowercase hex, and the fallback ensures the construction is defined for every sink, including module-scope statements and languages with no grammar bundled.

### Combining sinks into a VID

Given N OIDs from N sinks, the implementation sorts the OIDs lexicographically by byte (identical to ASCII sort, since OIDs are lowercase hex) and removes consecutive duplicates. Two sinks resolving to the same function, or to identical bytes across files, contribute one OID rather than several.

The resulting OIDs are joined with `\n` (0x0A) as a single ASCII string. That string is the *preimage*. The VID is then

    sha256(preimage)[:15]

encoded in lowercase base32 (RFC 4648 alphabet `a`-`z`, `2`-`7`, no padding) producing 24 characters, broken into six groups of four with `-` separators, and prefixed with `VID-`.

The single-sink case falls out as n=1: the preimage is just the one OID as a 64-character hex string, and the resulting VID identifies that one sink. Two sinks at different lines of the same function are also the n=1 case after deduplication.

## Function-boundary detection

The implementation locates the enclosing function with tree-sitter. The exact procedure:

1. Extract the file's basename and look up the extension in this table:

   | Extension | Language |
   |---|---|
   | `.js`, `.jsx`, `.mjs`, `.cjs` | javascript |
   | `.ts`, `.d.ts` | typescript |
   | `.tsx` | typescript (TSX variant) |
   | `.rb` | ruby |
   | `.py` | python |
   | `.go` | go |
   | `.rs` | rust |
   | `.java` | java |
   | `.c`, `.h` | c |
   | `.cpp`, `.cc`, `.cxx`, `.hpp` | cpp |
   | `.php` | php |

   Any other extension means no language is detected and the file fallback applies immediately.

2. Parse the source bytes with the language grammar.

3. Compute the column of the first non-whitespace character on the sink line. This handles languages where significant indentation places the tree-sitter cursor outside the indented block when querying at column zero.

4. Query the parse tree for the smallest named descendant at point `(sink_line - 1, first_non_whitespace_column)`.

5. Walk up the tree through `Parent()` until either the root is reached (no match: fall back to file) or a node whose type is in the function-node table for the detected language is reached. The first matching ancestor wins, and its byte range is the function bytes.

The grammar bindings used are the ones distributed under github.com/tree-sitter/tree-sitter-LANG, pinned via Go module versions in this repository's go.mod. A grammar bump is a spec bump: if a new grammar version reports different boundaries for any function in the test corpus, the spec version increments and old VIDs become uncomputable by tools on the new grammar.

## Function-node table

For each language, the construction treats an ancestor as a function if its tree-sitter node type appears in this set:

| Language | Function node types |
|---|---|
| javascript | `function_declaration`, `function_expression`, `generator_function_declaration`, `arrow_function`, `method_definition` |
| typescript | `function_declaration`, `function_expression`, `generator_function_declaration`, `arrow_function`, `method_definition` |
| ruby | `method`, `singleton_method` |
| python | `function_definition` |
| go | `function_declaration`, `method_declaration` |
| rust | `function_item` |
| java | `method_declaration`, `constructor_declaration` |
| c | `function_definition` |
| cpp | `function_definition` |
| php | `function_definition`, `method_declaration` |

Languages with closures, lambdas, or anonymous functions can have multiple matching ancestors; the innermost match wins because tree-sitter ancestor walking returns the closest first.

## Test vectors

Each vector below is verified by the test suite.

### File-fallback vector

Input bytes: `hello world\n` (12 bytes, `0x68 0x65 0x6c 0x6c 0x6f 0x20 0x77 0x6f 0x72 0x6c 0x64 0x0a`).

```
gitoid    = sha256("blob 12\x00hello world\n")
          = 0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d
preimage  = gitoid                  (single sink, file fallback)
sha256    = e16606c09d7bcf3d33d1d4dea777e2c2...
truncate  = first 15 bytes
encode    = ernd c2qt gs7l lkhg 32lv y4li
VID       = VID-ernd-c2qt-gs7l-lkhg-32lv-y4li
```

### Function-bytes vector (minimist 1.2.2 `setKey`)

The minimist npm package at version 1.2.2 contains the function `setKey` in `index.js` at line 69. Pointing any sink line in that function (e.g., line 73) at that file produces the golden VID recorded for [examples/GHSA-vh95-rmgr-6w4m](examples/GHSA-vh95-rmgr-6w4m):

```
function bytes = src[node.StartByte:node.EndByte] for the function_declaration node spanning lines 69-87
funcoid        = fc9cbcca1f66640a92003ac2eb5a8a1665a15d6eed5af2124dce8359d5466dd0
preimage       = funcoid                  (single sink, function mode)
VID            = VID-yssz-i3ln-4ob6-z2fx-3rpt-jve3
```

### Multi-sink vector

[examples/CVE-2014-0160](examples/CVE-2014-0160) (Heartbleed) is a two-sink fixture: the DTLS heartbeat handler `dtls1_process_heartbeat` in `ssl/d1_both.c` at line 1455 and the TLS heartbeat handler `tls1_process_heartbeat` in `ssl/t1_lib.c` at line 2554. Each resolves in function mode:

```
oid_dtls = 18498153b671a2e767ec5abb0190c111f6965ed28cc374f7beeaac16d75af286
oid_tls  = d0bf5dba7fec4dabb6f6ce1b12b7f5c79ca13d9921f266512a049aa17f6ed3ed
```

The sorted, deduped, newline-joined preimage is

```
18498153b671a2e767ec5abb0190c111f6965ed28cc374f7beeaac16d75af286
d0bf5dba7fec4dabb6f6ce1b12b7f5c79ca13d9921f266512a049aa17f6ed3ed
```

since `18...` sorts before `d0...`, and the resulting VID is `VID-yqay-acdk-fxsp-43nl-47bv-lylk`. Passing the two sinks to the CLI in either order produces the same string.

## Properties

### Stability and convergence

Two parties holding the same function bytes compute the same VID. The construction has no inputs the parties need to agree on beyond the bytes themselves, and the set of purls covering any given piece of code is open and unknowable, so the purl is not part of the hash.

The VID is stable across file renames and moves (only the bytes enter the hash), unrelated edits elsewhere in the file (the function path hashes only the function range), changes to per-package metadata (not an input), and changes to which package the file ships in (the same vulnerable function copied into many packages produces one VID matching every copy).

The VID changes when any byte inside the function range changes, including whitespace, comments, renamed locals, reformatter passes, line-ending changes, or `git autocrlf` translations on cross-platform checkouts.

It also changes when the function is moved to a file in a language with no bundled grammar, because the construction would shift from the function path to the file fallback and hash a different range of bytes.

The pinned tree-sitter grammar version can in principle change the byte range it reports for a given function, which would also change the VID. Boundary detection is the most stable property of these grammars in practice, but the test corpus does not yet measure cross-grammar boundary drift directly.

### Collision resistance

Truncating the SHA-256 digest to 120 bits leaves a collision-finding work factor of about 2^60 and a preimage-finding work factor of about 2^120 against arbitrary inputs. Accidental collision between unrelated functions is not a practical concern at any plausible corpus size.

Collision against a *known* target is cheaper. Someone who suspects which package and version a VID refers to can hash every function in that package and check matches, and for popular libraries the candidate space is small and fully public. A published VID is therefore closer to a pointer than a sealed envelope, which is acknowledged and intentional.

### What the construction does not do

The construction names code rather than packages. The same bytes copied into many packages produce one VID, and the mapping from VID to affected packages lives in whatever record is kept against it.

It names code rather than vulnerabilities. The VID identifies code being claimed about; the truth or severity of the claim comes from a different layer.

No database is required. Aliases between VIDs (different sinks for one bug, different artefact versions of one file) sit in whatever record a consumer chooses to keep, not in a central registry.

The model assumes good-faith participants and does not defend against adversaries trying to mint colliding identifiers or precompute VIDs over public code to claim credit.

Byte normalisation is not performed. The hash is over literal source bytes; LF-only line endings, trimmed trailing whitespace, and AST-based canonicalisation are all rejected in favour of grammar-version stability.

## Reference implementation

The reference implementation in this repository is a Go library and CLI. Library entry points:

```
vid.Gitoid(b []byte) string
vid.Encode(preimage string) string
vid.EnclosingFunction(src []byte, filename string, line int) (fn []byte, lang string, ok bool)
vid.Compute(sinks []vid.Sink) vid.Result
```

`Compute` is the high-level entry point used by the CLI and by the test runner. The CLI takes one or more `file:line` arguments and prints the VID:

    $ vid lib/query.js:42
    VID-...

    $ vid lib/query.js:42 lib/audit.js:88
    VID-...                        # multi-sink combined VID

Setting `VID_DEBUG=1` prints the preimage and per-sink (mode, language, OID) to stderr alongside the VID, which is what test vectors are confirmed against.

The test corpus lives in [examples/](examples/), with each fixture as a directory containing a vulnerable file, a fixed file, a perturbation file (the vulnerable file with edits outside the sink function), and an `advisory.yaml` manifest with the golden VID. Multi-sink fixtures use the plural `sinks`, `fixed_sinks`, and `perturbation_sinks` forms in place of the singular fields.

The test runner walks `examples/*/advisory.yaml`, computes the VID for each sink (or sink list), and asserts that the sink VID matches the golden, that the fixed VID differs from the vulnerable VID, and that the perturbed VID equals the vulnerable VID. What the corpus validates is therefore boundary stability across grammar versions and out-of-function invariance against perturbation. The cross-artefact byte-identity claim (registry tarball vs. git checkout vs. distro repackaging of the same release) is asserted by the construction but not yet measured by the corpus.
