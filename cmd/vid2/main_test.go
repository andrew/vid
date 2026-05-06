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

func TestEnclosingFunctionJS(t *testing.T) {
	src := []byte("const x = 1;\n\nfunction greet(name) {\n  exec('echo ' + name);\n}\n\nconst y = 2;\n")
	fn, lang, err := enclosingFunction(src, "a.js", 4)
	if err != nil {
		t.Fatal(err)
	}
	if lang != "javascript" {
		t.Fatalf("lang: got %s", lang)
	}
	want := "function greet(name) {\n  exec('echo ' + name);\n}"
	if string(fn) != want {
		t.Fatalf("func: got %q want %q", fn, want)
	}
}

func TestEnclosingFunctionStableAcrossFileEdits(t *testing.T) {
	a := []byte("function f() {\n  bad();\n}\n")
	b := []byte("// added comment\nconst k = 1;\n\nfunction f() {\n  bad();\n}\n\nfunction g() {}\n")
	fa, _, err := enclosingFunction(a, "x.js", 2)
	if err != nil {
		t.Fatal(err)
	}
	fb, _, err := enclosingFunction(b, "x.js", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gitoid(fa) != gitoid(fb) {
		t.Fatalf("function gitoid changed across unrelated file edits:\n%q\n%q", fa, fb)
	}
}

func TestNoEnclosingFunction(t *testing.T) {
	src := []byte("const x = 1;\n")
	_, _, err := enclosingFunction(src, "a.js", 1)
	if err == nil {
		t.Fatal("expected error for top-level statement")
	}
}
