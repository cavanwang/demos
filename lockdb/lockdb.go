package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	tryGetLockEachMilliseconds = 50
)

var (
	ErrLockTimeout = errors.New("lock timeout")
)

type Lock struct {
	ID          uint
	Name        string `gorm:"uniqueIndex"`
	CreateAt    time.Time
	HeartbeatAt time.Time
	Version     string
	stopCh      chan struct{}
}

func GetLock(db *gorm.DB, name string, timeoutSecond int) (lock *Lock, err error) {
	expire := time.Now().Add(time.Duration(timeoutSecond) * time.Second)
	for ; ; time.Sleep(time.Millisecond * tryGetLockEachMilliseconds) {
		// Detect if the lock record existed.
		lock = &Lock{}
		result := db.Take(lock, "name=?", name)
		if result.Error != nil {
			if isDuplicateKeyError(result.Error) {
				result.Error = gorm.ErrDuplicatedKey
				return nil, result.Error
			}
		}

		if result.Error == nil {
			// Try too many times, timeout.
			if time.Since(expire) > 0 {
				return nil, ErrLockTimeout
			}
			continue
		}

		// Try to insert a record to exclusively preempt the lock.
		lock = &Lock{Name: name, CreateAt: time.Now(), HeartbeatAt: time.Now(), Version: uuid.NewV4().String()}
		result = db.Create(lock)

		// Preempted the lock successfully, make a heartbeat goroutine and return.
		if result.Error == nil {
			lock.stopCh = make(chan struct{})
			go func() {
				for {
					select {
					case <-lock.stopCh:
						return
					case <-time.After(time.Second * 5):
						lock.HeartbeatAt = time.Now()
						result := db.Save(lock)
						if result.Error != nil {
							fmt.Printf("%v: save lock error=%v\n", time.Now(), result.Error)
							return
						}
						fmt.Printf("%v: save lock ok, heartbeat=%v\n", time.Now(), lock.HeartbeatAt)
					}
				}
			}()

			return lock, nil
		}

		if !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			fmt.Printf("%v: error is not gorm.ErrDuplicatedKey %v\n", time.Now(), result.Error)
			return nil, result.Error
		}

		// DB has already a record, wait for other locker exiting.
		if time.Since(expire) > 0 {
			return nil, ErrLockTimeout
		}
	}
}

func ReleaseLock(db *gorm.DB, lock *Lock) error {
	close(lock.stopCh)
	result := db.Delete(lock)
	return result.Error
}

func ReleaseTimeoutLock(db *gorm.DB, name string, heartbeatTimeoutSecond int) (released bool, err error) {
	var lock Lock
	result := db.Take(&lock, "name=?", name)
	if result.Error == gorm.ErrRecordNotFound {
		return true, nil
	}
	if result.Error != nil {
		return false, result.Error
	}

	if time.Since(lock.HeartbeatAt) > time.Duration(heartbeatTimeoutSecond)*time.Second {
		result = db.Where("name=? and version = ?", name, lock.Version).Delete(&Lock{})
		if result.Error != nil {
			return false, result.Error
		}
		return true, nil
	} else {
		fmt.Printf("heartbeat is %v\n", lock.HeartbeatAt)
	}
	return false, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint")
}

func main() {
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&Lock{}); err != nil {
		panic(err)
	}

	ReleaseTimeoutLock(db, "name1", 10)

	wg := sync.WaitGroup{}
	f := func(name string) {
		defer fmt.Printf("%v: %s exited\n", time.Now(), name)
		defer wg.Done()

		start := time.Now()
		var lock *Lock
		for {
			lock, err = GetLock(db, "name1", 10)
			if err != nil {
				fmt.Printf("%v: %s: lock name1 error=%v\n", time.Now(), name, err)
				if err != ErrLockTimeout {
					return
				}
			} else {
				break
			}
		}
		fmt.Printf("%v: %s locked cost=%dms\n", time.Now(), name, time.Since(start)/time.Millisecond)
		err = ReleaseLock(db, lock)
		if err != nil {
			fmt.Printf("%v: %s release name1 error=%v\n", time.Now(), name, err)
			return
		}
		fmt.Printf("%v: %s release name1 ok\n", time.Now(), name)
	}
	wg.Add(2)
	fmt.Printf("%v: started\n", time.Now())
	go f("lock1")
	go f("lock2")
	wg.Wait()
}
