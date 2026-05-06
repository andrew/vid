package main

import "testing"

func TestGitoid(t *testing.T) {
	got := gitoid([]byte("hello world\n"))
	want := "0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d"
	if got != want {
		t.Fatalf("gitoid: got %s want %s", got, want)
	}
}

func TestVID1Vector(t *testing.T) {
	got := vid(1, "pkg:generic/example", gitoid([]byte("hello world\n")))
	want := "VID-1-4fta-nqe5-ppht-2m6r-2tpk-o57c"
	if got != want {
		t.Fatalf("vid1: got %s want %s", got, want)
	}
}

func sexpAt(t *testing.T, src, filename string, line int) string {
	t.Helper()
	loc, err := enclosingFunction([]byte(src), filename, line)
	if err != nil {
		t.Fatal(err)
	}
	return canonicalSexp(loc.node, loc.lang, loc.src)
}

func TestSexpWhitespaceInvariant(t *testing.T) {
	a := sexpAt(t, "function f(x) {\n  return bad(x);\n}\n", "a.js", 2)
	b := sexpAt(t, "function f(x){return bad(x);}", "a.js", 1)
	if a != b {
		t.Fatalf("whitespace changed sexp:\n%s\n%s", a, b)
	}
}

func TestSexpCommentInvariant(t *testing.T) {
	a := sexpAt(t, "function f(x) {\n  return bad(x);\n}\n", "a.js", 2)
	b := sexpAt(t, "function f(x) {\n  // careful here\n  return bad(x);\n}\n", "a.js", 3)
	if a != b {
		t.Fatalf("comment changed sexp:\n%s\n%s", a, b)
	}
}

func TestSexpStableAcrossFileEdits(t *testing.T) {
	a := sexpAt(t, "function f() {\n  bad();\n}\n", "x.js", 2)
	b := sexpAt(t, "const k = 1;\n\nfunction f() {\n  bad();\n}\n\nfunction g() {}\n", "x.js", 4)
	if a != b {
		t.Fatalf("unrelated edit changed sexp:\n%s\n%s", a, b)
	}
}

func TestSexpDistinguishesContent(t *testing.T) {
	a := sexpAt(t, "function f() { exec('a'); }", "x.js", 1)
	b := sexpAt(t, "function f() { exec('b'); }", "x.js", 1)
	if a == b {
		t.Fatalf("different string literals produced same sexp: %s", a)
	}
}

func TestGrammarHashDeterministic(t *testing.T) {
	a := grammarHash("javascript")
	b := grammarHash("javascript")
	if a != b || len(a) != 64 {
		t.Fatalf("grammar hash unstable or wrong length: %s / %s", a, b)
	}
	if a == grammarHash("ruby") {
		t.Fatalf("javascript and ruby grammar hashes collide")
	}
}

func TestNoEnclosingFunction(t *testing.T) {
	_, err := enclosingFunction([]byte("const x = 1;\n"), "a.js", 1)
	if err == nil {
		t.Fatal("expected error for top-level statement")
	}
}
