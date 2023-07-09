package cpaccessor

type ResourceStream[T any] interface {
	Next() T
}

type CheckpointingAccessor[T any] struct {
	stream           ResourceStream[T]
	buf              []T
	saved            T
	bufIndicator     int
	checkpointsStack []int
}

const (
	bufferPreallocationSize           = 32
	checkpointsStackPreallocationSize = 16

	withoutBuffer = -1
)

func NewCheckpointingAccessor[T any](stream ResourceStream[T]) CheckpointingAccessor[T] {
	return CheckpointingAccessor[T]{
		stream:           stream,
		buf:              make([]T, 0, bufferPreallocationSize),
		bufIndicator:     withoutBuffer,
		checkpointsStack: make([]int, 0, checkpointsStackPreallocationSize),
	}
}

func (a *CheckpointingAccessor[T]) Next() T {
	var val T
	if a.bufIndicator == withoutBuffer {
		val = a.stream.Next()
		if len(a.checkpointsStack) > 0 {
			a.buf = append(a.buf, val)
		} else {
			a.saved = val
		}
	} else {
		val = a.buf[a.bufIndicator]
		a.bufIndicator++
		if a.bufIndicator == len(a.buf) {
			if len(a.checkpointsStack) == 0 {
				a.buf = a.buf[:0]
			}
			a.bufIndicator = withoutBuffer
		}
	}
	return val
}

func (a *CheckpointingAccessor[T]) SetCheckpoint() {
	if a.bufIndicator == withoutBuffer {
		a.checkpointsStack = append(a.checkpointsStack, len(a.buf)-1)
	} else {
		a.checkpointsStack = append(a.checkpointsStack, a.bufIndicator-1)
	}
}

func (a *CheckpointingAccessor[T]) Rollback() T {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
		return a.saved
	default:
		a.bufIndicator = a.checkpointsStack[stackLen-1]
		var restoredVal T
		if a.bufIndicator == withoutBuffer {
			restoredVal = a.saved
			if len(a.buf) > 0 {
				a.bufIndicator = 0
			}
		} else {
			restoredVal = a.buf[a.bufIndicator]
			a.bufIndicator++
			if a.bufIndicator == len(a.buf) {
				if len(a.checkpointsStack) == 0 {
					a.buf = a.buf[:0]
				}
				a.bufIndicator = withoutBuffer
			}
		}
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
		return restoredVal
	}
}

func (a *CheckpointingAccessor[T]) Commit() {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
	case 1:
		if a.bufIndicator == withoutBuffer && len(a.buf) != 0 {
			a.saved = a.buf[len(a.buf)-1]
		}
		a.bufIndicator = -1
		a.buf = a.buf[:0]
		fallthrough
	default:
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
	}
}
