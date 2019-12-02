package main

import (
	"context"
	"log"

	"go.etcd.io/etcd/clientv3"
)

const keyPrefix = "/my-data/"

func watchExpiredLease(cli *clientv3.Client) <-chan *clientv3.Event {
	events := make(chan *clientv3.Event)

	wCh := cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithFilterPut())

	go func() {
		defer close(events)
		for wResp := range wCh {
			for _, ev := range wResp.Events {
				expired, err := isExpired(cli, ev)
				if err != nil {
					log.Println("Error when checking expiry")
				} else if expired {
					events <- ev
				}
			}
		}
	}()

	return events
}

// isExpired decides if a DELETE event happended because of a lease expiry
func isExpired(cli *clientv3.Client, ev *clientv3.Event) (bool, error) {
	if ev.PrevKv == nil {
		return false, nil
	}

	leaseID := clientv3.LeaseID(ev.PrevKv.Lease)
	if leaseID == clientv3.NoLease {
		return false, nil
	}

	ttlResponse, err := cli.TimeToLive(context.Background(), leaseID)
	if err != nil {
		return false, err
	}

	return ttlResponse.TTL == -1, nil
}