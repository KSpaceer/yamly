package decode

import (
	"errors"
	"fmt"
	"time"

	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/encode"
	"github.com/KSpaceer/yamly/engines/yayamls/parser"
	"github.com/KSpaceer/yamly/engines/yayamls/schema"
)

var _ yamly.ExtendedDecoder[ast.Node] = (*ASTReader)(nil)

type ASTReader struct {
	route []routePoint

	currentExpecter    expecter
	lastVisitingResult visitingResult

	extractedCollectionState yamly.CollectionState
	extractedValue           string

	anchors anchorsKeeper

	multipleDenyErrors bool
	fatalError         error
	latestDenyError    error
	denyErrors         []error
}

type visitingConclusion int8

const (
	visitingConclusionUnknown visitingConclusion = iota
	visitingConclusionMatch
	visitingConclusionConsume
	visitingConclusionDeny
	visitingConclusionContinue
)

func (v visitingConclusion) String() string {
	switch v {
	case visitingConclusionUnknown:
		return "unknown"
	case visitingConclusionMatch:
		return "match"
	case visitingConclusionConsume:
		return "consume"
	case visitingConclusionDeny:
		return "deny"
	case visitingConclusionContinue:
		return "continue"
	default:
		return fmt.Sprintf("unsupported value (%d)", v)
	}
}

type visitingAction int8

const (
	visitingActionNothing visitingAction = iota
	visitingActionExtract
)

type visitingResult struct {
	conclusion visitingConclusion
	action     visitingAction
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
	node() ast.Node
	empty() bool
}

type collectionState struct {
	size int
	nodeIterator
}

func (s *collectionState) Size() int { return s.size }

func (s *collectionState) HasUnprocessedItems() bool { return !s.empty() }

var noopCollectionState yamly.CollectionState = noopState{}

type noopState struct{}

func (noopState) Size() int { return 0 }

func (noopState) HasUnprocessedItems() bool { return false }

func newCollectionState(iter nodeIterator, size int) yamly.CollectionState {
	return &collectionState{
		size:         size,
		nodeIterator: iter,
	}
}

type ReaderOption func(*ASTReader)

func WithMultipleDenyErrors() ReaderOption {
	return func(r *ASTReader) {
		r.multipleDenyErrors = true
	}
}

func NewASTReaderFromBytes(src []byte, opts ...ReaderOption) (*ASTReader, error) {
	tree, err := parser.ParseBytes(src, parser.WithOmitStream())
	if err != nil {
		return nil, err
	}
	return NewASTReader(tree, opts...), nil
}

func NewASTReader(tree ast.Node, opts ...ReaderOption) *ASTReader {
	r := ASTReader{anchors: newAnchorsKeeper()}

	for _, opt := range opts {
		opt(&r)
	}

	r.setAST(tree)
	return &r
}

func (r *ASTReader) setAST(tree ast.Node) {
	r.reset()
	r.pushRoutePoint(routePoint{
		node: tree,
	})
}

func (r *ASTReader) TryNull() bool {
	if r.hasFatalError() {
		return false
	}
	r.currentExpecter = expectNull{}
	r.visitCurrentNode()
	if r.hasFatalError() {
		return false
	}
	if r.latestDenyError != nil {
		r.latestDenyError = nil
		return false
	}
	return true
}

func (r *ASTReader) Integer(bitSize int) int64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectInteger{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return 0
	}
	v, err := schema.ToInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) Unsigned(bitSize int) uint64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectInteger{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return 0
	}
	v, err := schema.ToUnsignedInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) Boolean() bool {
	if r.hasFatalError() {
		return false
	}
	r.currentExpecter = expectBoolean{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return false
	}
	v, err := schema.ToBoolean(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return false
	}
	return v
}

func (r *ASTReader) Float(bitSize int) float64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectFloat{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return 0
	}
	v, err := schema.ToFloat(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) String() string {
	if r.hasFatalError() {
		return ""
	}
	r.currentExpecter = expectString{checkForNull: true}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return ""
	}
	return r.extractedValue
}

func (r *ASTReader) Timestamp() time.Time {
	if r.hasFatalError() {
		return time.Time{}
	}
	r.currentExpecter = expectTimestamp{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return time.Time{}
	}
	v, err := schema.ToTimestamp(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return time.Time{}
	}
	return v
}

func (r *ASTReader) Sequence() yamly.CollectionState {
	if r.hasFatalError() {
		return noopCollectionState
	}
	r.currentExpecter = expectSequence{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return noopCollectionState
	}
	return r.extractedCollectionState
}

