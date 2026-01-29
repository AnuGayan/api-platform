/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicManager(t *testing.T) {
	tm := NewTopicManager()

	configID := "api-1"
	topic1 := "news"
	topic2 := "sports"

	t.Run("Add", func(t *testing.T) {
		assert.True(t, tm.Add(configID, topic1))
		assert.False(t, tm.Add(configID, topic1)) // Duplicate
		assert.True(t, tm.Add(configID, topic2))
		assert.Equal(t, 2, tm.CountForConfig(configID))
	})

	t.Run("Contains", func(t *testing.T) {
		assert.True(t, tm.Contains(configID, topic1))
		assert.False(t, tm.Contains(configID, "weather"))
		assert.False(t, tm.Contains("non-existent", topic1))
	})

	t.Run("IsTopicExist", func(t *testing.T) {
		assert.True(t, tm.IsTopicExist(configID, topic1))
		assert.False(t, tm.IsTopicExist(configID, "weather"))
	})

	t.Run("GetAllByConfig", func(t *testing.T) {
		topics := tm.GetAllByConfig(configID)
		assert.ElementsMatch(t, []string{topic1, topic2}, topics)
		assert.Empty(t, tm.GetAllByConfig("empty-api"))
	})

	t.Run("GetAll", func(t *testing.T) {
		all := tm.GetAll()
		assert.True(t, all[topic1])
		assert.True(t, all[topic2])
		assert.Equal(t, 2, len(all))
	})

	t.Run("GetAllForConfig", func(t *testing.T) {
		full := tm.GetAllForConfig()
		assert.Contains(t, full, configID)
		assert.True(t, full[configID][topic1])
	})

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 2, tm.Count())
		tm.Add("api-2", "weather")
		assert.Equal(t, 3, tm.Count())
	})

	t.Run("Remove", func(t *testing.T) {
		assert.True(t, tm.Remove(configID, topic1))
		assert.False(t, tm.Remove(configID, topic1)) // Already removed
		assert.False(t, tm.Remove("non-existent", topic1))
		assert.Equal(t, 1, tm.CountForConfig(configID))
	})

	t.Run("RemoveAllForConfig", func(t *testing.T) {
		tm.RemoveAllForConfig(configID)
		assert.Equal(t, 0, tm.CountForConfig(configID))
	})

	t.Run("Clear", func(t *testing.T) {
		tm.Add(configID, topic1)
		tm.Clear()
		assert.Equal(t, 0, tm.Count())
	})
}
