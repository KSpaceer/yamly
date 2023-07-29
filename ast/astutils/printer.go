package astutils

import (
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"io"
	"strings"
)

const (
	edgeLink = "│"
	edgeMid  = "├──"
	edgeEnd  = "└──"
)

const indentSize = 3

// inspired by github.com/xlab/treeprint
type Printer struct {
	edgeType    string
	level       int
	levelsEnded []int
	w           io.Writer
	printedAny  bool
}

func NewPrinter() *Printer {
	return &Printer{
		edgeType:    edgeMid,
		level:       0,
		levelsEnded: nil,
	}
}

func (p *Printer) cloneWithWriter(w io.Writer) *Printer {
	levelsEnded := make([]int, len(p.levelsEnded))
	copy(levelsEnded, p.levelsEnded)
	return &Printer{
		edgeType:    p.edgeType,
		level:       p.level,
		levelsEnded: levelsEnded,
		w:           w,
	}
}

func (p *Printer) Print(root ast.Node, w io.Writer) error {
	if w == nil {
		return fmt.Errorf("nil io.Writer")
	}
	if root == nil {
		return fmt.Errorf("nil ast.Node")
	}
	workingPrinter := p.cloneWithWriter(w)
	workingPrinter.printValue(root)
	root.Accept(workingPrinter)
	return nil
}

func (p *Printer) printValueWithLevel(n ast.Node) {
	for i := 0; i < p.level; i++ {
		if p.isLevelEnded(i) {
			fmt.Fprint(p.w, strings.Repeat(" ", indentSize+1))
			continue
		}
		fmt.Fprintf(p.w, "%s%s", edgeLink, strings.Repeat(" ", indentSize))
	}
	switch casted := n.(type) {
	case ast.Texter:
		fmt.Fprintf(p.w, "%s [%s] %s\n",
			p.edgeType,
			n.Type(),
			strings.ReplaceAll(casted.Text(), "\n", `\n`),
		)
	default:
		fmt.Fprintf(p.w, "%s [%s] \n", p.edgeType, n.Type())
	}
}

func (p *Printer) printRoot(root ast.Node) {
	switch casted := root.(type) {
	case ast.Texter:
		fmt.Fprintf(p.w, "[%s] %s\n", root.Type(), casted.Text())
	default:
		fmt.Fprintf(p.w, "[%s] \n", root.Type())
	}
	p.printedAny = true
}

func (p *Printer) printValue(n ast.Node) {
	if p.printedAny {
		p.printValueWithLevel(n)
	} else {
		p.printRoot(n)
	}
}

func (p *Printer) isLevelEnded(lvl int) bool {
	for _, l := range p.levelsEnded {
		if l == lvl {
			return true
		}
	}
	return false
}

func (p *Printer) VisitStreamNode(n *ast.StreamNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	docs := n.Documents()

	for i, doc := range docs {
		p.edgeType = edgeMid
		if i == len(docs)-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		p.printValue(doc)
		p.level++
		doc.Accept(p)
		p.level--
	}
}

func (p *Printer) VisitTagNode(n *ast.TagNode) {}

func (p *Printer) VisitAnchorNode(n *ast.AnchorNode) {}

func (p *Printer) VisitAliasNode(n *ast.AliasNode) {}

func (p *Printer) VisitTextNode(n *ast.TextNode) {}

func (p *Printer) VisitContentNode(n *ast.ContentNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	properties, content := n.Properties(), n.Content()
	var count, maxCount int
	if properties != nil {
		maxCount++
	}
	if content != nil {
		maxCount++
	}

	if properties != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(properties)
		p.level++
		properties.Accept(p)
		p.level--
	}

	if content != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(content)
		p.level++
		content.Accept(p)
		p.level--
	}
}

func (p *Printer) VisitSequenceNode(n *ast.SequenceNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	entries := n.Entries()
	for i, entry := range entries {
		p.edgeType = edgeMid
		if i == len(entries)-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		p.printValue(entry)
		p.level++
		entry.Accept(p)
		p.level--
	}
}

func (p *Printer) VisitMappingNode(n *ast.MappingNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	entries := n.Entries()
	for i, entry := range entries {
		p.edgeType = edgeMid
		if i == len(entries)-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		p.printValue(entry)
		p.level++
		entry.Accept(p)
		p.level--
	}
}

func (p *Printer) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	key, value := n.Key(), n.Value()
	var count, maxCount int
	if key != nil {
		maxCount++
	}
	if value != nil {
		maxCount++
	}

	if key != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(key)
		p.level++
		key.Accept(p)
		p.level--
	}

	if value != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(value)
		p.level++
		value.Accept(p)
		p.level--
	}
}

func (p *Printer) VisitNullNode(n *ast.NullNode) {}

func (p *Printer) VisitPropertiesNode(n *ast.PropertiesNode) {
	levelsEnded := p.levelsEnded
	defer func() {
		p.levelsEnded = levelsEnded
	}()

	tag, anchor := n.Tag(), n.Anchor()
	var count, maxCount int
	if tag != nil {
		maxCount++
	}
	if anchor != nil {
		maxCount++
	}

	if tag != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(tag)
		p.level++
		tag.Accept(p)
		p.level--
	}

	if anchor != nil {
		p.edgeType = edgeMid
		if count == maxCount-1 {
			p.levelsEnded = append(p.levelsEnded, p.level)
			p.edgeType = edgeEnd
		}
		count++
		p.printValue(anchor)
		p.level++
		anchor.Accept(p)
		p.level--
	}
}
