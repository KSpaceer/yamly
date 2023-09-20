// Package cpaccessor contains methods to access streams with checkpointing conception.
package cpaccessor

// ResourceStream represents an underlying stream for accessing elements.
type ResourceStream[T any] interface {
	Next() T
}

// CheckpointingAccessor is used to access stream with checkpoints.
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

func NewCheckpointingAccessor[T any]() CheckpointingAccessor[T] {
	return CheckpointingAccessor[T]{
		buf:              make([]T, 0, bufferPreallocationSize),
		bufIndicator:     withoutBuffer,
		checkpointsStack: make([]int, 0, checkpointsStackPreallocationSize),
	}
}

func (a *CheckpointingAccessor[T]) SetStream(stream ResourceStream[T]) {
	a.Reset()
	a.stream = stream
}

func (a *CheckpointingAccessor[T]) Reset() {
	a.stream = nil
	a.buf = a.buf[:0]
	a.bufIndicator = withoutBuffer
	a.checkpointsStack = a.checkpointsStack[:0]
}

// Next moves accessor to the next element in the stream.
func (a *CheckpointingAccessor[T]) Next() T {
	var val T
	if a.bufIndicator == withoutBuffer { // nolint: nestif
		// if there is no buffer or we are ahead of it, get next element from the stream
		val = a.stream.Next()
		if len(a.checkpointsStack) > 0 {
			// if there any checkpoint, store element in buffer
			a.buf = append(a.buf, val)
		} else {
			// otherwise, remember it for the first checkpoint (to be able to rollback to it).
			a.saved = val
		}
	} else {
		// get element from the buffer
		val = a.buf[a.bufIndicator]
		a.bufIndicator++
		if a.bufIndicator == len(a.buf) {
			if len(a.checkpointsStack) == 0 {
				// there are no checkpoints - buffer elements will not be used anymore
				a.buf = a.buf[:0]
			}
			a.bufIndicator = withoutBuffer
		}
	}
	return val
}

// SetCheckpoint sets checkpoint at current position in the stream.
func (a *CheckpointingAccessor[T]) SetCheckpoint() {
	if a.bufIndicator == withoutBuffer {
		// if we are ahead of buffer, set checkpoint to it's last position (or show that no buffer used).
		a.checkpointsStack = append(a.checkpointsStack, len(a.buf)-1)
	} else {
		a.checkpointsStack = append(a.checkpointsStack, a.bufIndicator-1)
	}
}

// Rollback moves accessor to the latest uncommited checkpoint and removes checkpoint.
// Returned value - is the last element before checkpoint set.
func (a *CheckpointingAccessor[T]) Rollback() T {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
		return a.saved
	default:
		a.bufIndicator = a.checkpointsStack[stackLen-1]
		var restoredVal T
		if a.bufIndicator == withoutBuffer {
			// using saved value before buffer start
			restoredVal = a.saved
			if len(a.buf) > 0 {
				// since we returned to state before buffer even started,
				// move to it beginning position
				a.bufIndicator = 0
			}
		} else {
			// using element at stored position
			restoredVal = a.buf[a.bufIndicator]
			a.bufIndicator++
			if a.bufIndicator == len(a.buf) {
				// checkpoint was at the last position in buffer - we are now ahead of it.
				if len(a.checkpointsStack) == 0 {
					a.buf = a.buf[:0]
				}
				a.bufIndicator = withoutBuffer
			}
		}
		// remove checkpoint
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
		return restoredVal
	}
}

// Commit marks the latest checkpoint as successful and removes it.
// Further Rollback calls will not return accessor to this checkpoint.
func (a *CheckpointingAccessor[T]) Commit() {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
	case 1:
		if a.bufIndicator == withoutBuffer && len(a.buf) != 0 {
			// there not checkpoints anymore and we are ahead of buffer.
			// we have to remember the last value to be able to rollback to it.
			a.saved = a.buf[len(a.buf)-1]
			a.buf = a.buf[:0]
		}
		fallthrough
	default:
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
	}
}
