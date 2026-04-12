package stream

import "sync"

type Subscriber chan StatusUpdate

type StatusUpdate struct {
	OrderID string
	Status  string
	Message string
}

type OrderStreamManager struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
}

func NewOrderStreamManager() *OrderStreamManager {
	return &OrderStreamManager{
		subscribers: make(map[string][]Subscriber),
	}
}

func (m *OrderStreamManager) Subscribe(orderID string) Subscriber {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(Subscriber, 10)
	m.subscribers[orderID] = append(m.subscribers[orderID], ch)
	return ch
}

func (m *OrderStreamManager) Unsubscribe(orderID string, sub Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subs := m.subscribers[orderID]
	newSubs := make([]Subscriber, 0, len(subs))

	for _, s := range subs {
		if s != sub {
			newSubs = append(newSubs, s)
		}
	}

	m.subscribers[orderID] = newSubs
	close(sub)
}

func (m *OrderStreamManager) Publish(update StatusUpdate) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sub := range m.subscribers[update.OrderID] {
		select {
		case sub <- update:
		default:
		}
	}
}
