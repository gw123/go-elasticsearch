package estransport

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"sync"
	"time"
)

// ConnectionPool defines the interface for the connection pool.
//
type ConnectionPool interface {
	Next() (*Connection, error)
	Remove(*Connection) error
}

// Connection represents a connection to a node.
//
type Connection struct {
	sync.Mutex

	URL       *url.URL  `json:"url"`
	Dead      bool      `json:"dead"`
	DeadSince time.Time `json:"dead_since"`
	Failures  int       `json:"failures"`

	// ID         string
	// Name       string
	// Version    string
	// Roles      []string
	// Attributes map[string]interface{}
}

type singleConnectionPool struct {
	connection *Connection
}

type roundRobinConnectionPool struct {
	sync.Mutex

	list []*Connection
	curr int

	dead []*Connection
}

// newRoundRobinConnectionPool creates a new roundRobinConnectionPool.
//
func newRoundRobinConnectionPool(connections ...*Connection) *roundRobinConnectionPool {
	cp := roundRobinConnectionPool{
		list: connections,
	}

	if metrics != nil {
		metrics.Lock()
		metrics.Pool = cp.list
		metrics.Dead = cp.dead
		metrics.Unlock()
	}

	// DRAFT: Resurrector
	go func(cp *roundRobinConnectionPool) {
		var (
			timeoutInitial      = 60 * time.Second
			timeoutFactorCutoff = 5
		)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cp.Lock()
				for _, c := range cp.dead {
					factor := func(a, b int) float64 {
						if a > b {
							return float64(b)
						}
						return float64(a)
					}(c.Failures-1, timeoutFactorCutoff)

					// fmt.Printf("factor: %f, math.Exp2(factor): %f\n", factor, math.Exp2(factor))
					timeout := time.Duration(timeoutInitial.Seconds() * math.Exp2(factor) * float64(time.Second))
					fmt.Printf("Resurrect %s (failures=%d, factor=%1.1f, timeout=%s) in %s\n", c.URL, c.Failures, factor, timeout, c.DeadSince.Add(timeout).Sub(time.Now().UTC()))

					if time.Now().UTC().After(c.DeadSince.Add(timeout)) {
						fmt.Printf("Resurrecting %s, timeout passed\n", c.URL)
						c.Dead = false
						cp.list = append(cp.list, c)
						index := -1
						for i, conn := range cp.dead {
							if conn == c {
								index = i
							}
						}
						if index >= 0 {
							// Remove item; https://github.com/golang/go/wiki/SliceTricks
							copy(cp.dead[index:], cp.dead[index+1:])
							cp.dead[len(cp.dead)-1] = nil
							cp.dead = cp.dead[:len(cp.dead)-1]
						}
					}
				}
				cp.Unlock()
			}
		}
	}(&cp)

	return &cp
}

// Next returns a connection from pool, or an error.
//
func (cp *roundRobinConnectionPool) Next() (*Connection, error) {
	var c *Connection

	cp.Lock()
	defer cp.Unlock()

	// fmt.Println("Next()", cp.list)

	// Try to resurrect a dead connection if no healthy connections are available
	//
	if len(cp.list) < 1 {
		if len(cp.dead) > 0 {
			fmt.Println("Next()", cp.dead)
			fmt.Printf("Resurrecting connection...")
			c, cp.dead = cp.dead[len(cp.dead)-1], cp.dead[:len(cp.dead)-1] // Pop item
			fmt.Println(c)
			c.Dead = false
			cp.list = append(cp.list, c)

			if metrics != nil {
				metrics.Lock()
				metrics.Pool = cp.list
				metrics.Dead = cp.dead
				metrics.Unlock()
			}

			return c, nil
		}
		return nil, errors.New("no connection available")
	}

	if cp.curr >= len(cp.list) {
		cp.curr = len(cp.list) - 1
	}

	if cp.curr < 0 {
		return nil, errors.New("no connection available")
	}

	c = cp.list[cp.curr]
	cp.curr = (cp.curr + 1) % len(cp.list)

	return c, nil
}

// Remove removes a connection from the pool.
//
func (cp *roundRobinConnectionPool) Remove(c *Connection) error {
	c.Lock()
	fmt.Printf("Removing %s...\n", c.URL)
	c.Dead = true
	c.DeadSince = time.Now().UTC()
	c.Failures++
	c.Unlock()

	cp.Lock()
	defer cp.Unlock()

	// Push item to dead list and sort slice by number of failures
	cp.dead = append(cp.dead, c)
	sort.Slice(cp.dead, func(i, j int) bool {
		c1 := cp.dead[i]
		c2 := cp.dead[j]
		c1.Lock()
		c2.Lock()

		res := c1.Failures > c2.Failures
		c1.Unlock()
		c2.Unlock()
		return res
	})

	if metrics != nil {
		metrics.Lock()
		metrics.Dead = cp.dead
		metrics.Unlock()
	}

	// Check if connection exists in the list. Return nil if it doesn't, because another
	// goroutine might have already removed it.
	index := -1
	for i, conn := range cp.list {
		if conn == c {
			index = i
		}
	}
	if index < 0 {
		return nil
	}

	// Remove item; https://github.com/golang/go/wiki/SliceTricks
	copy(cp.list[index:], cp.list[index+1:])
	cp.list[len(cp.list)-1] = nil
	cp.list = cp.list[:len(cp.list)-1]

	if metrics != nil {
		metrics.Lock()
		metrics.Pool = cp.list
		metrics.Unlock()
	}

	return nil
}

func (c *Connection) String() string {
	c.Lock()
	defer c.Unlock()
	return fmt.Sprintf("<%s> dead=%v failures=%d", c.URL, c.Dead, c.Failures)
}
