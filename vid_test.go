package vid_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	vid "github.com/andrew/VID"
)

func TestGitoidHelloWorld(t *testing.T) {
	got := vid.Gitoid([]byte("hello world\n"))
	want := "0bd69098bd9b9cc5934a610ab65da429b525361147faa7b5b922919e9a23143d"
	if got != want {
		t.Fatalf("gitoid: got %s want %s", got, want)
	}
}

func TestEncodeFormat(t *testing.T) {
	s := vid.Encode("anything")
	if !strings.HasPrefix(s, "VID-") {
		t.Fatalf("missing VID- prefix: %s", s)
	}
	// "VID" + 6 groups of 4 chars = 7 segments
	parts := strings.Split(s, "-")
	if len(parts) != 7 {
		t.Fatalf("expected 7 hyphen-separated segments, got %d in %s", len(parts), s)
	}
	for _, g := range parts[1:] {
		if len(g) != 4 {
			t.Fatalf("expected 4-char groups, got %q in %s", g, s)
		}
	}
}

func TestEncodeDeterministic(t *testing.T) {
	a := vid.Encode("preimage-fixed")
	b := vid.Encode("preimage-fixed")
	if a != b {
		t.Fatalf("encode not deterministic: %s vs %s", a, b)
	}
}

func TestEncodeGitoidFileFallbackVector(t *testing.T) {
	// File-fallback preimage is just the gitoid string.
	gitoid := vid.Gitoid([]byte("hello world\n"))
	got := vid.Encode(gitoid)
	want := "VID-ernd-c2qt-gs7l-lkhg-32lv-y4li"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestComputeFunctionSink(t *testing.T) {
	src := []byte("const x = 1;\n\nfunction greet(name) {\n  exec('echo ' + name);\n}\n")
	r := vid.Compute([]vid.Sink{{Filename: "a.js", Source: src, Line: 4}})
	if len(r.Sinks) != 1 {
		t.Fatalf("expected 1 sink, got %d", len(r.Sinks))
	}
	if r.Sinks[0].Mode != vid.ModeFunction {
		t.Fatalf("expected function mode, got %s", r.Sinks[0].Mode)
	}
	if r.Sinks[0].Language != "javascript" {
		t.Fatalf("expected javascript, got %s", r.Sinks[0].Language)
	}
	if r.Preimage != r.Sinks[0].OID {
		t.Fatalf("single-sink preimage should equal OID; got preimage=%q oid=%q", r.Preimage, r.Sinks[0].OID)
	}
}

func TestComputeFileFallback(t *testing.T) {
	// Top-level statement, no enclosing function.
	src := []byte("const x = 1;\nconst y = 2;\n")
	r := vid.Compute([]vid.Sink{{Filename: "a.js", Source: src, Line: 1}})
	if r.Sinks[0].Mode != vid.ModeFile {
		t.Fatalf("expected file mode, got %s", r.Sinks[0].Mode)
	}
	if r.Sinks[0].OID != vid.Gitoid(src) {
		t.Fatalf("file-fallback OID should be file gitoid")
	}
}

func TestComputeStableAcrossUnrelatedFileEdits(t *testing.T) {
	a := []byte("function f() {\n  bad();\n}\n")
	b := []byte("// added comment\nconst k = 1;\n\nfunction f() {\n  bad();\n}\n\nfunction g() {}\n")
	ra := vid.Compute([]vid.Sink{{Filename: "x.js", Source: a, Line: 2}})
	rb := vid.Compute([]vid.Sink{{Filename: "x.js", Source: b, Line: 5}})
	if ra.VID != rb.VID {
		t.Fatalf("VID changed across unrelated file edits: %s vs %s", ra.VID, rb.VID)
	}
}

func TestComputeMultiSinkSortedAndDeduped(t *testing.T) {
	srcA := []byte("function a() {\n  badA();\n}\n")
	srcB := []byte("function b() {\n  badB();\n}\n")

	r1 := vid.Compute([]vid.Sink{
		{Filename: "a.js", Source: srcA, Line: 2},
		{Filename: "b.js", Source: srcB, Line: 2},
	})
	r2 := vid.Compute([]vid.Sink{
		{Filename: "b.js", Source: srcB, Line: 2},
		{Filename: "a.js", Source: srcA, Line: 2},
	})
	if r1.VID != r2.VID {
		t.Fatalf("multi-sink VID depends on input order: %s vs %s", r1.VID, r2.VID)
	}

	// Two pointers to the same function should dedup to one OID in the preimage.
	rDup := vid.Compute([]vid.Sink{
		{Filename: "a.js", Source: srcA, Line: 2},
		{Filename: "a.js", Source: srcA, Line: 2},
	})
	rOnce := vid.Compute([]vid.Sink{
		{Filename: "a.js", Source: srcA, Line: 2},
	})
	if rDup.VID != rOnce.VID {
		t.Fatalf("duplicate sinks not deduped: %s vs %s", rDup.VID, rOnce.VID)
	}
}

func TestComputeMultiSinkDifferentFromSingleSink(t *testing.T) {
	srcA := []byte("function a() {\n  badA();\n}\n")
	srcB := []byte("function b() {\n  badB();\n}\n")
	single := vid.Compute([]vid.Sink{{Filename: "a.js", Source: srcA, Line: 2}})
	multi := vid.Compute([]vid.Sink{
		{Filename: "a.js", Source: srcA, Line: 2},
		{Filename: "b.js", Source: srcB, Line: 2},
	})
	if single.VID == multi.VID {
		t.Fatalf("multi-sink VID equals single-sink VID; preimages should differ")
	}
}

type fileLoc struct {
	File string `yaml:"file"`
	Line int    `yaml:"line"`
}

type perturbation struct {
	File string `yaml:"file"`
	Line int    `yaml:"line"`
	Note string `yaml:"note"`
}

type expected struct {
	VID  string `yaml:"vid"`
	Mode string `yaml:"mode"`
}

type manifest struct {
	Advisory          string         `yaml:"advisory"`
	CVE               string         `yaml:"cve"`
	Purl              string         `yaml:"purl"`
	Sink              *fileLoc       `yaml:"sink"`
	Sinks             []fileLoc      `yaml:"sinks"`
	Fixed             *fileLoc       `yaml:"fixed"`
	FixedSinks        []fileLoc      `yaml:"fixed_sinks"`
	Perturbation      *perturbation  `yaml:"perturbation"`
	PerturbationSinks []perturbation `yaml:"perturbation_sinks"`
	Expected          expected       `yaml:"expected"`
	Refs              []string       `yaml:"refs"`
}

func (m manifest) sinkList() []fileLoc {
	if len(m.Sinks) > 0 {
		return m.Sinks
	}
	if m.Sink != nil {
		return []fileLoc{*m.Sink}
	}
	return nil
}

func (m manifest) fixedList() []fileLoc {
	if len(m.FixedSinks) > 0 {
		return m.FixedSinks
	}
	if m.Fixed != nil {
		return []fileLoc{*m.Fixed}
	}
	return nil
}

func (m manifest) perturbationList() []perturbation {
	if len(m.PerturbationSinks) > 0 {
		return m.PerturbationSinks
	}
	if m.Perturbation != nil {
		return []perturbation{*m.Perturbation}
	}
	return nil
}

func loadManifest(t *testing.T, path string) manifest {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m
}

func loadSinks(t *testing.T, dir string, locs []fileLoc) []vid.Sink {
	t.Helper()
	sinks := make([]vid.Sink, len(locs))
	for i, loc := range locs {
		full := filepath.Join(dir, loc.File)
		src, err := os.ReadFile(full)
		if err != nil {
			t.Fatalf("read %s: %v", full, err)
		}
		sinks[i] = vid.Sink{Filename: full, Source: src, Line: loc.Line}
	}
	return sinks
}

func TestAdvisoryFixtures(t *testing.T) {
	matches, err := filepath.Glob("examples/*/advisory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no advisory fixtures present yet")
	}
	for _, manifestPath := range matches {
		name := filepath.Base(filepath.Dir(manifestPath))
		t.Run(name, func(t *testing.T) {
			m := loadManifest(t, manifestPath)
			dir := filepath.Dir(manifestPath)

			sinkLocs := m.sinkList()
			if len(sinkLocs) == 0 {
				t.Fatalf("manifest %s has no sink or sinks", manifestPath)
			}
			sinkResult := vid.Compute(loadSinks(t, dir, sinkLocs))
			if sinkResult.VID != m.Expected.VID {
				t.Errorf("sink VID: got %s want %s", sinkResult.VID, m.Expected.VID)
			}
			if string(sinkResult.Sinks[0].Mode) != m.Expected.Mode {
				t.Errorf("sink mode: got %s want %s", sinkResult.Sinks[0].Mode, m.Expected.Mode)
			}

			if fixedLocs := m.fixedList(); len(fixedLocs) > 0 {
				fixed := vid.Compute(loadSinks(t, dir, fixedLocs))
				if fixed.VID == sinkResult.VID {
					t.Errorf("fixed VID equals vulnerable VID (%s); a real fix should change the hashed bytes", sinkResult.VID)
				}
			}

			if pertLocs := m.perturbationList(); len(pertLocs) > 0 {
				if len(pertLocs) != len(sinkLocs) {
					t.Fatalf("perturbation list length %d does not match sink list length %d", len(pertLocs), len(sinkLocs))
				}
				locs := make([]fileLoc, len(pertLocs))
				for i, p := range pertLocs {
					locs[i] = fileLoc{File: p.File, Line: p.Line}
					if locs[i].Line == 0 {
						locs[i].Line = sinkLocs[i].Line
					}
				}
				perturbed := vid.Compute(loadSinks(t, dir, locs))
				if perturbed.VID != sinkResult.VID {
					var notes []string
					for _, p := range pertLocs {
						notes = append(notes, p.Note)
					}
					t.Errorf("perturbation %v changed the VID (got %s, want same as sink %s)", notes, perturbed.VID, sinkResult.VID)
				}
			}
		})
	}
}
