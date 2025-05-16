package broadcast

import (
	"sync"
)

type Broadcaster[T any] struct {
	mu          sync.Mutex
	subscribers map[chan T]struct{}
}

func New[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subscribers: make(map[chan T]struct{}),
	}
}

// Subscribe returns a new channel that receives broadcasts
func (b *Broadcaster[T]) Subscribe() <-chan T {
	ch := make(chan T, 8) // buffered to prevent blocking
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber and closes its channel
func (b *Broadcaster[T]) Unsubscribe(ch chan T) {
	b.mu.Lock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.mu.Unlock()
}

// Publish sends the data to all subscribers (non-blocking)
func (b *Broadcaster[T]) Publish(data T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subscribers {
		func(ch chan T) {
			defer func() {
				if r := recover(); r != nil {
					// Channel is closed, remove it
					delete(b.subscribers, ch)
				}
			}()

			select {
			case ch <- data:
				// sent
			default:
				// drop if full
			}
		}(ch)
	}
}
