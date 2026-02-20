package main

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestClient(t *testing.T, alias string) (*Client, func()) {
	t.Helper()

	conn, cleanup := newTestWS(t)
	cli := NewClient(uuid.New(), conn)
	cli.SetInfo(ClientInfoWithoutId{Alias: alias})

	return cli, cleanup
}

func closeCore(c *Core) {
	c.closedOnce.Do(func() {
		close(c.closed)
	})
}

func TestCoreRegisterReturnsExistingPeersAndClients(t *testing.T) {
	core := NewCore()
	defer closeCore(core)

	first, cleanupFirst := newTestClient(t, "first")
	defer cleanupFirst()

	second, cleanupSecond := newTestClient(t, "second")
	defer cleanupSecond()

	firstResult, err := core.Register(first)
	if err != nil {
		t.Fatalf("register first client: %v", err)
	}
	if len(firstResult.Peers) != 0 || len(firstResult.Existing) != 0 {
		t.Fatalf("first register expected no peers/existing, got peers=%d existing=%d", len(firstResult.Peers), len(firstResult.Existing))
	}

	secondResult, err := core.Register(second)
	if err != nil {
		t.Fatalf("register second client: %v", err)
	}

	if len(secondResult.Peers) != 1 {
		t.Fatalf("expected 1 peer for second register, got %d", len(secondResult.Peers))
	}
	if secondResult.Peers[0].Id != first.ClientId {
		t.Fatalf("expected peer id %s, got %s", first.ClientId, secondResult.Peers[0].Id)
	}
	if secondResult.Peers[0].Alias != "first" {
		t.Fatalf("expected peer alias first, got %q", secondResult.Peers[0].Alias)
	}

	if len(secondResult.Existing) != 1 {
		t.Fatalf("expected 1 existing client for second register, got %d", len(secondResult.Existing))
	}
	if secondResult.Existing[0] != first {
		t.Fatal("expected existing client pointer to match first client")
	}
}

func TestCoreUnregisterRemovesAndClosesClient(t *testing.T) {
	core := NewCore()
	defer closeCore(core)

	cli, cleanup := newTestClient(t, "remove-me")
	defer cleanup()

	if _, err := core.Register(cli); err != nil {
		t.Fatalf("register client: %v", err)
	}

	if err := core.Unregister(cli.ClientId); err != nil {
		t.Fatalf("unregister client: %v", err)
	}

	select {
	case <-cli.Close:
	case <-time.After(2 * time.Second):
		t.Fatal("client Close channel not closed after unregister")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, ok := core.clients[cli.ClientId]; !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("client still present in core after unregister")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCoreEnqueueReturnsMailboxFullWithoutReceiver(t *testing.T) {
	core := &Core{
		MessChan: make(chan coreMessage),
		clients:  map[uuid.UUID]*Client{},
		closed:   make(chan struct{}),
	}

	err := core.Enqueue(unregisterMsg{ClientID: uuid.New()})
	if !errors.Is(err, ErrMailboxFull) {
		t.Fatalf("expected ErrMailboxFull, got %v", err)
	}
}

func TestCoreRegisterReturnsCoreClosed(t *testing.T) {
	core := NewCore()
	closeCore(core)

	cli, cleanup := newTestClient(t, "closed")
	defer cleanup()

	_, err := core.Register(cli)
	if !errors.Is(err, ErrCoreClosed) {
		t.Fatalf("expected ErrCoreClosed, got %v", err)
	}
}

func BenchmarkCoreHandleRegister(b *testing.B) {
	const existingClients = 1000

	core := &Core{
		clients: map[uuid.UUID]*Client{},
	}

	for i := 0; i < existingClients; i++ {
		id := uuid.New()
		cli := &Client{ClientId: id}
		cli.SetInfo(ClientInfoWithoutId{Alias: "peer"})
		core.clients[id] = cli
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := uuid.New()
		candidate := &Client{ClientId: id}
		candidate.SetInfo(ClientInfoWithoutId{Alias: "bench"})

		resp := make(chan RegisterResult, 1)
		core.handleRegister(registerMsg{Client: candidate, Response: resp})
		<-resp

		delete(core.clients, id)
	}
}
