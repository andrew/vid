package vid

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

var enc = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

type Mode string

const (
	ModeFunction Mode = "function"
	ModeFile     Mode = "file"
)

type Sink struct {
	Filename string
	Source   []byte
	Line     int
}

type SinkResult struct {
	Filename string
	Line     int
	OID      string
	Mode     Mode
	Language string
}

type Result struct {
	VID      string
	Preimage string
	Sinks    []SinkResult
}

type langEntry struct {
	name     string
	language func() *tree_sitter.Language
}

var languageByExt = map[string]langEntry{
	".js":  {"javascript", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_javascript.Language()) }},
	".jsx": {"javascript", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_javascript.Language()) }},
	".mjs": {"javascript", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_javascript.Language()) }},
	".cjs": {"javascript", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_javascript.Language()) }},
	".ts": {"typescript", func() *tree_sitter.Language {
		return tree_sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript())
	}},
	".tsx":  {"typescript", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_typescript.LanguageTSX()) }},
	".rb":   {"ruby", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_ruby.Language()) }},
	".py":   {"python", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_python.Language()) }},
	".go":   {"go", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_go.Language()) }},
	".rs":   {"rust", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_rust.Language()) }},
	".java": {"java", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_java.Language()) }},
	".c":    {"c", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_c.Language()) }},
	".h":    {"c", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_c.Language()) }},
	".cpp":  {"cpp", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_cpp.Language()) }},
	".cc":   {"cpp", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_cpp.Language()) }},
	".cxx":  {"cpp", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_cpp.Language()) }},
	".hpp":  {"cpp", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_cpp.Language()) }},
	".php":  {"php", func() *tree_sitter.Language { return tree_sitter.NewLanguage(tree_sitter_php.LanguagePHP()) }},
}

func detectLanguage(filename string) (langEntry, bool) {
	ext := strings.ToLower(filepath.Ext(filepath.Base(filename)))
	if strings.HasSuffix(filename, ".d.ts") {
		ext = ".ts"
	}
	entry, ok := languageByExt[ext]
	return entry, ok
}

func Gitoid(b []byte) string {
	h := sha256.New()
	// hash.Hash.Write is documented never to return an error.
	_, _ = fmt.Fprintf(h, "blob %d\x00", len(b))
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Encode(preimage string) string {
	sum := sha256.Sum256([]byte(preimage))
	s := enc.EncodeToString(sum[:15])
	var b strings.Builder
	b.WriteString("VID")
	for i := 0; i < len(s); i += 4 {
		b.WriteByte('-')
		b.WriteString(s[i : i+4])
	}
	return b.String()
}

func EnclosingFunction(src []byte, filename string, line int) (fn []byte, lang string, ok bool) {
	entry, found := detectLanguage(filename)
	if !found {
		return nil, "", false
	}
	language := entry.language()
	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(language); err != nil {
		return nil, "", false
	}
	tree := parser.Parse(src, nil)
	if tree == nil {
		return nil, "", false
	}
	defer tree.Close()
	col := firstNonWhitespaceColumn(src, line)
	pt := tree_sitter.Point{Row: uint(line - 1), Column: uint(col)}
	node := tree.RootNode().NamedDescendantForPointRange(pt, pt)
	if node == nil {
		return nil, "", false
	}
	for n := node; n != nil; n = n.Parent() {
		if funcNodes[entry.name][n.Kind()] {
			start, end := n.ByteRange()
			return src[start:end], entry.name, true
		}
	}
	return nil, "", false
}

func firstNonWhitespaceColumn(src []byte, line int) int {
	lineStart := 0
	current := 1
	for i, b := range src {
		if current == line {
			lineStart = i
			break
		}
		if b == '\n' {
			current++
		}
	}
	col := 0
	for lineStart+col < len(src) {
		b := src[lineStart+col]
		if b == ' ' || b == '\t' {
			col++
			continue
		}
		break
	}
	return col
}

func resolveSink(s Sink) SinkResult {
	if fn, lang, ok := EnclosingFunction(s.Source, s.Filename, s.Line); ok {
		return SinkResult{
			Filename: s.Filename,
			Line:     s.Line,
			OID:      Gitoid(fn),
			Mode:     ModeFunction,
			Language: lang,
		}
	}
	return SinkResult{
		Filename: s.Filename,
		Line:     s.Line,
		OID:      Gitoid(s.Source),
		Mode:     ModeFile,
	}
}

func Compute(sinks []Sink) Result {
	resolved := make([]SinkResult, len(sinks))
	oids := make([]string, len(sinks))
	for i, s := range sinks {
		resolved[i] = resolveSink(s)
		oids[i] = resolved[i].OID
	}
	sort.Strings(oids)
	unique := oids[:0]
	var last string
	for _, oid := range oids {
		if oid != last {
			unique = append(unique, oid)
			last = oid
		}
	}
	preimage := strings.Join(unique, "\n")
	return Result{
		VID:      Encode(preimage),
		Preimage: preimage,
		Sinks:    resolved,
	}
}
