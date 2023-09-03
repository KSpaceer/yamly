package decode

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/encode"
	"github.com/KSpaceer/yayamls/schema"
	"time"
)

type ASTReader struct {
	route []routePoint

	currentExpecter    expecter
	lastVisitingResult visitingResult

	extractedCollectionState yayamls.CollectionState
	extractedValue           string
	isExtractedNull          bool

	anchors anchorsKeeper

	multipleDenyErrors bool
	fatalError         error
	denyErrors         []error
	denied             bool
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
	visitingActionSetNull
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

var noopCollectionState yayamls.CollectionState = noopState{}

type noopState struct{}

func (noopState) Size() int { return 0 }

func (noopState) HasUnprocessedItems() bool { return false }

func newCollectionState(iter nodeIterator, size int) yayamls.CollectionState {
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

func (r *ASTReader) ExpectInteger(bitSize int) int64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectInteger{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0
	}
	v, err := schema.ToInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) ExpectNullableInteger(bitSize int) (int64, bool) {
	if r.hasFatalError() {
		return 0, false
	}
	r.currentExpecter = expectNullable{underlying: expectInteger{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0, false
	}

	if r.isExtractedNull {
		return 0, false
	}
	v, err := schema.ToInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0, false
	}
	return v, true
}

func (r *ASTReader) ExpectUnsigned(bitSize int) uint64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectInteger{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0
	}
	v, err := schema.ToUnsignedInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) ExpectNullableUnsigned(bitSize int) (uint64, bool) {
	if r.hasFatalError() {
		return 0, false
	}
	r.currentExpecter = expectNullable{underlying: expectInteger{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0, false
	}
	if r.isExtractedNull {
		return 0, false
	}
	v, err := schema.ToUnsignedInteger(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0, false
	}
	return v, true
}

func (r *ASTReader) ExpectBoolean() bool {
	if r.hasFatalError() {
		return false
	}
	r.currentExpecter = expectBoolean{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return false
	}
	v, err := schema.ToBoolean(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return false
	}
	return v
}

func (r *ASTReader) ExpectNullableBoolean() (bool, bool) {
	if r.hasFatalError() {
		return false, false
	}
	r.currentExpecter = expectNullable{underlying: expectBoolean{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return false, false
	}
	if r.isExtractedNull {
		return false, false
	}
	v, err := schema.ToBoolean(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return false, false
	}
	return v, true
}

func (r *ASTReader) ExpectFloat(bitSize int) float64 {
	if r.hasFatalError() {
		return 0
	}
	r.currentExpecter = expectFloat{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0
	}
	v, err := schema.ToFloat(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0
	}
	return v
}

func (r *ASTReader) ExpectNullableFloat(bitSize int) (float64, bool) {
	if r.hasFatalError() {
		return 0, false
	}
	r.currentExpecter = expectNullable{underlying: expectFloat{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return 0, false
	}
	if r.isExtractedNull {
		return 0, false
	}
	v, err := schema.ToFloat(r.extractedValue, bitSize)
	if err != nil {
		r.appendError(err)
		return 0, false
	}
	return v, true
}

func (r *ASTReader) ExpectString() string {
	if r.hasFatalError() {
		return ""
	}
	r.currentExpecter = expectString{checkForNull: true}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return ""
	}
	return r.extractedValue
}

func (r *ASTReader) ExpectNullableString() (string, bool) {
	if r.hasFatalError() {
		return "", false
	}
	r.currentExpecter = expectNullable{underlying: expectString{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return "", false
	}
	if r.isExtractedNull {
		return "", false
	}
	return r.extractedValue, true
}

func (r *ASTReader) ExpectTimestamp() time.Time {
	if r.hasFatalError() {
		return time.Time{}
	}
	r.currentExpecter = expectTimestamp{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return time.Time{}
	}
	v, err := schema.ToTimestamp(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return time.Time{}
	}
	return v
}

func (r *ASTReader) ExpectNullableTimestamp() (time.Time, bool) {
	if r.hasFatalError() {
		return time.Time{}, false
	}
	r.currentExpecter = expectNullable{underlying: expectTimestamp{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return time.Time{}, false
	}
	if r.isExtractedNull {
		return time.Time{}, false
	}
	v, err := schema.ToTimestamp(r.extractedValue)
	if err != nil {
		r.appendError(err)
		return time.Time{}, false
	}
	return v, true
}

func (r *ASTReader) ExpectSequence() yayamls.CollectionState {
	if r.hasFatalError() {
		return noopCollectionState
	}
	r.currentExpecter = expectSequence{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return noopCollectionState
	}
	return r.extractedCollectionState
}

func (r *ASTReader) ExpectNullableSequence() (yayamls.CollectionState, bool) {
	if r.hasFatalError() {
		return nil, false
	}
	r.currentExpecter = expectNullable{underlying: expectSequence{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return nil, false
	}
	if r.isExtractedNull {
		return nil, false
	}
	return r.extractedCollectionState, true
}

func (r *ASTReader) ExpectMapping() yayamls.CollectionState {
	if r.hasFatalError() {
		return noopCollectionState
	}
	r.currentExpecter = expectMapping{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return noopCollectionState
	}
	return r.extractedCollectionState
}

func (r *ASTReader) ExpectNullableMapping() (yayamls.CollectionState, bool) {
	if r.hasFatalError() {
		return nil, false
	}
	r.currentExpecter = expectNullable{underlying: expectMapping{}}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
		return nil, false
	}
	if r.isExtractedNull {
		return nil, false
	}
	return r.extractedCollectionState, true
}

func (r *ASTReader) ExpectAny() any {
	if r.hasFatalError() {
		return nil
	}
	r.currentExpecter = expectAny{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
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

func (r *ASTReader) ExpectRaw() []byte {
	if r.hasFatalError() {
		return nil
	}
	r.currentExpecter = expectRaw{}
	r.visitCurrentNode()
	if r.denied || r.hasFatalError() {
		r.denied = false
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
		r.appendError(yayamls.DenyError(&denyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		}))
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %s", point.visitingResult.conclusion))
	}

	switch point.visitingResult.action {
	case visitingActionExtract:
		r.extractedCollectionState = nil
		r.extractedValue = ""
		r.isExtractedNull = false
	case visitingActionSetNull:
		r.isExtractedNull = true
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
		r.appendError(yayamls.DenyError(&denyError{
			expecter: r.currentExpecter,
			nt:       n.Type(),
		}))
	case visitingConclusionContinue:
		r.popRoutePoint()
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %v", point.visitingResult.conclusion))
	}

	switch point.visitingResult.action {
	case visitingActionExtract:
		r.extractedCollectionState = nil
		r.extractedValue = n.Text()
		r.isExtractedNull = false
	case visitingActionSetNull:
		r.isExtractedNull = true
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
		r.appendError(yayamls.DenyError(&denyError{
			expecter: r.currentExpecter,
			nt:       point.node.Type(),
		}))
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

	switch point.visitingResult.action {
	case visitingActionExtract:
		r.extractedCollectionState = newCollectionState(point.iter, childrenSize)
		r.extractedValue = ""
		r.isExtractedNull = false
	case visitingActionSetNull:
		r.isExtractedNull = true
	}
}

func (r *ASTReader) visitCurrentNode() {
	n := r.currentNode()
	if n == nil {
		r.appendError(yayamls.EndOfStream)
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
	r.isExtractedNull = false
	r.anchors.clear()
	r.denied = false
	r.fatalError = nil
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

func (r *ASTReader) appendError(err error) {
	if r.multipleDenyErrors && errors.Is(err, yayamls.Denied) {
		r.denied = true
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
