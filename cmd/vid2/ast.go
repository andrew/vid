package main

import (
	"strconv"
	"strings"

	ts "github.com/odvcencio/gotreesitter"
)

func canonicalSexp(n *ts.Node, lang *ts.Language, src []byte) string {
	var b strings.Builder
	writeSexp(&b, n, lang, src)
	return b.String()
}

func writeSexp(b *strings.Builder, n *ts.Node, lang *ts.Language, src []byte) {
	typ := n.Type(lang)
	if typ == "comment" {
		return
	}
	nc := n.NamedChildCount()
	if nc == 0 {
		b.WriteByte('(')
		b.WriteString(typ)
		b.WriteByte(' ')
		b.WriteString(strconv.Quote(string(src[n.StartByte():n.EndByte()])))
		b.WriteByte(')')
		return
	}
	b.WriteByte('(')
	b.WriteString(typ)
	for i := range nc {
		c := n.NamedChild(i)
		if c.Type(lang) == "comment" {
			continue
		}
		b.WriteByte(' ')
		writeSexp(b, c, lang, src)
	}
	b.WriteByte(')')
}
