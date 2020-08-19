// Copyright 2015 Matthew Holt and The Caddy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"minGateway/util/log"
	"runtime/debug"
	"sync"
	"time"
)

// SessionTicketService configures and manages TLS session tickets.
type SessionTicketService struct {
	// How often Caddy rotates STEKs. Default: 12h.
	RotationInterval time.Duration `json:"rotation_interval,omitempty"`

	// The maximum number of keys to keep in rotation. Default: 4.
	MaxKeys int `json:"max_keys,omitempty"`

	// Disables STEK rotation.
	DisableRotation bool `json:"disable_rotation,omitempty"`

	// Disables TLS session resumption by tickets.
	Disabled bool `json:"disabled,omitempty"`

	ticker *time.Ticker

	configs     map[*tls.Config]struct{}
	stopChan    chan struct{}
	currentKeys [][32]byte
	mu          *sync.Mutex
}

func (s *SessionTicketService) Run() error {
	s.configs = make(map[*tls.Config]struct{})
	s.mu = new(sync.Mutex)

	// establish sane defaults
	if s.RotationInterval == 0 {
		s.RotationInterval = defaultSTEKRotationInterval
	}

	if s.MaxKeys <= 0 {
		s.MaxKeys = defaultMaxSTEKs
	}

	// start the STEK module; this ensures we have
	// a starting key before any config needs one
	return s.start()
}

// start loads the starting STEKs and spawns a goroutine
// which loops to rotate the STEKs, which continues until
// stop() is called. If start() was already called, this
// is a no-op.
func (s *SessionTicketService) start() error {
	if s.stopChan != nil {
		return nil
	}
	s.stopChan = make(chan struct{})

	// initializing the key source gives us our
	// initial key(s) to start with; if successful,
	// we need to be sure to call Next() so that
	// the key source can know when it is done
	initialKey, err := s.generateSTEK()
	if err != nil {
		return fmt.Errorf("setting STEK module configuration: %v", err)
	}

	s.mu.Lock()
	s.currentKeys = [][32]byte{initialKey}
	s.mu.Unlock()

	// keep the keys rotated
	go s.stayUpdated()

	return nil
}

// stayUpdated is a blocking function which rotates
// the keys whenever new ones are sent. It reads
// from keysChan until s.stop() is called.
func (s *SessionTicketService) stayUpdated() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[PANIC] session ticket service: %v\n%s", err, debug.Stack())
		}
	}()

	s.ticker = time.NewTicker(s.RotationInterval)
	defer func() {
		_ = s.ticker.Stop
	}()

	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()

			if s.DisableRotation {
				s.mu.Unlock()
				continue
			}

			newKeys, err := s.RotateSTEKs(s.currentKeys)
			if err == nil {
				s.currentKeys = newKeys
			}
			fmt.Println(newKeys)

			configs := s.configs

			s.mu.Unlock()

			for cfg := range configs {
				cfg.SetSessionTicketKeys(newKeys)
			}
		case <-s.stopChan:
			return
		}
	}
}

// stop terminates the key rotation goroutine.
func (s *SessionTicketService) Stop() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}

// register sets the session ticket keys on cfg
// and keeps them updated. Any values registered
// must be unregistered, or they will not be
// garbage-collected. s.start() must have been
// called first. If session tickets are disabled
// or if ticket key rotation is disabled, this
// function is a no-op.
func (s *SessionTicketService) Register(cfg *tls.Config) {
	s.mu.Lock()
	cfg.SetSessionTicketKeys(s.currentKeys)
	s.configs[cfg] = struct{}{}
	s.mu.Unlock()
}

// unregister stops session key management on cfg and
// removes the internal stored reference to cfg. If
// session tickets are disabled or if ticket key rotation
// is disabled, this function is a no-op.
func (s *SessionTicketService) Unregister(cfg *tls.Config) {
	s.mu.Lock()
	delete(s.configs, cfg)
	s.mu.Unlock()
}

// RotateSTEKs rotates the keys in keys by producing a new key and eliding
// the oldest one. The new slice of keys is returned.
func (s SessionTicketService) RotateSTEKs(keys [][32]byte) ([][32]byte, error) {
	// produce a new key
	newKey, err := s.generateSTEK()
	if err != nil {
		return nil, fmt.Errorf("generating STEK: %v", err)
	}

	// we need to prepend this new key to the list of
	// keys so that it is preferred, but we need to be
	// careful that we do not grow the slice larger
	// than MaxKeys, otherwise we'll be storing one
	// more key in memory than we expect; so be sure
	// that the slice does not grow beyond the limit
	// even for a brief period of time, since there's
	// no guarantee when that extra allocation will
	// be overwritten; this is why we first trim the
	// length to one less the max, THEN prepend the
	// new key
	if len(keys) >= s.MaxKeys {
		keys[len(keys)-1] = [32]byte{} // zero-out memory of oldest key
		keys = keys[:s.MaxKeys-1]      // trim length of slice
	}
	keys = append([][32]byte{newKey}, keys...) // prepend new key

	return keys, nil
}

// generateSTEK generates key material suitable for use as a
// session ticket ephemeral key.
func (s *SessionTicketService) generateSTEK() ([32]byte, error) {
	var newTicketKey [32]byte
	_, err := io.ReadFull(rand.Reader, newTicketKey[:])
	return newTicketKey, err
}

const (
	defaultSTEKRotationInterval = 12 * time.Hour
	defaultMaxSTEKs             = 4
)
