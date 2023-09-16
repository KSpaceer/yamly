package decode

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/goyaml/schema"
	"gopkg.in/yaml.v3"
	"time"
)

var _ yamly.ExtendedDecoder[*yaml.Node] = (*ASTReader)(nil)

type ASTReader struct {
	route []routePoint

	currentExpecter    expecter
	lastVisitingResult visitingResult

	extractedCollectionState yamly.CollectionState
	extractedValue           string

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
	node           *yaml.Node
	iter           nodeIterator
}

type expecter interface {
	name() string
	process(n *yaml.Node, previousResult visitingResult) visitingResult
}

type nodeIterator interface {
	node() *yaml.Node
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
	return func(reader *ASTReader) {
		reader.multipleDenyErrors = true
	}
}

func NewASTReader(tree *yaml.Node, opts ...ReaderOption) *ASTReader {
	r := ASTReader{}

	for _, opt := range opts {
		opt(&r)
	}

	r.setAST(tree)
	return &r
}

func (r *ASTReader) setAST(tree *yaml.Node) {
	r.reset()
	if tree.Kind == yaml.DocumentNode {
		tree = tree.Content[0]
	}
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
	valueBuilder := anyBuilder{}
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
	v, err := yaml.Marshal(r.currentNode())
	if err != nil {
		r.appendError(err)
		return nil
	}
	r.popRoutePoint()
	return v
}

func (r *ASTReader) Node() *yaml.Node {
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

type complexPointOptions struct {
	beforeVisitFuncs []func(*yaml.Node)
}

type complexPointOption func(*complexPointOptions)

func beforeVisit(f func(*yaml.Node)) complexPointOption {
	return func(options *complexPointOptions) {
		options.beforeVisitFuncs = append(options.beforeVisitFuncs, f)
	}
}

func (r *ASTReader) visitNode(n *yaml.Node) {
	switch n.Kind {
	case yaml.DocumentNode:
		r.visitNode(n.Content[0])
	case yaml.SequenceNode, yaml.MappingNode:
		point := r.peekRoutePoint()
		if point.visitingResult.conclusion == visitingConclusionUnknown {
			point.iter = newNodeIterator(n)
		}
		r.processComplexPoint(point, len(n.Content))
	case yaml.ScalarNode:
		r.visitScalarNode(n)
	case yaml.AliasNode:
		r.pushRoutePoint(routePoint{
			node: n.Alias,
		})
		r.visitNode(n.Alias)
	}
}

func (r *ASTReader) visitScalarNode(n *yaml.Node) {
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
			expecter:    r.currentExpecter,
			nodeContent: point.node.Value,
		})
	case visitingConclusionContinue:
		r.popRoutePoint()
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %v", point.visitingResult.conclusion))
	}

	switch point.visitingResult.action {
	case visitingActionExtract:
		r.extractedCollectionState = nil
		r.extractedValue = n.Value
	}
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
			expecter:    r.currentExpecter,
			nodeContent: point.node.Value,
		})
	case visitingConclusionContinue:
		for r.lastVisitingResult.conclusion == visitingConclusionContinue && !point.iter.empty() {
			node := point.iter.node()

			for _, beforeVisitFunc := range o.beforeVisitFuncs {
				beforeVisitFunc(node)
			}

			if node != nil {
				r.pushRoutePoint(routePoint{
					node: node,
				})
				r.visitNode(node)
			}
		}
		if point.iter.empty() && r.lastVisitingResult.conclusion == visitingConclusionContinue {
			r.popRoutePoint()
		}
	default:
		r.appendError(fmt.Errorf("unexpected conclusion: %s", point.visitingResult.conclusion))
	}

	switch point.visitingResult.action {
	case visitingActionExtract:
		r.extractedCollectionState = newCollectionState(point.iter, childrenSize)
		r.extractedValue = ""
	}
}

func (r *ASTReader) reset() {
	r.route = r.route[:0]
	r.currentExpecter = nil
	r.lastVisitingResult = visitingResult{}
	r.extractedCollectionState = nil
	r.extractedValue = ""
	r.fatalError = nil
	r.latestDenyError = nil
	r.denyErrors = r.denyErrors[:0]
}

func (r *ASTReader) visitCurrentNode() {
	n := r.currentNode()
	if n == nil {
		r.appendError(yamly.EndOfStream)
	} else {
		r.visitNode(n)
	}
}

func (r *ASTReader) currentNode() *yaml.Node {
	if len(r.route) == 0 {
		return nil
	}
	return r.route[len(r.route)-1].node
}

func (r *ASTReader) pushRoutePoint(point routePoint) {
	r.route = append(r.route, point)
}

func (r *ASTReader) popRoutePoint() {
	if len(r.route) == 0 {
		return
	}
	r.route = r.route[:len(r.route)-1]
	return
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
	if r.multipleDenyErrors && errors.Is(err, yamly.Denied) {
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
