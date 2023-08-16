package reader

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
)

type Reader struct {
	route []routePoint

	currentExpecter      expecter
	lastExpectancyResult expectancyResult

	extractedValue string

	errors []error
}

type routePoint struct {
	initialized bool
	node        ast.Node
	iter        nodeIterator
}

type expecter interface {
	name() string
	process(n ast.Node) expectancyResult
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
		node:        tree,
		initialized: false,
	})
}

func (r *Reader) ExpectInteger() (int64, error) {
	r.currentExpecter = expectInteger{}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return 0, r.error()
	}
}

func (r *Reader) ExpectNullableInteger() (*int64, error) {
	r.currentExpecter = expectNullable{underlying: expectInteger{}}
	r.currentNode().Accept(r)
	if r.hasErrors() {
		return nil, r.error()
	}
}

func (r *Reader) VisitStreamNode(n *ast.StreamNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.iter = newStreamIterator(n)
		point.initialized = true
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch:
		r.swapRoutePoint(point)
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case expectancyResultContinue:
		for r.lastExpectancyResult == expectancyResultContinue && point.iter.next() {
			doc := point.iter.node()
			r.pushRoutePoint(routePoint{
				node:        doc,
				initialized: false,
			})
			doc.Accept(r)
		}
		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitTagNode(n *ast.TagNode) {

}

func (r *Reader) VisitAnchorNode(n *ast.AnchorNode) {

}

func (r *Reader) VisitAliasNode(n *ast.AliasNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitTextNode(n *ast.TextNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.initialized = true
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch:
		r.extractedValue = n.Text()
		r.popRoutePoint()
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case expectancyResultContinue:
		r.popRoutePoint()
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitSequenceNode(n *ast.SequenceNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.initialized = true
		point.iter = newSequenceIterator(n)
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch:
		r.swapRoutePoint(point)
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case expectancyResultContinue:
		for r.lastExpectancyResult == expectancyResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node:        entry,
				initialized: false,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitMappingNode(n *ast.MappingNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.initialized = true
		point.iter = newMappingIterator(n)
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch:
		r.swapRoutePoint(point)
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case expectancyResultContinue:
		for r.lastExpectancyResult == expectancyResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node:        entry,
				initialized: false,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.initialized = true
		point.iter = newMappingEntryIterator(n)
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch:
		r.swapRoutePoint(point)
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	case expectancyResultContinue:
		for r.lastExpectancyResult == expectancyResultContinue && point.iter.next() {
			entry := point.iter.node()
			r.pushRoutePoint(routePoint{
				node:        entry,
				initialized: false,
			})
			entry.Accept(r)
		}

		if point.iter.empty() {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitNullNode(n *ast.NullNode) {
	point := r.peekRoutePoint()
	if !point.initialized {
		point.initialized = true
	}

	result := r.currentExpecter.process(n)
	r.lastExpectancyResult = result
	switch result {
	case expectancyResultMatch, expectancyResultContinue:
		r.popRoutePoint()
	case expectancyResultDeny:
		r.swapRoutePoint(point)
		r.appendError(&DenyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		})
	default:
		r.appendError(fmt.Errorf("unexpected result: %s", result))
	}
}

func (r *Reader) VisitPropertiesNode(n *ast.PropertiesNode) {
	//TODO implement me
	panic("implement me")
}

func (r *Reader) VisitContentNode(n *ast.ContentNode) {
	//TODO implement me
	panic("implement me")
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
