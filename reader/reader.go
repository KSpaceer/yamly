package reader

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/schema"
)

type Reader struct {
	route []routePoint

	currentExpecter    expecter
	lastVisitingResult visitingResult

	extractedValue  string
	isExtractedNull bool

	anchors    map[string]ast.Node
	metAnchor  bool
	anchorName string

	errors []error
}

type visitingResult int8

const (
	visitingResultUnknown visitingResult = iota
	visitingResultMatch
	visitingResultDeny
	visitingResultContinue
)

func (v visitingResult) String() string {
	switch v {
	case visitingResultUnknown:
		return "unknown"
	case visitingResultMatch:
		return "match"
	case visitingResultDeny:
		return "deny"
	case visitingResultContinue:
		return "continue"
	default:
		return fmt.Sprintf("unsupported value (%d)", v)
	}
}

type routePoint struct {
	visitingResult visitingResult
	node           ast.Node
	iter           nodeIterator
}

type expecter interface {
	name() string
	process(n ast.Node, previousResult visitingResult) visitingResult
}

type nodeIterator interface {
	next() bool
	node() ast.Node
	empty() bool
}

func NewReader() *Reader {
	return &Reader{
		route: nil,
	}
}

func (r *Reader) SetAST(tree ast.Node) {
	r.route = r.route[:0]
	r.pushRoutePoint(routePoint{
		node: tree,
	})
}

func (r *Reader) ExpectInteger() (int64, error) {
	r.currentExpecter = expectInteger{}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return 0, r.error()
	}
	v, err := schema.ToInteger(r.extractedValue)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (r *Reader) ExpectNullableInteger() (*int64, error) {
	r.currentExpecter = expectNullable{underlying: expectInteger{}}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return nil, r.error()
	}
	if r.isExtractedNull {
		r.isExtractedNull = false
		return nil, nil
	}
	v, err := schema.ToInteger(r.extractedValue)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Reader) ExpectBoolean() (bool, error) {
	r.currentExpecter = expectBoolean{}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return false, r.error()
	}
	v, err := schema.ToBoolean(r.extractedValue)
	if err != nil {
		return false, err
	}
	return v, nil
}

func (r *Reader) ExpectNullableBoolean() (*bool, error) {
	r.currentExpecter = expectNullable{underlying: expectBoolean{}}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return nil, r.error()
	}
	if r.isExtractedNull {
		r.isExtractedNull = false
		return nil, nil
	}
	v, err := schema.ToBoolean(r.extractedValue)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Reader) VisitStreamNode(n *ast.StreamNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newStreamIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			doc := point.iter.node()
			r.pushRoutePoint(routePoint{
				node: doc,
			})
			doc.Accept(r)
		}
		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitTagNode(n *ast.TagNode) {
	r.visitTexterNode(n)
}

func (r *Reader) VisitAnchorNode(n *ast.AnchorNode) {
	r.metAnchor = true
	r.anchorName = n.Text()

	r.visitTexterNode(n)
}

func (r *Reader) VisitAliasNode(n *ast.AliasNode) {
	alias := n.Text()
	anchored, ok := r.anchors[alias]
	if !ok {
		r.appendError(AliasDereferenceError{name: alias})
	}
	r.pushRoutePoint(routePoint{
		node: anchored,
	})
	anchored.Accept(r)
}

func (r *Reader) VisitTextNode(n *ast.TextNode) {
	r.visitTexterNode(n)
}

func (r *Reader) VisitSequenceNode(n *ast.SequenceNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newSequenceIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node: entry,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitMappingNode(n *ast.MappingNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newMappingIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node: entry,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newMappingEntryIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node: entry,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitNullNode(n *ast.NullNode) {
	point := r.peekRoutePoint()

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.extractedValue = ""
		r.isExtractedNull = true
		r.popRoutePoint()
	case visitingResultContinue:
		r.popRoutePoint()
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitPropertiesNode(n *ast.PropertiesNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newPropertiesIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			property := point.iter.node()
			r.pushRoutePoint(routePoint{
				node: property,
			})
			property.Accept(r)
		}
		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) VisitContentNode(n *ast.ContentNode) {
	point := r.peekRoutePoint()
	if point.visitingResult == visitingResultUnknown {
		point.iter = newContentIterator(n)
	}

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.swapRoutePoint(point)
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		for r.lastVisitingResult == visitingResultContinue && point.iter.next() {
			node := point.iter.node()
			if r.metAnchor {
				r.anchors[r.anchorName] = node
				r.metAnchor = false
			}
			r.pushRoutePoint(routePoint{
				node: node,
			})
			node.Accept(r)
		}
		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) visitTexterNode(n ast.TexterNode) {
	point := r.peekRoutePoint()

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult {
	case visitingResultMatch:
		r.extractedValue = n.Text()
		r.isExtractedNull = false
		r.popRoutePoint()
	case visitingResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case visitingResultContinue:
		r.popRoutePoint()
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", point.visitingResult))
	}
}

func (r *Reader) currentNode() ast.Node {
	if len(r.route) == 0 {
		return nil
	}
	return r.route[len(r.route)-1].node
}

func (r *Reader) pushRoutePoint(point routePoint) {
	r.route = append(r.route, point)
}

func (r *Reader) popRoutePoint() routePoint {
	if len(r.route) == 0 {
		return routePoint{}
	}
	point := r.route[len(r.route)-1]
	r.route = r.route[:len(r.route)-1]
	return point
}

func (r *Reader) peekRoutePoint() routePoint {
	if len(r.route) == 0 {
		return routePoint{}
	}
	point := r.route[len(r.route)-1]
	return point
}

func (r *Reader) swapRoutePoint(point routePoint) {
	if len(r.route) == 0 {
		return
	}
	r.route[len(r.route)-1] = point
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
