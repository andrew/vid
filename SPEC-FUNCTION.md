# VID v2 (function AST), exploratory

This variant hashes a canonical serialisation of the sink function's syntax tree rather than file bytes. Whitespace, formatting, and comments inside the function do not affect the identifier. Because the serialised form depends on the grammar that produced it, the grammar is pinned by hash and that hash is part of the preimage.

## Inputs

`purl` is unchanged from v1.

`grammar` is the SHA-256 of the tree-sitter grammar blob used to parse the file, as 64 lowercase hex characters. The reference implementation uses the language blobs shipped by `github.com/odvcencio/gotreesitter` v0.15.3 (`grammars.BlobByName(lang)`). For javascript at that version the value is `6706f93890f24d8ea90d6a140df5dde29c02ec8a3213bae16e8cc4df37e33ee0`. A different grammar version is a different `grammar` input and yields different VIDs by design; two parties only converge if they pin the same grammar.

`astoid` is the SHA-256 gitoid of the canonical S-expression of the smallest enclosing function-like node at the sink line. The serialisation walks named children only, skips `comment` nodes, and emits

    inner node:  (type child child ...)
    leaf node:   (type "text")

where `type` is the grammar's node-type name and `text` is the leaf's source bytes Go-quoted. Anonymous tokens (punctuation, keywords) and whitespace between tokens are not visited and so do not appear.

Function-like node types per language are listed in `cmd/vid2/funcnodes.go`. A sink line with no enclosing function, or in an unsupported language, has no v2 VID; fall back to v1.

## Construction

    purl 0x0A grammar 0x0A astoid

Hash with SHA-256, take the first 15 bytes, encode as lowercase base32 in groups of four, prefix `VID-2-`.

## Properties relative to v1

A v2 VID is stable across edits outside the sink function, across reindentation and reformatting inside it, and across added or removed comments inside it. It changes when any named token in the function changes (an identifier, a literal, an operator's operand structure) or when the pinned grammar changes.

The grammar pin makes reproducibility exact at the cost of coupling every VID to a grammar release. Bumping the pinned grammar is a v2-spec-version bump: every existing v2 VID becomes uncomputable by parties on the new grammar, so the old and new are linked as aliases the same way file-content drift is handled in v1.

Consumer-side matching requires parsing every file in the installed tree with the pinned grammar and computing one astoid per function.

## Test vector

    purl       pkg:npm/acme-greeter
    file       examples/acme-greeter-1.0.0/lib/greet.js
    sink line  4
    grammar    6706f93890f24d8ea90d6a140df5dde29c02ec8a3213bae16e8cc4df37e33ee0
    sexp       (function_expression (identifier "greet") (formal_parameters (identifier "name")) (statement_block (expression_statement (call_expression (identifier "exec") (arguments (binary_expression (string (string_fragment "echo Hello ")) (identifier "name")))))))
    astoid     cd98c3dcfc864620676dc9c3cc125aa45f7d5663d33dd81986aefa2199eb9ddd

    VID-2-xhuj-wflj-lcdn-3xfc-wdlf-blkx

The same VID is produced by 1.0.1 (unrelated helper added to the file), 1.0.2 (function reindented with a comment added), and 1.1.0 (file unchanged). Under v1 those are three distinct identifiers.
