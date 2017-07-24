// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package payment

import (
	"sync"
	"time"

	"github.com/bitmark-inc/bitmarkd/constants"
	"github.com/bitmark-inc/bitmarkd/fault"
	"github.com/bitmark-inc/bitmarkd/zmqutil"
	"github.com/bitmark-inc/logger"
	zmq "github.com/pebbe/zmq4"
)

const (
	discovererStopSignal = "inproc://discoverer-stop-signal"

	blockchainCheckIntervel = 60 * time.Second
)

// discoverer listens to discovery proxy to get the possible txs
type discoverer struct {
	log  *logger.L
	push *zmq.Socket
	pull *zmq.Socket
	sub  *zmq.Socket
	req  *zmq.Socket
}

func newDiscoverer(log *logger.L, subAddr, reqAddr string) (*discoverer, error) {
	push, pull, err := zmqutil.NewSignalPair(discovererStopSignal)
	if err != nil {
		return nil, fault.ErrNoConnectionsAvailable
	}

	sub, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, fault.ErrNoConnectionsAvailable
	}
	sub.Connect(subAddr)
	sub.SetSubscribe("")

	req, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fault.ErrNoConnectionsAvailable
	}
	req.Connect(reqAddr)

	return &discoverer{log, push, pull, sub, req}, nil
}

func (d *discoverer) Run(args interface{}, shutdown <-chan struct{}) {
	d.retrievePastTxs()

	go func() {
		poller := zmq.NewPoller()
		poller.Add(d.sub, zmq.POLLIN)
		poller.Add(d.pull, zmq.POLLIN)

	loop:
		for {
			polled, _ := poller.Poll(-1)

			// TODO: add hearbeat
			for _, p := range polled {
				switch s := p.Socket; s {
				case d.pull:
					if _, err := s.RecvMessageBytes(0); err != nil {
						d.log.Errorf("pull receive error: %v", err)
						break loop
					}
					break loop

				default:
					msg, err := s.RecvMessageBytes(0)
					if err != nil {
						d.log.Errorf("sub receive error: %v", err)
					}

					d.assignHandler(msg)
				}
			}
		}

		d.pull.Close()
		d.sub.Close()

		d.log.Info("stopped")
	}()

	d.log.Info("started")

	<-shutdown

	d.push.SendMessage("stop")
	d.push.Close()
	d.req.Close()
}

func (d *discoverer) retrievePastTxs() {
	originTime := time.Now().Add(-constants.ReservoirTimeout)

	for currency, handler := range globalData.handlers {
		d.log.Infof("start to fetch possible %s txs since time at %d", currency, originTime.Unix())

		d.req.SendMessage(currency, originTime.Unix())
		msg, err := d.req.RecvMessageBytes(0)
		if err != nil {
			d.log.Errorf("failed to receive message: %v", err)
		}

		handler.processPastTxs(msg[1])
	}
}

func (d *discoverer) assignHandler(data [][]byte) {
	if len(data) != 2 {
		d.log.Errorf("invalid message: %v", data)
		return
	}

	currency := string(data[0])
	globalData.handlers[currency].processIncomingTx(data[1])
}

// checker periodically extracts possible txs in the latest block
type checker struct {
}

func (c *checker) Run(args interface{}, shutdown <-chan struct{}) {
	for {
		select {
		case <-shutdown:
			break

		case <-time.After(blockchainCheckIntervel):
			var wg sync.WaitGroup
			for _, handler := range globalData.handlers {
				wg.Add(1)
				go handler.checkLatestBlock(&wg)
			}
			wg.Wait()
			globalData.log.Info("block check finished")
		}
	}
}