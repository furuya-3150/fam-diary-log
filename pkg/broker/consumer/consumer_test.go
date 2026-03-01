package consumer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// --- モック ---

// mockAcknowledger は amqp.Acknowledger のテスト用モック
type mockAcknowledger struct {
	mu              sync.Mutex
	ackCount        int
	nackCount       int
	lastNackRequeue bool
}

func (m *mockAcknowledger) Ack(_ uint64, _ bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ackCount++
	return nil
}

func (m *mockAcknowledger) Nack(_ uint64, _ bool, requeue bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nackCount++
	m.lastNackRequeue = requeue
	return nil
}

func (m *mockAcknowledger) Reject(_ uint64, _ bool) error { return nil }

// mockHandler は EventHandler のテスト用モック
type mockHandler struct {
	mu  sync.Mutex
	err error
}

func (h *mockHandler) Handle(_ context.Context, _ string, _ []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.err
}

// --- ヘルパー ---

func newTestConsumer(maxRetries int, handler EventHandler) *RabbitMQConsumer {
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}
	return &RabbitMQConsumer{
		config: Config{
			ExchangeName: "test.exchange",
			QueueName:    "test.queue",
			RoutingKeys:  []string{"test.key"},
			MaxRetries:   maxRetries,
		},
		handler:     handler,
		l:           slog.Default(),
		retryCounts: make(map[string]int),
	}
}

func newDelivery(body []byte, ack *mockAcknowledger) amqp.Delivery {
	return amqp.Delivery{
		Acknowledger: ack,
		Body:         body,
		RoutingKey:   "test.key",
	}
}

// --- テスト ---

// TestHandleMessage_Success は正常処理で Ack が呼ばれることを確認する
func TestHandleMessage_Success(t *testing.T) {
	c := newTestConsumer(3, &mockHandler{err: nil})
	ack := &mockAcknowledger{}

	c.handleMessage(context.Background(), newDelivery([]byte("hello"), ack))

	if ack.ackCount != 1 {
		t.Errorf("expected ackCount=1, got %d", ack.ackCount)
	}
	if ack.nackCount != 0 {
		t.Errorf("expected nackCount=0, got %d", ack.nackCount)
	}
}

// TestHandleMessage_RetryUnderLimit はリトライ上限未満で requeue=true の Nack が呼ばれることを確認する
func TestHandleMessage_RetryUnderLimit(t *testing.T) {
	c := newTestConsumer(3, &mockHandler{err: errors.New("temporary error")})
	ack := &mockAcknowledger{}

	c.handleMessage(context.Background(), newDelivery([]byte("hello"), ack))

	if ack.nackCount != 1 {
		t.Errorf("expected nackCount=1, got %d", ack.nackCount)
	}
	if !ack.lastNackRequeue {
		t.Error("expected requeue=true for retry under limit")
	}
}

// TestHandleMessage_MaxRetriesExceeded はリトライ上限到達で requeue=false の Nack が呼ばれることを確認する
func TestHandleMessage_MaxRetriesExceeded(t *testing.T) {
	const maxRetries = 3
	c := newTestConsumer(maxRetries, &mockHandler{err: errors.New("permanent error")})
	body := []byte("hello")

	for i := 0; i < maxRetries; i++ {
		ack := &mockAcknowledger{}
		c.handleMessage(context.Background(), newDelivery(body, ack))
		if i < maxRetries-1 {
			if !ack.lastNackRequeue {
				t.Errorf("call %d: expected requeue=true", i+1)
			}
		} else {
			if ack.lastNackRequeue {
				t.Error("final call: expected requeue=false when max retries exceeded")
			}
		}
	}

	// 破棄後にカウンターがクリアされていることを確認
	key := msgKey(newDelivery(body, nil))
	c.retryMu.Lock()
	count := c.retryCounts[key]
	c.retryMu.Unlock()
	if count != 0 {
		t.Errorf("expected retryCounts to be cleared after discard, got %d", count)
	}
}

// TestHandleMessage_SuccessAfterRetry はリトライ後の成功でカウンターがクリアされることを確認する
func TestHandleMessage_SuccessAfterRetry(t *testing.T) {
	handler := &mockHandler{err: errors.New("temporary error")}
	c := newTestConsumer(3, handler)
	body := []byte("hello")

	// 1回失敗
	c.handleMessage(context.Background(), newDelivery(body, &mockAcknowledger{}))

	// 成功に切り替え
	handler.mu.Lock()
	handler.err = nil
	handler.mu.Unlock()

	ack := &mockAcknowledger{}
	c.handleMessage(context.Background(), newDelivery(body, ack))

	if ack.ackCount != 1 {
		t.Errorf("expected ackCount=1 after success, got %d", ack.ackCount)
	}

	// 成功後にカウンターがクリアされていることを確認
	key := msgKey(newDelivery(body, nil))
	c.retryMu.Lock()
	count := c.retryCounts[key]
	c.retryMu.Unlock()
	if count != 0 {
		t.Errorf("expected retryCounts to be cleared after success, got %d", count)
	}
}

// TestHandleMessage_RaceCondition は複数 goroutine から同時に handleMessage を呼んでも
// データ競合が発生しないことを確認する（go test -race で検証）
func TestHandleMessage_RaceCondition(t *testing.T) {
	const goroutines = 50
	c := newTestConsumer(100, &mockHandler{err: errors.New("always fail")})

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		body := []byte{byte(i)}
		go func() {
			defer wg.Done()
			c.handleMessage(context.Background(), newDelivery(body, &mockAcknowledger{}))
		}()
	}
	wg.Wait()
}
