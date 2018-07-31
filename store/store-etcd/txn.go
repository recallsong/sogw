package etcd

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

// logSlowTxn wraps etcd transaction and log slow Commits.
type logSlowTxn struct {
	clientv3.Txn
	cancel context.CancelFunc
}

func (s *EtcdStore) txn() clientv3.Txn {
	ctx, cancel := context.WithTimeout(s.client.Ctx(), DefaultSlowTxnTimeToLog)
	return &logSlowTxn{
		Txn:    s.client.Txn(ctx),
		cancel: cancel,
	}
}

func (t *logSlowTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &logSlowTxn{
		Txn:    t.Txn.If(cs...),
		cancel: t.cancel,
	}
}

func (t *logSlowTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &logSlowTxn{
		Txn:    t.Txn.Then(ops...),
		cancel: t.cancel,
	}
}

func (t *logSlowTxn) Commit() (*clientv3.TxnResponse, error) {
	start := time.Now()
	resp, err := t.Txn.Commit()
	t.cancel()
	cost := time.Now().Sub(start)
	if err != nil {
		log.Errorf("[etcd] txn.Commit error : %s", err.Error())
		return nil, err
	}
	if cost > DefaultSlowTxnTimeToLog {
		log.Warnf("[etcd] slow txn, resp=<%v> cost=<%s> error:\n %+v", resp, cost, err)
	}
	return resp, err
}
