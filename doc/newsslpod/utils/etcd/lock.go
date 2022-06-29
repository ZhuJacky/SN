// Package etcd provides ...
package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// Lock concurrency lock
func (cli *MyEtcd) Lock(ctx context.Context, key string) (*concurrency.Session, *concurrency.Mutex, error) {
	session, err := concurrency.NewSession(cli.etcd, concurrency.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}

	mu := concurrency.NewMutex(session, key)
	return session, mu, mu.Lock(ctx)
}

// Unlock concurrency unlock
func (cli *MyEtcd) Unlock(ctx context.Context, session *concurrency.Session, mu *concurrency.Mutex) error {
	err := mu.Unlock(ctx)
	if err != nil {
		return err
	}
	return session.Close()
}

// GetClient get clientv3
func (cli *MyEtcd) GetClient() *clientv3.Client {
	return cli.etcd
}
