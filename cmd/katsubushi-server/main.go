package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const WorkerIdEnd = 1024

type IdMap struct {
	ExpireDuration time.Duration

	lastId int
	list   []*IdInfo
	mu     sync.Mutex
}

type IdInfo struct {
	Id         int       `json:"id"`
	ReleasedAt time.Time `json:"released_at"`
	ExpireAt   time.Time `json:"expire_at"`
}

func NewIdMap() *IdMap {
	list := make([]*IdInfo, WorkerIdEnd)

	// Id: 0 is reserved.
	list[0] = &IdInfo{
		Id:       0,
		ExpireAt: time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC),
	}

	return &IdMap{
		ExpireDuration: 24 * time.Hour,
		list:           list,
	}
}

func (m *IdMap) New() *IdInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	// search unused worker id
	for i := 1; i <= WorkerIdEnd; i++ {
		id := (m.lastId + i) % WorkerIdEnd
		info := m.list[id]
		if info == nil {
			now := time.Now()
			newInfo := &IdInfo{
				Id:         id,
				ReleasedAt: now,
				ExpireAt:   now.Add(m.ExpireDuration),
			}
			m.list[id] = newInfo
			m.lastId = id
			return newInfo
		}
	}
	return nil
}

func (m *IdMap) Update(id int) *IdInfo {
	if id <= 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	idInfo := m.list[id]
	if idInfo == nil {
		return nil
	}

	now := time.Now()
	idInfo.ExpireAt = now.Add(m.ExpireDuration)

	return idInfo
}

func (m *IdMap) Delete(id int) bool {
	if id <= 0 {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.list[id] == nil {
		return false
	}
	m.list[id] = nil
	return true
}

func (m *IdMap) HandlerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	info := m.New()
	encoder := json.NewEncoder(w)
	encoder.Encode(info)
}

func (m *IdMap) HandlerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	info := m.Update(id)

	encoder := json.NewEncoder(w)
	encoder.Encode(info)
}

func (m *IdMap) HandlerDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	deleted := m.Delete(id)

	encoder := json.NewEncoder(w)
	encoder.Encode(struct {
		Deleted bool `json:"deleted"`
	}{
		Deleted: deleted,
	})
}

func main() {
	flag.Parse()

	m := NewIdMap()

	http.HandleFunc("/new", m.HandlerNew)
	http.HandleFunc("/update", m.HandlerUpdate)
	http.HandleFunc("/delete", m.HandlerDelete)
	http.ListenAndServe(":8080", nil)
}