func (r *ASTReader) Mapping() yamly.CollectionState {
	if r.hasFatalError() {
		return noopCollectionState
	}
	r.currentExpecter = expectMapping{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return noopCollectionState
	}
	return r.extractedCollectionState
}

func (r *ASTReader) Any() any {
	if r.hasFatalError() {
		return nil
	}
	r.currentExpecter = expectAny{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return nil
	}
	valueBuilder := newAnyBuilder(&r.anchors)
	v, err := valueBuilder.extractAnyValue(r.currentNode())
	if err != nil {
		r.appendError(err)
		return nil
	}
	r.popRoutePoint()
	return v
}

func (r *ASTReader) Raw() []byte {
	if r.hasFatalError() {
		return nil
	}
	r.currentExpecter = expectRaw{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return nil
	}
	w := encode.NewASTWriter()
	v, err := w.WriteBytes(r.currentNode())
	if err != nil {
		r.appendError(err)
		return nil
	}
	r.popRoutePoint()
	return v
}

func (r *ASTReader) Node() ast.Node {
	if r.hasFatalError() {
		return nil
	}
	r.currentExpecter = expectNode{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
		return nil
	}
	n := r.currentNode()
	r.popRoutePoint()
	return n
}

func (r *ASTReader) Skip() {
	if r.hasFatalError() {
		return
	}
	r.currentExpecter = expectSkip{}
	r.visitCurrentNode()
	if r.latestDenyError != nil || r.hasFatalError() {
		r.appendError(r.latestDenyError)
		r.latestDenyError = nil
	}
}

func (r *ASTReader) VisitStreamNode(n *ast.StreamNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newStreamIterator(n)
	}

	r.processComplexPoint(point, len(n.Documents()))
}

func (r *ASTReader) VisitTagNode(n *ast.TagNode) {
	r.visitTexterNode(n)
}

func (r *ASTReader) VisitAnchorNode(n *ast.AnchorNode) {
	r.anchors.StoreAnchor(n.Text())

	r.visitTexterNode(n)
}

func (r *ASTReader) VisitAliasNode(n *ast.AliasNode) {
	anchored, err := r.anchors.DereferenceAlias(n.Text())
	if err != nil {
		r.appendError(err)
	} else {
		r.pushRoutePoint(routePoint{
			node: anchored,
		})
		anchored.Accept(r)
	}
}

func (r *ASTReader) VisitTextNode(n *ast.TextNode) {
	r.visitTexterNode(n)
}

func (r *ASTReader) VisitSequenceNode(n *ast.SequenceNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newSequenceIterator(n)
	}

	r.processComplexPoint(point, len(n.Entries()))
}

func (r *ASTReader) VisitMappingNode(n *ast.MappingNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newMappingIterator(n)
	}

	r.processComplexPoint(point, len(n.Entries()))
}

func (r *ASTReader) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newMappingEntryIterator(n)
	}

	r.processComplexPoint(point, 2)
}

func (r *ASTReader) VisitNullNode(n *ast.NullNode) {
	point := r.peekRoutePoint()

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult.conclusion {
	case visitingConclusionConsume:
		r.popRoutePoint()
	case visitingConclusionMatch:
	case visitingConclusionContinue:
		r.popRoutePoint()
	case visitingConclusionDeny:
		r.swapRoutePoint(point)
		r.appendError(yamly.DenyError(&denyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		}))
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %s", point.visitingResult.conclusion))
	}

	if point.visitingResult.action == visitingActionExtract {
		r.extractedCollectionState = nil
		r.extractedValue = ""
	}
}

func (r *ASTReader) VisitPropertiesNode(n *ast.PropertiesNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newPropertiesIterator(n)
	}

	r.processComplexPoint(point, 2)
}

func (r *ASTReader) VisitContentNode(n *ast.ContentNode) {
	point := r.peekRoutePoint()
	if point.visitingResult.conclusion == visitingConclusionUnknown {
		point.iter = newContentIterator(n)
	}

	r.processComplexPoint(point, 2, beforeVisit(r.anchors.BindToLatestAnchor))
}

