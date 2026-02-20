package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func newTestWS(t *testing.T) (*websocket.Conn, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		// keep connection open
		select {}
	}))

	wsURL := "ws" + server.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Close()
	}

	return conn, cleanup
}

func TestCloseCon_Idempotent(t *testing.T) {
	conn, cleanup := newTestWS(t)
	defer cleanup()

	client := NewClient(uuid.New(), conn, nil)

	// Call CloseCon concurrently multiple times
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.CloseCon()
		}()
	}
	wg.Wait()

	select {
	case <-client.Close:

	default:
		t.Fatal("Close channel is not closed")
	}
}

func TestWriteLoop_StopsOnClose(t *testing.T) {
	conn, cleanup := newTestWS(t)
	defer cleanup()

	client := NewClient(uuid.New(), conn, nil)

	done := make(chan struct{})
	go func() {
		client.writeLoop()
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	client.CloseCon()

	select {
	case <-done:

	case <-time.After(2 * time.Second):
		t.Fatal("writeLoop did not exit after CloseCon")
	}
}

func TestWriteMessage(t *testing.T) {
	conn, cleanup := newTestWS(t)
	defer cleanup()

	client := NewClient(uuid.New(), conn, nil)

	go client.writeLoop()

	client.Msgch <- []byte("hello")

	time.Sleep(100 * time.Millisecond)

	client.CloseCon()
}

func BenchmarkClient_WriteLoop(b *testing.B) {
	conn, cleanup := newTestWS(&testing.T{})
	defer cleanup()

	client := NewClient(uuid.New(), conn, nil)

	go client.writeLoop()
	defer client.CloseCon()

	time.Sleep(50 * time.Millisecond)

	payload := []byte("benchmark-message")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		client.Msgch <- payload
	}
}
