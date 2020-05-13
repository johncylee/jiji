package jiji

import (
	"encoding/json"
	"testing"
	"time"
)

const (
	DBPath = "testdb"
)

type MockNetError struct {
	IsTimeout   bool
	IsTemporary bool
}

func init() {
	Verbose = true
}

func (t *MockNetError) Error() string {
	return "MockNetError"
}

func (t *MockNetError) Timeout() bool {
	return t.IsTimeout
}

func (t *MockNetError) Temporary() bool {
	return t.IsTemporary
}

type MockTransport struct {
	Available bool
	connected bool
	last      int64
	T         *testing.T
}

func (t *MockTransport) Connect() error {
	if t.Available {
		t.connected = true
		return nil
	}
	return &MockNetError{
		IsTimeout:   true,
		IsTemporary: true,
	}
}

func (t *MockTransport) Close() {
	t.connected = false
	return
}

func (t *MockTransport) Send(buf []byte) error {
	if !t.Available {
		return &MockNetError{
			IsTimeout:   true,
			IsTemporary: true,
		}
	}
	Logger.Println("MockTransport.Send:", string(buf))
	var i int64
	err := json.Unmarshal(buf, &i)
	if err != nil {
		return err
	}
	if i < t.last {
		t.T.Error("Incorrect order:", i)
	}
	t.last = i
	return nil
}

func TestDelivery(t *testing.T) {
	transport := MockTransport{
		Available: true,
		T:         t,
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		Logger.Println("MockTransport unavailable")
		transport.Available = false
		time.Sleep(400 * time.Millisecond)
		Logger.Println("MockTransport available")
		transport.Available = true
	}()
	testDelivery(&transport, t)
}

func testDelivery(transport Transport, t *testing.T) {
	Retry = 500 * time.Millisecond
	delivery := Delivery{
		DBPath:    DBPath,
		Send:      make(chan interface{}, 10),
		Transport: transport,
	}
	go func() {
		for i := 1; i < 10; i++ {
			unix := time.Now().Unix()
			Logger.Println("client deliver:", unix)
			delivery.Send <- unix
			time.Sleep(100 * time.Millisecond)
		}
		delivery.Close()
	}()
	err := delivery.Run()
	if err != nil {
		t.Error(err)
	}
}

func TestCloseSend(t *testing.T) {
	transport := MockTransport{
		Available: true,
		T:         t,
	}
	delivery := Delivery{
		DBPath:    DBPath,
		Send:      make(chan interface{}),
		Transport: &transport,
	}
	go func() {
		time.Sleep(time.Second)
		close(delivery.Send)
		delivery.Close()
	}()
	err := delivery.Run()
	if err != nil {
		t.Error(err)
	}
}
