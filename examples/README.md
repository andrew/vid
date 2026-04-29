# Worked example

Three versions of a trivial npm package, `acme-greeter`, and a reference script that computes a VID per [SPEC.md](../SPEC.md).

## The package

`lib/greet.js` shells out to `echo` with a user-supplied name. In 1.0.0 and 1.1.0 the name is concatenated into the command string, which is command injection (CWE-78). In 1.2.0 the call is changed to `execFile` with an argument array, which fixes it.

The only difference between 1.0.0 and 1.1.0 is the `version` field in `package.json`. The vulnerable file is byte-identical across both.

## Computing VIDs

`vid.rb` takes a purl and a file path and prints the VID. It does no purl normalisation; pass the canonical form.

    $ ruby vid.rb pkg:npm/acme-greeter acme-greeter-1.0.0/lib/greet.js
    VID-1-3gaw-jglb-6jhv-la2v-t2u5-gq5d

    $ ruby vid.rb pkg:npm/acme-greeter acme-greeter-1.1.0/lib/greet.js
    VID-1-3gaw-jglb-6jhv-la2v-t2u5-gq5d

    $ ruby vid.rb pkg:npm/acme-greeter acme-greeter-1.2.0/lib/greet.js
    VID-1-p364-l3rk-s62v-ckbc-n5er-hdtb

## What this shows

1.0.0 and 1.1.0 share a VID because the vulnerable file has the same bytes in both. The purl carries no version, so a finding reported against 1.0.0 matches a user running 1.1.0 with no extra work.

1.2.0 has a different VID because the file changed. The original VID no longer matches anything in this release. A surrounding record would mark the original VID as fixed in 1.2.0.

If a hypothetical 1.3.0 reverted `lib/greet.js` to the 1.0.0 content, `VID-1-3gaw-jglb-6jhv-la2v-t2u5-gq5d` would match it again. Copy `acme-greeter-1.0.0/lib/greet.js` over `acme-greeter-1.2.0/lib/greet.js` and re-run to see this.

## Intermediate values

For implementers checking their own code:

    purl            pkg:npm/acme-greeter
    file            lib/greet.js (1.0.0 and 1.1.0)
    gitoid          c11ea89b91d920150cf5e904ba8889355d92fb65ce2a5e7383a2054414c37e4f
    VID             VID-1-3gaw-jglb-6jhv-la2v-t2u5-gq5d

    file            lib/greet.js (1.2.0)
    gitoid          f736b73c84abde7a66006ffcb49cb59a11bd6cbdeac01738399e5a438fe75c3e
    VID             VID-1-p364-l3rk-s62v-ckbc-n5er-hdtb

The script also reproduces both test vectors in SPEC.md when run against a file containing `hello world\n`.
