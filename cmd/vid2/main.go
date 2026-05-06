package main

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

var enc = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

func gitoid(b []byte) string {
	h := sha256.New()
	fmt.Fprintf(h, "blob %d\x00", len(b))
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func vid(version int, parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	s := enc.EncodeToString(sum[:15])
	var b strings.Builder
	fmt.Fprintf(&b, "VID-%d", version)
	for i := 0; i < len(s); i += 4 {
		b.WriteByte('-')
		b.WriteString(s[i : i+4])
	}
	return b.String()
}

type located struct {
	node     *ts.Node
	lang     *ts.Language
	langName string
	src      []byte
}

func enclosingFunction(src []byte, filename string, line int) (*located, error) {
	entry := grammars.DetectLanguage(filepath.Base(filename))
	if entry == nil {
		return nil, fmt.Errorf("unsupported language for %s", filename)
	}
	lang := entry.Language()
	tree, err := ts.NewParser(lang).Parse(src)
	if err != nil {
		return nil, err
	}
	pt := ts.Point{Row: uint32(line - 1), Column: 0}
	node := tree.RootNode().NamedDescendantForPointRange(pt, pt)
	for n := node; n != nil; n = n.Parent() {
		if funcNodes[entry.Name][n.Type(lang)] {
			return &located{node: n, lang: lang, langName: entry.Name, src: src}, nil
		}
	}
	return nil, fmt.Errorf("no enclosing function at %s:%d", filename, line)
}

func grammarHash(langName string) string {
	blob := grammars.BlobByName(langName)
	sum := sha256.Sum256(blob)
	return fmt.Sprintf("%x", sum[:])
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: %s <purl> <file> <line>\n", os.Args[0])
		os.Exit(1)
	}
	purl, path, lineArg := os.Args[1], os.Args[2], os.Args[3]
	line, err := strconv.Atoi(lineArg)
	if err != nil || line < 1 {
		fmt.Fprintln(os.Stderr, "line must be a positive integer")
		os.Exit(1)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	loc, err := enclosingFunction(src, path, line)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sexp := canonicalSexp(loc.node, loc.lang, loc.src)
	gh := grammarHash(loc.langName)
	astOID := gitoid([]byte(sexp))
	fmt.Println(vid(2, purl, gh, astOID))
	if os.Getenv("VID_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "lang     %s\n", loc.langName)
		fmt.Fprintf(os.Stderr, "grammar  %s\n", gh)
		fmt.Fprintf(os.Stderr, "ast oid  %s\n", astOID)
		fmt.Fprintf(os.Stderr, "sexp     %s\n", sexp)
		fmt.Fprintf(os.Stderr, "VID-1    %s\n", vid(1, purl, gitoid(src)))
	}
}
