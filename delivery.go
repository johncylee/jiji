package jiji

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	bucket_SYNC = "sync"
)

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// retry interval
var Retry = 10 * time.Second

type Delivery struct {
	// DBPath is the path of bbolt db.
	DBPath string
	// Send is the entrance of delivery service. Send anything supported by json.Marshal
	// to it. Close it to stop Delivery.Run().
	Send <-chan any
	// Transport agent. Could be Print or HTTP.
	Transport  Transport
	db         *bolt.DB
	connecting sync.Mutex
	connected  chan struct{}
	ctx        context.Context
}

type Transport interface {
	Connect(ctx context.Context) error
	Close()
	Send(b []byte, ctx context.Context) error
}

func NewDelivery(dbpath string, send <-chan any, transport Transport) (d *Delivery) {
	d = &Delivery{
		DBPath:    dbpath,
		Send:      send,
		Transport: transport,
	}
	d.connected = make(chan struct{})
	return
}

func (t *Delivery) enqueue(b []byte) (err error) {
	var id uint64
	tx, err := t.db.Begin(true)
	if err != nil {
		return
	}
	bkt, err := tx.CreateBucketIfNotExists([]byte(bucket_SYNC))
	if err != nil {
		tx.Rollback()
		return
	}
	if id, err = bkt.NextSequence(); err != nil {
		tx.Rollback()
		return
	}
	Logger.Debug("Bucket.Put", "key", id, "value", string(b))
	if err = bkt.Put(itob(id), b); err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (t *Delivery) send(b []byte) error {
	Logger.Debug("send", "msg", string(b))
	ctx, cancel := context.WithTimeout(t.ctx, Retry)
	defer cancel()
	return t.Transport.Send(b, ctx)
}

func (t *Delivery) send_queue() (connected bool, err error) {
	Logger.Debug("send_queue")
	connected = true
	err = t.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_SYNC))
		if b == nil { // empty
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if e := t.send(v); e != nil {
				Logger.Warn("Transport.Send", "error", e)
				connected = false
				return nil // commit
			}
			if e := c.Delete(); e != nil {
				Logger.Error("Cursor.Delete", "error", e)
				return nil // at least once
			}
		}
		return nil
	})
	return
}

func (t *Delivery) connect() {
	defer t.connecting.Unlock()
	var err error
	for {
		t.Transport.Close()
		Logger.Info("connect", "timeout", Retry)
		timer := time.After(Retry)
		ctx, cancel := context.WithTimeout(t.ctx, Retry)
		if err = t.Transport.Connect(ctx); err == nil {
			Logger.Debug("connected")
			cancel()
			t.connected <- struct{}{}
			return
		}
		cancel()
		Logger.Warn("Transport.Connect", "error", err)
		select {
		case <-t.ctx.Done():
			Logger.Debug("connect canceled")
			return
		case <-timer:
		}
	}
}

// Run Delivery service. Close Delivery.Send channel to stop it.
func (t *Delivery) Run() (err error) {
	if t.db, err = bolt.Open(t.DBPath, 0600, nil); err != nil {
		return
	}
	defer t.db.Close()
	defer t.Transport.Close()
	defer t.connecting.Lock()
	var cancel context.CancelFunc
	t.ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	connected := false
	for {
		if !connected && t.connecting.TryLock() {
			Logger.Debug("call t.connect")
			go t.connect()
		}
		select {
		case msg, ok := <-t.Send:
			if !ok {
				Logger.Info("send channel closed, exit")
				return nil // channel closed, normal exit
			}
			var b []byte
			if b, err = json.Marshal(msg); err != nil {
				return
			}
			if connected {
				if err = t.send(b); err == nil {
					continue
				}
				connected = false
			}
			// not connected
			if err = t.enqueue(b); err != nil {
				return
			}
		case <-t.connected:
			if connected, err = t.send_queue(); err != nil {
				return
			}
		}
	}
}
