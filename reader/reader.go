package reader

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
)

type Reader struct {
	root                 ast.Node
	cur                  ast.Node
	route                []ast.Node
	currentExpecter      expecter
	lastExpectancyResult expectancyResult
	iteratorsStack       []nodeIterator
	errors               []error
}

type expecter interface {
	process(n ast.Node) expectancyResult
}

type nodeIterator interface {
	next() bool
	node() ast.Node
	empty() bool
}

func NewReader() *Reader {
	return &Reader{
		root:  nil,
		cur:   nil,
		route: nil,
	}
}

func (r *Reader) SetAST(tree ast.Node) {
	r.root = tree
	r.cur = tree
	r.route = r.route[:0]
}

func (r *Reader) ExpectInteger() (int64, error) {
	r.currentExpecter = expectInteger{}
	r.cur.Accept(r)
	if r.hasErrors() {
		return 0, r.error()
	}
}

func (r *Reader) ExpectNullableInteger() (*int64, error) {
	r.currentExpecter = expectNullable{underlying: expectInteger{}}
	r.cur.Accept(r)
	if r.hasErrors() {
		return nil, r.error()
	}
}

func (r *Reader) VisitStreamNode(n *ast.StreamNode) {
	r.enterNode(n)

	result := r.currentExpecter.process(n)
	switch result {
	case expectancyResultMatch, expectancyResultDeny:
		r.lastExpectancyResult = result
		r.leaveNode()
	case expectancyResultContinue:
		r.lastExpectancyResult = result
		streamIter := newStreamIterator(n)
		r.pushIterator(streamIter)
		for r.lastExpectancyResult == expectancyResultContinue && streamIter.next() {
			n := streamIter.node()
			n.Accept(r)
		}
		if streamIter.empty() {
			r.popIterator()
			r.leaveNode()
		}
	default:
		r.appendError(fmt.Errorf("unexpected expectancy result: %s", result))
	}
}

func (r *Reader) VisitTagNode(n *ast.TagNode) {
	r.route = append(r.route, n)
	result := r.currentExpecter.process(n)
	switch result {
	case expectancyResultMatch, expectancyResultDeny, expectancyResultContinue:
		r.lastExpectancyResult = result
	default:
		r.appendError(fmt.Errorf("unexpected expectancy result: %s", result))
	}
	r.leaveNode()
}

func (r *Reader) VisitAnchorNode(n *ast.AnchorNode) {
	r.route = append(r.route, n)
	result := r.currentExpecter.process(n)
	switch result {
	case expectancyResultMatch, expectancyResultDeny, expectancyResultContinue:
		r.lastExpectancyResult = result
	default:
		r.appendError(fmt.Errorf("unexpected expectancy result: %s", result))
	}
	r.leaveNode()
}

func (r *Reader) VisitAliasNode(n *ast.AliasNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitTextNode(n *ast.TextNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitSequenceNode(n *ast.SequenceNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitMappingNode(n *ast.MappingNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitNullNode(n *ast.NullNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitPropertiesNode(n *ast.PropertiesNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitContentNode(n *ast.ContentNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) appendError(err error) {
	r.errors = append(r.errors, err)
}

func (r *Reader) hasErrors() bool {
	return len(r.errors) > 0
}

func (r *Reader) error() error {
	return errors.Join(r.errors...)
}

func (r *Reader) enterNode(n ast.Node) {
	r.route = append(r.route, n)
	r.cur = n
}

func (r *Reader) leaveNode() {
	if len(r.route) > 1 {
		r.route = r.route[:len(r.route)-1]
		r.cur = r.route[len(r.route)-1]
	}
}

func (r *Reader) pushIterator(i nodeIterator) {
	r.iteratorsStack = append(r.iteratorsStack, i)
}

func (r *Reader) popIterator() nodeIterator {
	if len(r.iteratorsStack) == 0 {
		return nil
	}
	i := r.iteratorsStack[len(r.iteratorsStack)-1]
	r.iteratorsStack = r.iteratorsStack[:len(r.iteratorsStack)-1]
	return i
}
