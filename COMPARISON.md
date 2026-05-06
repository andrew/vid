# v1 file-hash vs v2 function-bytes vs v2 function-AST

## Summary

Both function-level variants are strictly better than v1 at their stated goal: on the worked example, four still-vulnerable releases collapse to one identifier under v2-AST, two under v2-bytes, three under v1. The cost is a tree-sitter dependency and one more judgment call (which function, on top of which file).

v2-bytes is the safer bet. The grammar is only used to find the function boundary, so the hash is plain source bytes that anyone can recompute with any tool, and grammar upgrades don't invalidate existing identifiers. It loses to v2-AST only when the vulnerable function itself is reformatted without being fixed, which is a narrow case.

v2-AST is the most stable across cosmetic edits but pays for it by baking a specific grammar blob hash into every identifier. Two researchers on different gotreesitter releases will not converge, and a grammar bump reissues every identifier in that language. That is manageable if the spec pins one grammar set per `VID-N` version and treats a grammar bump as a spec bump, but it ties the whole scheme to one parser implementation in a way v1 deliberately avoided.

Recommendation: ship v2-bytes as the function-scoped construction, keep v1 as the fallback for module-scope sinks and unsupported languages, and park v2-AST unless reformatting churn turns out to be a real source of identifier drift in practice.

## Results

Run against the five `acme-greeter` releases in `examples/`. The sink is the `exec` call inside `greet`.

| ver   | change from 1.0.0                          | v1 (file)    | v2 bytes     | v2 AST       |
|-------|--------------------------------------------|--------------|--------------|--------------|
| 1.0.0 | baseline, vulnerable                       | `3gaw…gq5d`  | `tk4r…dt6q`  | `xhuj…blkx`  |
| 1.0.1 | unrelated `version()` added below `greet`  | `5hky…oevp`  | `tk4r…dt6q`  | `xhuj…blkx`  |
| 1.0.2 | `greet` reindented, comment added inside   | `3sma…4ghc`  | `wc2t…wonj`  | `xhuj…blkx`  |
| 1.1.0 | file unchanged, only `package.json` bumped | `3gaw…gq5d`  | `tk4r…dt6q`  | `xhuj…blkx`  |
| 1.2.0 | `greet` rewritten to `execFile`, fixed     | `p364…hdtb`  | `t4jx…fpmh`  | `lt6t…a7ia`  |

Distinct identifiers across the four still-vulnerable releases: v1 produces three, v2-bytes produces two, v2-AST produces one. All three correctly give 1.2.0 a fresh identifier.

What each scheme is sensitive to:

| perturbation                       | v1     | v2 bytes | v2 AST |
|------------------------------------|--------|----------|--------|
| edit elsewhere in file             | new id | same     | same   |
| whitespace/comment inside function | new id | new id   | same   |
| rename or move file                | same   | same     | same   |
| rename a local in the function     | new id | new id   | new id |
| tree-sitter grammar bump           | same   | same*    | new id |
| language tree-sitter can't parse   | works  | falls back to v1 | falls back to v1 |

*v2-bytes only depends on the grammar agreeing where the function starts and ends, which is stable across grammar releases in practice; the hashed bytes themselves are raw source.

Costs. v1 needs nothing but `sha256`. Both v2 variants need a tree-sitter parser per language and a per-language list of function node types. v2-AST additionally needs every party to hold the exact pinned grammar blob, since its hash is in the preimage; bumping that grammar is effectively a spec-version bump and reissues every identifier in that language.

Open questions for either v2: what to do when the sink is at module scope (no enclosing function), nested functions (innermost is currently chosen), and whether the function-node-type list belongs in the spec or is derived from outline's queries.
