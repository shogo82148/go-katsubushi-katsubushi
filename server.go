package katsubushi

import (
	"errors"
	"time"
)

const WorkerIdEnd = 1024

var NoAvailableId = errors.New("no available id")

type IdInfo struct {
	Id         int       `json:"id"`
	ReleasedAt time.Time `json:"released_at"`
	ExpireAt   time.Time `json:"expire_at"`
}

type IdGenerator interface {
	// New creates new worker-id
	New() (*IdInfo, error)

	// Update extends expire time of id
	Update(id int) (*IdInfo, error)

	// Delete deletes id
	Delete(id int) error
}
