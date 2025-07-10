package door

import (
	"bytes"
	cryptoRand "crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

type Door struct {
	pin                rpio.Pin
	mutex              sync.Mutex
	delayedUnlockNonce []byte
}

func New() Door {
	pin := rpio.Pin(21)
	pin.Output()
	pin.Low()
	return Door{pin: pin}
}

const maxDuration = time.Second * 30

func (c *Door) UnlockForDuration(duration time.Duration, authorizedBy string) error {
	if duration > maxDuration {
		return fmt.Errorf("duration (%.0f) is longer than maximum allowed (%.0f)", duration.Seconds(), maxDuration.Seconds())
	}

	if duration <= 0 {
		return fmt.Errorf("duration (%.0f) must be greater than 0", duration.Seconds())
	}

	c.unlock(authorizedBy)

	delayedUnlockNonce := make([]byte, 32)
	cryptoRand.Read(delayedUnlockNonce)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delayedUnlockNonce = delayedUnlockNonce

	go func() {
		time.Sleep(duration)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		if bytes.Equal(c.delayedUnlockNonce, delayedUnlockNonce) {
			c.lock(authorizedBy)
		}
	}()

	return nil
}

func (c *Door) unlock(authorizedBy string) {
	c.pin.High()
	log.Printf("door unlocked (%s)\n", authorizedBy)
}

func (c *Door) Unlock(authorizedBy string) {
	c.unlock(authorizedBy)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delayedUnlockNonce = nil
}

func (c *Door) lock(authorizedBy string) {
	c.pin.Low()
	log.Printf("door locked (%s)\n", authorizedBy)
}

func (c *Door) Lock(authorizedBy string) {
	c.lock(authorizedBy)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delayedUnlockNonce = nil
}
