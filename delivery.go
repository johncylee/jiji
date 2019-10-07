package jiji

import (
	"encoding/binary"
	"encoding/json"
	"github.com/boltdb/bolt"
	"time"
)

const (
	SYNC_BUCKET = "sync"
)

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// retry interval
var Retry = 5 * time.Second

type Delivery struct {
	DBPath      string
	Recv        chan interface{}
	Transport   Transport
	db          *bolt.DB
	connected   bool
	reconnected chan bool
	quit        chan struct{}
}

type Transport interface {
	Connect() error
	Close()
	Send([]byte) error
}

func (t *Delivery) send_msg(msg interface{}) (err error) {
	var id uint64
	tx, err := t.db.Begin(true)
	bkt, err := tx.CreateBucketIfNotExists([]byte(SYNC_BUCKET))
	if err != nil {
		tx.Rollback()
		return
	}
	for msg != nil {
		buf, err := json.Marshal(msg)
		if err != nil {
			Logger.Println("json.Marshal:", err)
			// ignore marshal error
			goto NEXT
		}
		if t.connected {
			err = t.Transport.Send(buf)
			if err == nil {
				goto NEXT
			}
			Logger.Println("Transport.Send:", err)
			t.connected = false
			go t.reconnect()
		}
		id, err = bkt.NextSequence()
		if err != nil {
			break
		}
		debugln("send_msg.Put:", string(buf))
		err = bkt.Put(itob(id), buf)
		if err != nil {
			break
		}
	NEXT: // drain the channel
		select {
		case msg = <-t.Recv:
		default:
			msg = nil
		}
	}
	e := tx.Commit()
	if e != nil {
		err = e
	}
	return
}

func (t *Delivery) send_queue() (err error) {
	err = t.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SYNC_BUCKET))
		if b == nil { // empty
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			debugln("send_queue:", string(v))
			err = t.Transport.Send(v)
			if err != nil {
				Logger.Println("Transport.Send:", err)
				t.connected = false
				go t.reconnect()
				return nil // commit
			}
			err = c.Delete()
			if err != nil {
				return nil // at least once
			}
		}
		return nil
	})
	return
}

func (t *Delivery) reconnect() {
	debugln("reconnect")
	t.Transport.Close()
	for {
		Logger.Printf("Reconnect in %.2f seconds...",
			float64(Retry)/float64(time.Second))
		time.Sleep(Retry)
		err := t.Transport.Connect()
		if err == nil {
			break
		}
	}
	t.reconnected <- true
	return
}

func (t *Delivery) close() {
	defer close(t.quit)
	if t.connected {
		t.Transport.Close()
	}
	defer t.db.Close()
	err := t.db.Update(func(tx *bolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists([]byte(SYNC_BUCKET))
		if err != nil {
			return
		}
	FOR:
		for {
			select {
			case msg, ok := <-t.Recv:
				if !ok {
					break FOR
				}
				debugln("Close.Recv:", msg)
				buf, err := json.Marshal(msg)
				if err != nil {
					Logger.Println("json.Marshal:", err)
					continue
				}
				id, err := b.NextSequence()
				if err != nil {
					break FOR
				}
				err = b.Put(itob(id), buf)
				if err != nil {
					break FOR
				}
			default:
				break FOR
			}
		}
		return
	})
	if err != nil {
		Logger.Println("dump failed, data lost:", err)
	}
	return
}

func (t *Delivery) Close() {
	if t.quit == nil {
		return
	}
	t.quit <- struct{}{}
	<-t.quit
}

func (t *Delivery) Run() (err error) {
	t.db, err = bolt.Open(t.DBPath, 0600, nil)
	if err != nil {
		return
	}
	t.quit = make(chan struct{})
	defer t.close()
	err = t.Transport.Connect()
	if err != nil {
		return
	}
	t.connected = true
	t.reconnected = make(chan bool)
FOR:
	for {
		if t.connected {
			err = t.send_queue()
			if err != nil {
				break FOR
			}
		}
		select {
		case msg := <-t.Recv:
			err = t.send_msg(msg)
			if err != nil {
				break FOR
			}
		case t.connected = <-t.reconnected:
			Logger.Println("Reconnected")
		case <-t.quit:
			debugln("Quitting Delivery.Run()")
			break FOR
		}
	}
	return
}
