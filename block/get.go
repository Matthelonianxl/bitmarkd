// Copyright (c) 2014-2018 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package block

import (
	"encoding/binary"

	"github.com/bitmark-inc/bitmarkd/blockdigest"
	"github.com/bitmark-inc/bitmarkd/blockrecord"
	"github.com/bitmark-inc/bitmarkd/blockring"
	"github.com/bitmark-inc/bitmarkd/fault"
	"github.com/bitmark-inc/bitmarkd/genesis"
	"github.com/bitmark-inc/bitmarkd/mode"
	"github.com/bitmark-inc/bitmarkd/storage"
)

// get block data for initialising a new block
// returns: previous block digest and the number for the new block
func Get() (blockdigest.Digest, uint64) {
	globalData.Lock()
	defer globalData.Unlock()
	nextBlockNumber := globalData.height + 1
	return globalData.previousBlock, nextBlockNumber
}

// get the current height
func GetHeight() uint64 {
	globalData.Lock()
	height := globalData.height
	globalData.Unlock()
	return height
}

func DigestForBlock(number uint64) (blockdigest.Digest, error) {
	globalData.Lock()
	defer globalData.Unlock()

	// valid block number
	if number <= genesis.BlockNumber {
		if mode.IsTesting() {
			return genesis.TestGenesisDigest, nil
		}
		return genesis.LiveGenesisDigest, nil
	}

	// check if in the cache
	if number > genesis.BlockNumber && number <= globalData.height {
		d := blockring.DigestForBlock(number)
		if nil != d {
			return *d, nil
		}
	}

	// no cache, fetch block and compute digest
	n := make([]byte, 8)
	binary.BigEndian.PutUint64(n, number)
	packed := storage.Pool.Blocks.Get(n) // ***** FIX THIS: possible optimisation is to store the block hashes in a separate index
	if nil == packed {
		return blockdigest.Digest{}, fault.ErrBlockNotFound
	}

	_, digest, _, err := blockrecord.ExtractHeader(packed)

	return digest, err
}
