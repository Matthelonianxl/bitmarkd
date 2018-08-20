// Copyright (c) 2014-2018 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package announce

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/bitmark-inc/bitmarkd/fault"
	"github.com/bitmark-inc/bitmarkd/mode"
	"github.com/bitmark-inc/bitmarkd/zmqutil"
)

type pubkey []byte

type peerEntry struct {
	publicKey []byte
	listeners []byte
	timestamp time.Time
}

func (p peerEntry) String() string {
	return fmt.Sprintf("PK:%x@%x-%v", p.publicKey, p.listeners, p.timestamp)
}

// called by the peering initialisation to set up this node's
// announcement data
func SetPeer(publicKey []byte, listeners []byte) error {
	globalData.Lock()
	defer globalData.Unlock()

	if globalData.peerSet {
		return fault.ErrAlreadyInitialised
	}
	globalData.publicKey = publicKey
	globalData.listeners = listeners
	globalData.peerSet = true

	addPeer(publicKey, listeners, 0)

	globalData.thisNode, _ = globalData.peerTree.Search(pubkey(publicKey))

	determineConnections(globalData.log)

	return nil
}

// add a peer announcement to the in-memory tree
// returns:
//   true  if this was a new/updated entry
//   false if the update was within the limits (to prevent continuous relaying)
func AddPeer(publicKey []byte, listeners []byte, timestamp uint64) bool {
	globalData.Lock()
	rc := addPeer(publicKey, listeners, timestamp)
	globalData.Unlock()
	return rc
}

// internal add a peer announcement, hold lock before calling
func addPeer(publicKey []byte, listeners []byte, timestamp uint64) bool {

	// disallow future timestamps: require timestamp >= Now
	ts := time.Now()
	if timestamp != 0 && timestamp <= uint64(ts.Unix()) {
		ts = time.Unix(int64(timestamp), 0)
	}

	// ignore expired request
	if time.Since(ts) >= announceExpiry {
		return false
	}

	peer := &peerEntry{
		publicKey: publicKey,
		listeners: listeners,
		timestamp: ts,
	}

	if node, _ := globalData.peerTree.Search(pubkey(publicKey)); nil != node {
		peer := node.Value().(*peerEntry)

		if ts.Sub(peer.timestamp) < announceRebroadcast {
			return false
		}
	}

	// add or update the timestamp in the tree
	recordAdded := globalData.peerTree.Insert(pubkey(publicKey), peer)

	globalData.log.Debugf("added: %t  nodes in the peer tree: %d", recordAdded, globalData.peerTree.Count())

	// if adding this nodes data
	if bytes.Equal(globalData.publicKey, publicKey) {
		return false
	}

	if recordAdded {
		globalData.treeChanged = true
	}

	return true
}

// fetch the data for the next node in the ring for a given public key
func GetNext(publicKey []byte) ([]byte, []byte, time.Time, error) {
	globalData.Lock()
	defer globalData.Unlock()

	node, _ := globalData.peerTree.Search(pubkey(publicKey))
	if nil != node {
		node = node.Next()
	}
	if nil == node {
		node = globalData.peerTree.First()
	}
	if nil == node {
		return nil, nil, time.Now(), fault.ErrInvalidPublicKey
	}
	peer := node.Value().(*peerEntry)
	return peer.publicKey, peer.listeners, peer.timestamp, nil
}

// send a peer registration request to a client channel
func SendRegistration(client *zmqutil.Client, fn string) error {
	chain := mode.ChainName()

	// get a big endian timestamp
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().Unix()))

	return client.Send(fn, chain, globalData.publicKey, globalData.listeners, timestamp)
}

// public key comparison for AVL interface
func (p pubkey) Compare(q interface{}) int {
	return bytes.Compare(p, q.(pubkey))
}

// public key string convert for AVL interface
func (p pubkey) String() string {
	return fmt.Sprintf("%x", []byte(p))
}
