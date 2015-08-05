package katsubushi

import (
	"sync"
	"time"
)

type MemoryServer struct {
	ExpireDuration time.Duration

	lastId int
	list   []IdInfo
	mu     sync.Mutex
}

func NewMemoryServer() *MemoryServer {
	list := make([]IdInfo, WorkerIdEnd)

	// Id: 0 is reserved.
	list[0].ExpireAt = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)

	for i := 0; i < WorkerIdEnd; i++ {
		list[i].Id = i
	}

	return &MemoryServer{
		ExpireDuration: 24 * time.Hour,
		list:           list,
	}
}

func (m *MemoryServer) New() (*IdInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// search oldest worker id
	idInfo := &m.list[(m.lastId+1)%WorkerIdEnd]
	for i := 2; i <= WorkerIdEnd; i++ {
		id := (m.lastId + i) % WorkerIdEnd
		info := &m.list[id]
		if info.ExpireAt.Before(idInfo.ExpireAt) {
			idInfo = info
		}
	}

	// check expired
	now := time.Now()
	if idInfo.ExpireAt.Before(now) {
		idInfo.ReleasedAt = now
		idInfo.ExpireAt = now.Add(m.ExpireDuration)
		return idInfo, nil
	}
	return nil, NoAvailableId
}

func (m *MemoryServer) Update(id int) (*IdInfo, error) {
	if id <= 0 || id >= WorkerIdEnd {
		return nil, NoAvailableId
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	idInfo := &m.list[id]
	if idInfo.ExpireAt.Before(now) {
		return nil, NoAvailableId
	}

	idInfo.ExpireAt = now.Add(m.ExpireDuration)

	return idInfo, nil
}

func (m *MemoryServer) Delete(id int) error {
	if id <= 0 || id >= WorkerIdEnd {
		return NoAvailableId
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	idInfo := &m.list[id]
	if idInfo.ExpireAt.Before(now) {
		return NoAvailableId
	}

	idInfo.ExpireAt = now.Add(-1 * time.Nanosecond)

	return nil
}
