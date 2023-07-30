package balanceaccount

type BalanceAccounter struct {
	openers          map[rune]rune
	closers          map[rune]struct{}
	stack            []rune
	cannotBeBalanced bool
}

const preallocationSize = 8

func NewBalancer(pairs [][2]rune) *BalanceAccounter {
	b := BalanceAccounter{
		openers: make(map[rune]rune, len(pairs)),
		closers: make(map[rune]struct{}, len(pairs)),
		stack:   make([]rune, 0, preallocationSize),
	}
	for _, pair := range pairs {
		b.openers[pair[0]] = pair[1]
		b.closers[pair[1]] = struct{}{}
	}
	return &b
}

func (b *BalanceAccounter) AccountRune(r rune) bool {
	if b.cannotBeBalanced {
		return false
	}
	_, isCloser := b.closers[r]
	if isCloser {
		if len(b.stack) == 0 || b.openers[b.stack[len(b.stack)-1]] != r {
			b.cannotBeBalanced = true
			return false
		}
		b.stack = b.stack[:len(b.stack)-1]
	}
	_, isOpener := b.openers[r]
	if isOpener {
		b.stack = append(b.stack, r)
	}
	return true
}

func (b *BalanceAccounter) IsBalanced() bool {
	return !b.cannotBeBalanced && len(b.stack) == 0
}

type BalanceAccountMemento struct {
	stackSize        int
	cannotBeBalanced bool
}

func (b *BalanceAccounter) Memento() BalanceAccountMemento {
	return BalanceAccountMemento{
		stackSize:        len(b.stack),
		cannotBeBalanced: b.cannotBeBalanced,
	}
}

func (b *BalanceAccounter) SetMemento(m BalanceAccountMemento) {
	stackSize := m.stackSize
	if cap(b.stack) < stackSize {
		stackSize = cap(b.stack)
	}
	b.stack = b.stack[:stackSize]
	b.cannotBeBalanced = m.cannotBeBalanced
}
