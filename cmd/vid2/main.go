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

func enclosingFunction(src []byte, filename string, line int) ([]byte, string, error) {
	entry := grammars.DetectLanguage(filepath.Base(filename))
	if entry == nil {
		return nil, "", fmt.Errorf("unsupported language for %s", filename)
	}
	lang := entry.Language()
	tree, err := ts.NewParser(lang).Parse(src)
	if err != nil {
		return nil, "", err
	}
	pt := ts.Point{Row: uint32(line - 1), Column: 0}
	node := tree.RootNode().NamedDescendantForPointRange(pt, pt)
	for n := node; n != nil; n = n.Parent() {
		if funcNodes[entry.Name][n.Type(lang)] {
			return src[n.StartByte():n.EndByte()], entry.Name, nil
		}
	}
	return nil, "", fmt.Errorf("no enclosing function at %s:%d", filename, line)
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
	fn, langName, err := enclosingFunction(src, path, line)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	funcOID := gitoid(fn)
	fmt.Println(vid(2, purl, funcOID))
	if os.Getenv("VID_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "lang     %s\n", langName)
		fmt.Fprintf(os.Stderr, "file oid %s\n", gitoid(src))
		fmt.Fprintf(os.Stderr, "func oid %s\n", funcOID)
		fmt.Fprintf(os.Stderr, "func     %q\n", fn)
		fmt.Fprintf(os.Stderr, "VID-1    %s\n", vid(1, purl, gitoid(src)))
	}
}
