package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func TestWatchExpiredEventsReceiveExpiry(t *testing.T) {
	key := keyPrefix + "my-test-key"
	value := "my value that expired"

	endpoint, close := startETCDServer(t)
	defer close()

	cfg := clientv3.Config{Endpoints: []string{endpoint}}
	cli, err := clientv3.New(cfg)
	assert.Nil(t, err)
	defer cli.Close()

	go func() {
		lease, err := cli.Grant(context.Background(), 1)
		assert.Nil(t, err)
		_, err = cli.Put(context.Background(), key, value, clientv3.WithLease(lease.ID))
		assert.Nil(t, err)
	}()

	select {
	case ev := <-watchExpiredLease(cli):
		log.Printf("Expiry event for key: '%s' and value: '%s'\n", ev.PrevKv.Key, ev.PrevKv.Value)
		assert.Equal(t, key, string(ev.PrevKv.Key))
		assert.Equal(t, value, string(ev.PrevKv.Value))
	case <-time.After(5 * time.Second):
		t.Fatalf("Expiry event did not fire")
	}
}

// Based on: https://github.com/etcd-io/etcd/blob/v3.5.4/tests/integration/embed/embed_test.go#L74
func startETCDServer(t *testing.T) (endpoint string, close func()) {
	cfg := embed.NewConfig()
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{"/dev/null"}
	cfg.Dir = filepath.Join(os.TempDir(), fmt.Sprint(time.Now().Nanosecond()))

	srv, err := embed.StartEtcd(cfg)
	assert.Nil(t, err)

	select {
	case <-srv.Server.ReadyNotify():
	case <-time.After(3 * time.Second):
		t.Fatalf("Failed to start embed.Etcd for tests")
	}

	return cfg.ACUrls[0].String(), func() {
		os.RemoveAll(cfg.Dir)
		srv.Close()
	}
}
