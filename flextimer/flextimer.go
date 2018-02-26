// Copyright (c) 2018 Zededa, Inc.
// All rights reserved.

// Provide randomized timers - botn based on range and binary exponential
// backoff.
// Usage:
//  ticker := NewRangeTicker(min, max)
//  select ticker.C
//  ticker.UpdateRangeTicker(newmin, newmix)
//  ticker.StopTicker()
// Usage:
//  ticker := NewExpTicker(start, max, randomFactor)
//  select ticker.C
//  ticker.UpdateRangeTicker(newstart, newmax, newRandomFactor)
//  ticker.StopTicker()

package flextimer

import (
	// "fmt"
	"math/rand"
	"time"
)

// Take min, max, exp bool
// If exp false then [min, max] is random range
// If exp true then start at min and do binary exponential backoff
// until hitting max, then stay at max. Randomize +/- randomFactor
// When config is all zeros, then stop and close channel

// XXX test that it can handle the TCP timeout and space out the next timers
// based on processing time ...

// Ticker handle for caller
type flexTickerHandle struct {
	C          <-chan time.Time
	configChan chan<- flexTickerConfig
}

// Arguments fed over configChan
type flexTickerConfig struct {
	exponential  bool
	minTime      time.Duration
	maxTime      time.Duration
	randomFactor float64
}

func NewRangeTicker(minTime time.Duration, maxTime time.Duration) flexTickerHandle {
	initialConfig := flexTickerConfig{minTime: minTime,
		maxTime: maxTime}
	configChan := make(chan flexTickerConfig, 1)
	tickChan := newFlexTicker(configChan)
	configChan <- initialConfig
	return flexTickerHandle{C: tickChan, configChan: configChan}
}

func NewExpTicker(minTime time.Duration, maxTime time.Duration, randomFactor float64) flexTickerHandle {
	initialConfig := flexTickerConfig{minTime: minTime,
		maxTime: maxTime, exponential: true,
		randomFactor: randomFactor}
	configChan := make(chan flexTickerConfig, 1)
	tickChan := newFlexTicker(configChan)
	configChan <- initialConfig
	return flexTickerHandle{C: tickChan, configChan: configChan}
}

func (f flexTickerHandle) UpdateRangeTicker(minTime time.Duration, maxTime time.Duration) {
	config := flexTickerConfig{minTime: minTime,
		maxTime: maxTime}
	f.configChan <- config
}

func (f flexTickerHandle) UpdateExpTicker(minTime time.Duration, maxTime time.Duration, randomFactor float64) {
	config := flexTickerConfig{minTime: minTime,
		maxTime: maxTime, exponential: true,
		randomFactor: randomFactor}
	f.configChan <- config
}

func (f flexTickerHandle) StopTicker() {
	f.configChan <- flexTickerConfig{}
}

// Implementation functions

func newFlexTicker(config <-chan flexTickerConfig) <-chan time.Time {
	tick := make(chan time.Time, 1)
	go flexTicker(config, tick)
	return tick
}

func flexTicker(config <-chan flexTickerConfig, tick chan<- time.Time) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	// fmt.Printf("flexTicker: waiting for intial config\n")
	// Wait for initial config
	c := <-config
	// fmt.Printf("flexTicker: got intial config %v\n", c)
	expFactor := 1
	for {
		var d time.Duration
		if c.exponential {
			rf := c.randomFactor
			if rf == 0 {
				rf = 1.0
			} else if rf > 1.0 {
				rf = 1.0 / rf
			}
			min := float64(c.minTime) * float64(expFactor) * rf
			max := float64(c.minTime) * float64(expFactor) / rf
			base := float64(c.minTime) * float64(expFactor)
			if time.Duration(base) < c.maxTime {
				expFactor *= 2
			}
			if max == min {
				d = time.Duration(min)
			} else {
				// fmt.Printf("base %v range %v\n", int64(min), int64(max - min))
				r := r1.Int63n(int64(max-min)) + int64(min)
				d = time.Duration(r)
			}
			// fmt.Printf("Exponential %v %d secs\n", d, d/time.Second)
		} else {
			r := r1.Int63n(int64(c.maxTime-c.minTime)) + int64(c.minTime)
			d = time.Duration(r)
			// fmt.Printf("Random %v %d secs\n", d, d/time.Second)
		}
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
			// fmt.Printf("flexTicker: timer fired\n")
			tick <- time.Now()
		case c = <-config:
			// Replace current parameters without
			// looking at when current timer would fire
			// fmt.Printf("flexTicker: got updated config %v\n", c)
			timer.Stop()
			expFactor = 1
			if c.maxTime == 0 && c.minTime == 0 {
				// fmt.Printf("flexTicker: got stop\n")
				close(tick)
				return
			}
		}
	}
}
