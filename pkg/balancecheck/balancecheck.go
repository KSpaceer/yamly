package balancecheck

type BalanceChecker[T comparable] struct {
	openers          map[T]T
	closers          map[T]struct{}
	stack            []T
	cannotBeBalanced bool
}

const preallocationSize = 8

func NewBalanceChecker[T comparable](pairs [][2]T) BalanceChecker[T] {
	b := BalanceChecker[T]{
		openers: make(map[T]T, len(pairs)),
		closers: make(map[T]struct{}, len(pairs)),
		stack:   make([]T, 0, preallocationSize),
	}
	for _, pair := range pairs {
		b.openers[pair[0]] = pair[1]
		b.closers[pair[1]] = struct{}{}
	}
	return b
}

func (b *BalanceChecker[T]) Add(r T) bool {
	if b.cannotBeBalanced {
		return false
	}
	_, isCloser := b.closers[r]
	_, isOpener := b.openers[r]
	canPop := len(b.stack) > 0 && b.openers[b.stack[len(b.stack)-1]] == r
	if isCloser && !isOpener && !canPop {
		b.cannotBeBalanced = true
		return false
	}
	if isCloser && canPop {
		b.stack = b.stack[:len(b.stack)-1]
	} else if isOpener {
		b.stack = append(b.stack, r)
	}
	return true
}

func (b *BalanceChecker[T]) IsBalanced() bool {
	return !b.cannotBeBalanced && len(b.stack) == 0
}

func (b *BalanceChecker[T]) PeekLastUnbalanced() (T, bool) {
	var peeked T
	if len(b.stack) == 0 {
		return peeked, false
	}
	peeked = b.stack[len(b.stack)-1]
	return peeked, true
}

type BalanceCheckerMemento struct {
	stackSize        int
	cannotBeBalanced bool
}

func (b *BalanceChecker[T]) Memento() BalanceCheckerMemento {
	return BalanceCheckerMemento{
		stackSize:        len(b.stack),
		cannotBeBalanced: b.cannotBeBalanced,
	}
}

func (b *BalanceChecker[T]) SetMemento(m BalanceCheckerMemento) {
	stackSize := m.stackSize
	if cap(b.stack) < stackSize {
		stackSize = cap(b.stack)
	}
	b.stack = b.stack[:stackSize]
	b.cannotBeBalanced = m.cannotBeBalanced
}
