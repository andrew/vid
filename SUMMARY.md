# VID in brief

Finding vulnerabilities in open source used to be slow work, and the disclosure process was sized for that pace. Automated scanners now surface real bugs in minutes, and many people run them against the same popular libraries at once. Maintainers get the same report five times over and stop reading; users see nothing until a CVE appears months later.

A VID is computed, not issued. Hash the package's purl together with the gitoid of the file where the bug lives and out comes a short string like `VID-1-4fta-nqe5-ppht-2m6r-2tpk-o57c`. Anyone looking at the same bug gets the same string without asking an authority. CWE, severity, and line number travel as metadata rather than inputs to the hash.

That string is a dedup key in a maintainer's inbox, a timestamp a researcher can publish before disclosure without revealing the finding, and something a consumer can match against their installed dependency tree before any CVE exists. It does not replace CVE, requires no database, and can be computed for a false positive as easily as a real bug.

v1 hashes the whole sink file, so unrelated edits produce a new identifier for the same unfixed bug. Two branches narrow the unit to the enclosing function via tree-sitter: `vid2-function-bytes` hashes the function's source bytes, `vid2-function-ast` hashes its syntax tree with the grammar pinned. On the worked example four still-vulnerable releases give three identifiers under v1, two under bytes, one under AST; see [COMPARISON.md](COMPARISON.md).
