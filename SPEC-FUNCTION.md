# VID v2 (function-bytes), exploratory

This variant narrows the hashed unit from the sink file to the sink function. Tree-sitter is used to locate the function; what gets hashed is still raw bytes.

## Inputs

`purl` is unchanged from v1.

`funcoid` replaces `gitoid`. It is the SHA-256 gitoid of the byte range of the smallest enclosing function-like AST node at the sink line:

    sha256( "blob " <bytelen> 0x00 <bytes> )

where `<bytes>` is `src[node.start_byte : node.end_byte]` for that node. The byte range is taken verbatim from the source file with no normalisation. Tree-sitter determines where the function starts and ends; it does not contribute to the bytes that are hashed.

Function-like node types per language are listed in `cmd/vid2/funcnodes.go`. The set is derived from the grammars shipped by `github.com/odvcencio/gotreesitter` v0.15.3 and the queries in `github.com/git-pkgs/outline`. A sink line with no enclosing function (top-level code, or an unsupported language) has no v2 VID; fall back to v1.

## Construction

    purl 0x0A funcoid

Hash with SHA-256, take the first 15 bytes, encode as lowercase base32 in groups of four, prefix `VID-2-`.

## Properties relative to v1

A v2 VID is stable across edits anywhere in the file outside the sink function: adding an import, adding an unrelated helper, reformatting another function. It changes when the sink function's bytes change, including whitespace inside it.

Renaming or moving the file does not change the VID (same as v1). Moving the function to a different file in the same package does not change the VID either, since file path is not an input.

The same function bytes appearing in two files of one package produce one VID. The same function copied between packages produces distinct VIDs because the purl differs.

Reproducibility depends on tree-sitter only for locating the byte range. Two grammar versions that agree on where a given function starts and ends produce the same VID even if their internal node structure differs. In practice function boundaries are the most stable thing a grammar reports.

Consumer-side matching now requires parsing every file in the installed tree and computing one funcoid per function, rather than one gitoid per file.

## Test vector

    purl       pkg:npm/acme-greeter
    file       examples/acme-greeter-1.0.0/lib/greet.js
    sink line  4
    func bytes "function greet(name) {\n  exec('echo Hello ' + name);\n}"
    funcoid    0c875fcdffc082ece536ffc1a1d67c81fd7d8dfa216e2b4225b8b55f741d6744

    VID-2-tk4r-cagh-uqzt-4dri-orrz-dt6q

The same VID is produced by `examples/acme-greeter-1.0.1/lib/greet.js`, which adds an unrelated `version()` function below `greet`. Under v1 those two files have different identifiers.
