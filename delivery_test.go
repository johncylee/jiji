package jiji

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"
)

const (
	TESTDB = "testdb"
)

var (
	ConnectError error = errors.New("connect error")
	SendError    error = errors.New("send error")
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	Retry = 50 * time.Millisecond
}

type MockTransport struct {
	Available bool
	connected bool
	last      int
}

func (t *MockTransport) Connect(ctx context.Context) error {
	t.connected = t.Available
	if !t.connected {
		return ConnectError
	}
	return nil
}

func (t *MockTransport) Close() {
	t.connected = false
}

func (t *MockTransport) Send(buf []byte, ctx context.Context) error {
	t.connected = t.Available
	if !t.connected {
		return SendError
	}
	Logger.Info("MockTransport.Send", "msg", string(buf))
	var i int
	err := json.Unmarshal(buf, &i)
	if err != nil {
		return err
	}
	if i <= t.last {
		return fmt.Errorf("Incorrect order: %d", i)
	}
	t.last = i
	return nil
}

func TestDelivery(t *testing.T) {
	transport := &MockTransport{
		Available: true,
	}
	on := func() {
		Logger.Info("MockTransport becomes available")
		transport.Available = true
	}
	off := func() {
		Logger.Info("MockTransport becomes unavailable")
		transport.Available = false
	}
	var d time.Duration
	done := make(chan struct{})
	for i := range 8 {
		d += time.Duration(rand.Int64N(50)+50) * time.Millisecond
		switch {
		case i%2 == 0:
			time.AfterFunc(d, off)
		case i == 7:
			time.AfterFunc(d, func() {
				on()
				close(done)
			})
		default:
			time.AfterFunc(d, on)
		}
	}
	n := 10
	testDelivery(t, transport, n)
	if transport.last != n {
		t.Fatal("expect", n)
	}
	<-done
}

func testDelivery(t *testing.T, transport Transport, n int) {
	send := make(chan any, n)
	db := filepath.Join(t.TempDir(), TESTDB)
	delivery := NewDelivery(db, send, transport)
	done := make(chan error)
	go func() {
		done <- delivery.Run()
	}()
	for counter := range n {
		send <- counter + 1
		time.Sleep(100 * time.Millisecond)
	}
	close(send)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
