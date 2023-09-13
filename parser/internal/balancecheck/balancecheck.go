package balancecheck

import "github.com/KSpaceer/yamly/token"

type BalanceChecker struct {
	openers          map[token.Type]token.Type
	closers          map[token.Type]struct{}
	stack            []token.Type
	cannotBeBalanced bool
}

const preallocationSize = 8

func NewBalanceChecker(pairs [][2]token.Type) BalanceChecker {
	b := BalanceChecker{
		openers: make(map[token.Type]token.Type, len(pairs)),
		closers: make(map[token.Type]struct{}, len(pairs)),
		stack:   make([]token.Type, 0, preallocationSize),
	}
	for _, pair := range pairs {
		b.openers[pair[0]] = pair[1]
		b.closers[pair[1]] = struct{}{}
	}
	return b
}

func (b *BalanceChecker) Add(r token.Type) bool {
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

func (b *BalanceChecker) IsBalanced() bool {
	return !b.cannotBeBalanced && len(b.stack) == 0
}

func (b *BalanceChecker) PeekLastUnbalanced() (token.Type, bool) {
	var peeked token.Type
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

func (b *BalanceChecker) Memento() BalanceCheckerMemento {
	return BalanceCheckerMemento{
		stackSize:        len(b.stack),
		cannotBeBalanced: b.cannotBeBalanced,
	}
}

func (b *BalanceChecker) SetMemento(m BalanceCheckerMemento) {
	stackSize := m.stackSize
	if cap(b.stack) < stackSize {
		stackSize = cap(b.stack)
	}
	b.stack = b.stack[:stackSize]
	b.cannotBeBalanced = m.cannotBeBalanced
}

func (b *BalanceChecker) Reset() {
	b.stack = b.stack[:0]
}
