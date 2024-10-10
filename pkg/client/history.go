// Copyright 2024 Qian Yao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bufio"
	"container/ring"
	"os"
	"path/filepath"
	"sync"

	"gsmate/config"

	"github.com/vimiix/pkg/file"
)

const MaxHistory = 1000

type History struct {
	mu      *sync.Mutex
	records *ring.Ring
}

func NewHistory(n int) (*History, error) {
	if n <= 0 {
		n = MaxHistory
	}
	h := &History{
		mu:      &sync.Mutex{},
		records: ring.New(n),
	}
	if err := h.loadRecords(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *History) Records() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	records := make([]string, 0, h.records.Len())
	h.records.Do(func(a any) {
		if a == nil {
			return
		}
		records = append(records, a.(string))
	})
	return records
}

func (h *History) loadRecords() error {
	file := historyFile()
	if _, err := os.Stat(file); err != nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		h.records.Value = scanner.Text()
		h.records = h.records.Next()
	}
	return scanner.Err()
}

func (h *History) Add(s string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records.Value = s
	h.records = h.records.Next()
}

func (h *History) Persist() error {
	hisFile := historyFile()
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := file.EnsureDirExists(hisFile); err != nil {
		return err
	}
	f, err := os.Create(hisFile)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	h.records.Do(func(a any) {
		if a == nil {
			return
		}
		_, _ = w.WriteString(a.(string) + "\n")
	})
	return w.Flush()
}

func historyFile() string {
	return filepath.Join(config.DefaultLocation(), "history")
}
