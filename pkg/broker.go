package pkg

import (
	"errors"
	"fmt"
	"github.com/go-redsync/redsync"
	"github.com/golang/glog"
	"time"
)

type broker struct {
	mutexes map[string]*redsync.Mutex
}

func (b *broker) CheckIn(rangeValue string) (err error) {
	mutex := b.mutexes[rangeValue]

	succeeded := mutex.Unlock()

	if !succeeded {
		message := fmt.Sprintf("Unlock failed for [%s]", rangeValue)
		err = errors.New(message)
	}

	return err
}

func (b *broker) CheckOut() (string, error) {
	var acquiredMutex *string
	var err error
	checkoutSleepSeconds := 1

	for acquiredMutex == nil {
		for name, mutex := range b.mutexes {
			err := mutex.Lock()

			if err == nil {
				acquiredMutex = &name
				break
			} else {
				glog.Infof("Attempt to lock [%s] failed: %s", name, err)
			}
		}

		glog.Info("Iterated all mutexes, none available, sleeping [%v] seconds", checkoutSleepSeconds)

		time.Sleep(time.Duration(checkoutSleepSeconds) * time.Second)
		if checkoutSleepSeconds < 32 {
			checkoutSleepSeconds *= 2
		}
	}

	return *acquiredMutex, err
}

func NewBroker(rangeValues []string) broker {
	mutexes := createMutexes(rangeValues)

	return broker{
		mutexes: mutexes,
	}
}
