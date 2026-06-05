package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	vid "github.com/andrew/VID"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <file>:<line> [<file>:<line> ...]\n", os.Args[0])
}

func parseSink(arg string) (vid.Sink, error) {
	sep := strings.LastIndex(arg, ":")
	if sep < 0 {
		return vid.Sink{}, fmt.Errorf("invalid sink %q: expected file:line", arg)
	}
	file := arg[:sep]
	line, err := strconv.Atoi(arg[sep+1:])
	if err != nil {
		return vid.Sink{}, fmt.Errorf("invalid line in %q: %w", arg, err)
	}
	if line < 1 {
		return vid.Sink{}, fmt.Errorf("invalid line in %q: must be >= 1", arg)
	}
	src, err := os.ReadFile(file)
	if err != nil {
		return vid.Sink{}, err
	}
	return vid.Sink{Filename: file, Source: src, Line: line}, nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	sinks := make([]vid.Sink, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		s, err := parseSink(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sinks = append(sinks, s)
	}
	r := vid.Compute(sinks)
	fmt.Println(r.VID)
	if os.Getenv("VID_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "preimage: %s\n", strings.ReplaceAll(r.Preimage, "\n", `\n`))
		for _, s := range r.Sinks {
			if s.Language != "" {
				fmt.Fprintf(os.Stderr, "sink %s:%d  mode=%s  lang=%s  oid=%s\n", s.Filename, s.Line, s.Mode, s.Language, s.OID)
			} else {
				fmt.Fprintf(os.Stderr, "sink %s:%d  mode=%s  oid=%s\n", s.Filename, s.Line, s.Mode, s.OID)
			}
		}
	}
}
