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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunks(t *testing.T) {
	t.Run("Empty Input", func(t *testing.T) {
		vals := []string{}
		chunks := Chunks(vals)
		assert.Empty(t, chunks)
	})

	t.Run("Divides Without Remainder", func(t *testing.T) {
		vals := []string{"apple", "banana", "cherry", "date"}
		getWindowSize = func(int) (int, int, error) { return 20, 0, nil }
		chunks := Chunks(vals)
		expected := [][]string{{"apple", "banana"}, {"cherry", "date"}}
		assert.Equal(t, expected, chunks)
	})

	t.Run("Divides With Remainder", func(t *testing.T) {
		vals := []string{"apple", "banana", "cherry", "date", "elderberry"}
		getWindowSize = func(int) (int, int, error) { return 40, 0, nil }
		chunks := Chunks(vals)
		expected := [][]string{{"apple", "banana", "cherry"}, {"date", "elderberry"}}
		assert.Equal(t, expected, chunks)
	})
}