func (r *ASTReader) visitTexterNode(n ast.TexterNode) {
	point := r.peekRoutePoint()

	point.visitingResult = r.currentExpecter.process(n, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	switch point.visitingResult.conclusion {
	case visitingConclusionConsume:
		r.popRoutePoint()
	case visitingConclusionMatch:
	case visitingConclusionDeny:
		r.swapRoutePoint(point)
		r.setLatestDeny(&denyError{
			expecter: r.currentExpecter,
			nt:       point.node.Type(),
		})
	case visitingConclusionContinue:
		r.popRoutePoint()
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %v", point.visitingResult.conclusion))
	}

	if point.visitingResult.action == visitingActionExtract {
		r.extractedCollectionState = nil
		r.extractedValue = n.Text()
	}
}

type complexPointOptions struct {
	beforeVisitFuncs []func(ast.Node)
}

type complexPointOption func(*complexPointOptions)

func beforeVisit(f func(ast.Node)) complexPointOption {
	return complexPointOption(func(options *complexPointOptions) {
		options.beforeVisitFuncs = append(options.beforeVisitFuncs, f)
	})
}

func (r *ASTReader) processComplexPoint(point routePoint, childrenSize int, opts ...complexPointOption) {
	var o complexPointOptions
	for _, opt := range opts {
		opt(&o)
	}

	point.visitingResult = r.currentExpecter.process(point.node, point.visitingResult)
	r.lastVisitingResult = point.visitingResult
	r.swapRoutePoint(point)
	switch point.visitingResult.conclusion {
	case visitingConclusionConsume:
		r.popRoutePoint()
	case visitingConclusionMatch:
	case visitingConclusionDeny:
		r.setLatestDeny(&denyError{
			expecter: r.currentExpecter,
			nt:       point.node.Type(),
		})
	case visitingConclusionContinue:
		for r.lastVisitingResult.conclusion == visitingConclusionContinue && !point.iter.empty() {
			node := point.iter.node()

			for _, beforeVisitFunc := range o.beforeVisitFuncs {
				beforeVisitFunc(node)
			}

			if ast.ValidNode(node) {
				r.pushRoutePoint(routePoint{
					node: node,
				})
				node.Accept(r)
			}
		}
		if point.iter.empty() && r.lastVisitingResult.conclusion == visitingConclusionContinue {
			r.popRoutePoint()
			r.visitCurrentNode()
		}
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %s", point.visitingResult.conclusion))
	}

	if point.visitingResult.action == visitingActionExtract {
		r.extractedCollectionState = newCollectionState(point.iter, childrenSize)
		r.extractedValue = ""
	}
}

func (r *ASTReader) visitCurrentNode() {
	n := r.currentNode()
	if n == nil {
		r.appendError(yamly.ErrEndOfStream)
	} else {
		n.Accept(r)
	}
}

func (r *ASTReader) reset() {
	r.route = r.route[:0]
	r.currentExpecter = nil
	r.lastVisitingResult = visitingResult{}
	r.extractedCollectionState = nil
	r.extractedValue = ""
	r.anchors.clear()
	r.fatalError = nil
	r.latestDenyError = nil
	r.denyErrors = r.denyErrors[:0]
}

func (r *ASTReader) currentNode() ast.Node {
	if len(r.route) == 0 {
		return nil
	}
	return r.route[len(r.route)-1].node
}

func (r *ASTReader) pushRoutePoint(point routePoint) {
	r.route = append(r.route, point)
}

func (r *ASTReader) popRoutePoint() routePoint {
	if len(r.route) == 0 {
		return routePoint{}
	}
	point := r.route[len(r.route)-1]
	r.route = r.route[:len(r.route)-1]
	return point
}

func (r *ASTReader) peekRoutePoint() routePoint {
	if len(r.route) == 0 {
		return routePoint{}
	}
	point := r.route[len(r.route)-1]
	return point
}

func (r *ASTReader) swapRoutePoint(point routePoint) {
	if len(r.route) == 0 {
		return
	}
	r.route[len(r.route)-1] = point
}

func (r *ASTReader) setLatestDeny(err *denyError) {
	r.latestDenyError = yamly.DenyError(err)
}

func (r *ASTReader) appendError(err error) {
	if r.multipleDenyErrors && errors.Is(err, yamly.ErrDenied) {
		r.denyErrors = append(r.denyErrors, err)
	} else if r.fatalError == nil {
		r.fatalError = err
	}
}

func (r *ASTReader) hasFatalError() bool {
	return r.fatalError != nil
}

func (r *ASTReader) hasAnyError() bool {
	return r.fatalError != nil || len(r.denyErrors) != 0
}

func (r *ASTReader) Error() error {
	return errors.Join(append([]error{r.fatalError}, r.denyErrors...)...)
}

func (r *ASTReader) AddError(err error) {
	if r.fatalError == nil {
		r.fatalError = err
	}
}
