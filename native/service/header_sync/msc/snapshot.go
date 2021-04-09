/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */
package msc

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

// Tally is a simple vote tally to keep the current score of votes. Votes that
// go against the proposal aren't counted since it's equivalent to not voting.
type Tally struct {
	Authorize bool `json:"authorize"` // Whether the vote is about authorizing or kicking someone
	Votes     int  `json:"votes"`     // Number of votes until now wanting to pass the proposal
}

// Vote represents a single vote that an authorized signer made to modify the
// list of authorizations.
type Vote struct {
	Signer    common.Address `json:"signer"`    // Authorized signer that cast this vote
	Block     uint64         `json:"block"`     // Block number the vote was cast in (expire old votes)
	Address   common.Address `json:"address"`   // Account being voted on to change its authorization
	Authorize bool           `json:"authorize"` // Whether to authorize or deauthorize the voted account
}

// Snapshot ...
type Snapshot struct {
	Number  uint64                      `json:"number"`  // Block number where the snapshot was created
	Hash    common.Hash                 `json:"hash"`    // Block hash where the snapshot was created
	Signers map[common.Address]struct{} `json:"signers"` // Set of authorized signers at this moment
	Votes   []*Vote                     `json:"votes"`   // List of votes cast in chronological order
	Tally   map[common.Address]Tally    `json:"tally"`   // Current vote tally to avoid recalculating
	Ctx     *Context
}

func newSnapshot(number uint64, hash common.Hash, signers []common.Address, ctx *Context) *Snapshot {
	snap := &Snapshot{
		Number:  number,
		Hash:    hash,
		Signers: make(map[common.Address]struct{}),
		Tally:   make(map[common.Address]Tally),
		Ctx:     ctx,
	}
	for _, signer := range signers {
		snap.Signers[signer] = struct{}{}
	}
	return snap
}

// validVote returns whether it makes sense to cast the specified vote in the
// given snapshot context (e.g. don't try to add an already authorized signer).
func (s *Snapshot) validVote(address common.Address, authorize bool) bool {
	_, signer := s.Signers[address]
	return (signer && !authorize) || (!signer && authorize)
}

// cast adds a new vote into the tally.
func (s *Snapshot) cast(address common.Address, authorize bool) bool {
	// Ensure the vote is meaningful
	if !s.validVote(address, authorize) {
		return false
	}
	// Cast the vote into an existing or new tally
	if old, ok := s.Tally[address]; ok {
		old.Votes++
		s.Tally[address] = old
	} else {
		s.Tally[address] = Tally{Authorize: authorize, Votes: 1}
	}
	return true
}

// uncast removes a previously cast vote from the tally.
func (s *Snapshot) uncast(address common.Address, authorize bool) bool {
	// If there's no tally, it's a dangling vote, just drop
	tally, ok := s.Tally[address]
	if !ok {
		return false
	}
	// Ensure we only revert counted votes
	if tally.Authorize != authorize {
		return false
	}
	// Otherwise revert the vote
	if tally.Votes > 1 {
		tally.Votes--
		s.Tally[address] = tally
	} else {
		delete(s.Tally, address)
	}
	return true
}

// apply creates a new authorization snapshot by applying the given headers to
// the original one.
func (s *Snapshot) apply(headers []*HeaderWithDifficultySum, targetSigner common.Address, lastSeenAddress *uint64) (err error) {
	// Allow passing in no headers for cleaner code
	if len(headers) == 0 {
		return
	}

	var signer common.Address
	for _, headerWS := range headers {
		header := headerWS.Header
		// Remove any votes on checkpoint blocks
		number := header.Number.Uint64()
		// Resolve the authorization key and check against signers
		signer, err = ecrecover(header)
		if err != nil {
			err = fmt.Errorf("ecrecover err %v", err)
			return
		}
		if targetSigner == signer {
			*lastSeenAddress = number
		}
		if _, ok := s.Signers[signer]; !ok {
			err = fmt.Errorf("unauthorized signer for block %d", number)
			return
		}

		// Header authorized, discard any previous votes from the signer
		for i, vote := range s.Votes {
			if vote.Signer == signer && vote.Address == header.Coinbase {
				// Uncast the vote from the cached tally
				s.uncast(vote.Address, vote.Authorize)

				// Uncast the vote from the chronological list
				s.Votes = append(s.Votes[:i], s.Votes[i+1:]...)
				break // only one vote allowed
			}
		}
		// Tally up the new vote from the signer
		var authorize bool
		switch {
		case bytes.Equal(header.Nonce[:], nonceAuthVote):
			authorize = true
		case bytes.Equal(header.Nonce[:], nonceDropVote):
			authorize = false
		default:
			err = errInvalidVote
			return
		}
		if s.cast(header.Coinbase, authorize) {
			s.Votes = append(s.Votes, &Vote{
				Signer:    signer,
				Block:     number,
				Address:   header.Coinbase,
				Authorize: authorize,
			})
		}
		// If the vote passed, update the list of signers
		if tally := s.Tally[header.Coinbase]; tally.Votes > len(s.Signers)/2 {
			if tally.Authorize {
				s.Signers[header.Coinbase] = struct{}{}
			} else {
				delete(s.Signers, header.Coinbase)

				// Discard any previous votes the deauthorized signer cast
				for i := 0; i < len(s.Votes); i++ {
					if s.Votes[i].Signer == header.Coinbase {
						// Uncast the vote from the cached tally
						s.uncast(s.Votes[i].Address, s.Votes[i].Authorize)

						// Uncast the vote from the chronological list
						s.Votes = append(s.Votes[:i], s.Votes[i+1:]...)

						i--
					}
				}
			}
			// Discard any previous votes around the just changed account
			for i := 0; i < len(s.Votes); i++ {
				if s.Votes[i].Address == header.Coinbase {
					s.Votes = append(s.Votes[:i], s.Votes[i+1:]...)
					i--
				}
			}
			delete(s.Tally, header.Coinbase)
		}

	}

	s.Number += uint64(len(headers))
	s.Hash = headers[len(headers)-1].Header.Hash()

	return
}

// signersAscending implements the sort interface to allow sorting a list of addresses
type signersAscending []common.Address

func (s signersAscending) Len() int           { return len(s) }
func (s signersAscending) Less(i, j int) bool { return bytes.Compare(s[i][:], s[j][:]) < 0 }
func (s signersAscending) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// signers retrieves the list of authorized signers in ascending order.
func (s *Snapshot) signers() []common.Address {
	sigs := make([]common.Address, 0, len(s.Signers))
	for sig := range s.Signers {
		sigs = append(sigs, sig)
	}
	sort.Sort(signersAscending(sigs))
	return sigs
}

// inturn returns if a signer at a given block height is in-turn or not.
func (s *Snapshot) inturn(number uint64, signer common.Address, offsetPointer *int) bool {
	signers, offset := s.signers(), 0
	for offset < len(signers) && signers[offset] != signer {
		offset++
	}
	if offsetPointer != nil {
		*offsetPointer = offset
	}
	return (number % uint64(len(signers))) == uint64(offset)
}
