/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package eventhub

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// setupFileTestDB mirrors the production shape the in-memory helper does not:
// a file-backed SQLite database with a real connection pool, so the poller and
// publishers run on different connections like in gateway-controller.
func setupFileTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "eventhub.db")
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS gateway_states (
			gateway_id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS events (
			gateway_id TEXT NOT NULL,
			processed_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			originated_timestamp TIMESTAMP NOT NULL,
			entity_type TEXT NOT NULL,
			action TEXT NOT NULL CHECK(action IN ('CREATE', 'UPDATE', 'DELETE')),
			entity_id TEXT NOT NULL,
			event_id TEXT NOT NULL,
			event_data TEXT NOT NULL,
			PRIMARY KEY (gateway_id, event_id),
			FOREIGN KEY (gateway_id) REFERENCES gateway_states(gateway_id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_events_gateway_id_processed_timestamp ON events(gateway_id, processed_timestamp);
	`)
	require.NoError(t, err)

	t.Cleanup(func() { db.Close() })
	return db
}

// collectDelivered drains ch until no event arrives for the quiet period,
// returning the first-delivery order of event IDs and the set of all
// deliveries (duplicates from boundary replay are tolerated).
func collectDelivered(ch <-chan Event, quiet time.Duration) (firstSeenOrder []string, seen map[string]int) {
	seen = make(map[string]int)
	for {
		select {
		case evt := <-ch:
			if seen[evt.EventID] == 0 {
				firstSeenOrder = append(firstSeenOrder, evt.EventID)
			}
			seen[evt.EventID]++
		case <-time.After(quiet):
			return firstSeenOrder, seen
		}
	}
}

// TestRapidSequentialPublishesAllDelivered reproduces an event loss observed in
// the gateway integration tests once fixed inter-mutation sleeps were removed:
// events published back-to-back (like consecutive REST deploy/delete calls)
// while the poller is concurrently delivering could be skipped or delivered out
// of order. Every published event must be delivered, and first deliveries must
// preserve publish order.
//
// In gateway-controller the eventhub shares its SQLite file with the storage
// layer, whose writes contend for the database write lock. A publish can then
// stall between binding processed_timestamp (time.Now at Exec) and committing,
// letting a later publish commit first with an earlier timestamp — after which
// the poller's lastPolled cursor (>= max delivered timestamp) skips the stalled
// event forever. The writePressure goroutine recreates that contention.
func TestRapidSequentialPublishesAllDelivered(t *testing.T) {
	db := setupFileTestDB(t)

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS pressure (id INTEGER PRIMARY KEY AUTOINCREMENT, payload TEXT)`)
	require.NoError(t, err)

	config := Config{
		PollInterval:    2 * time.Millisecond,
		CleanupInterval: 5 * time.Minute,
		RetentionPeriod: 1 * time.Hour,
	}
	hub := New(db, testLogger(), config)
	require.NoError(t, hub.Initialize())
	defer hub.Close()

	const gatewayID = "rapid-gw"
	require.NoError(t, hub.RegisterGateway(gatewayID))

	ch, err := hub.Subscribe(gatewayID)
	require.NoError(t, err)

	// Contending writer, standing in for gateway-controller's storage layer
	// sharing the database file with the eventhub.
	pressureDone := make(chan struct{})
	pressureStop := make(chan struct{})
	go func() {
		defer close(pressureDone)
		payload := make([]byte, 4096)
		for i := 0; ; i++ {
			select {
			case <-pressureStop:
				return
			default:
			}
			tx, err := db.Begin()
			if err != nil {
				continue
			}
			for j := 0; j < 8; j++ {
				_, _ = tx.Exec(`INSERT INTO pressure (payload) VALUES (?)`, string(payload))
			}
			_ = tx.Commit()
		}
	}()
	defer func() { close(pressureStop); <-pressureDone }()

	const total = 400
	published := make([]string, 0, total)
	for i := 0; i < total; i++ {
		action := "CREATE"
		if i%2 == 1 {
			action = "DELETE"
		}
		eventID := fmt.Sprintf("evt-%04d", i)
		require.NoError(t, hub.PublishEvent(gatewayID, Event{
			GatewayID:           gatewayID,
			OriginatedTimestamp: time.Now(),
			EventType:           EventTypeAPI,
			Action:              action,
			EntityID:            fmt.Sprintf("entity-%04d", i/2),
			EventID:             eventID,
			EventData:           EmptyEventData,
		}))
		published = append(published, eventID)
	}

	firstSeenOrder, seen := collectDelivered(ch, 2*time.Second)

	var missing []string
	for _, id := range published {
		if seen[id] == 0 {
			missing = append(missing, id)
		}
	}
	require.Emptyf(t, missing, "%d of %d events were never delivered: %v", len(missing), total, missing)
	require.Equal(t, published, firstSeenOrder, "first deliveries must preserve publish order")
}

// TestConcurrentPublishersNoEventLoss covers the multi-writer shape: the REST
// handler and the control-plane sync client can publish for the same gateway
// from different goroutines. Regardless of interleaving, no event may be lost.
func TestConcurrentPublishersNoEventLoss(t *testing.T) {
	db := setupFileTestDB(t)

	config := Config{
		PollInterval:    2 * time.Millisecond,
		CleanupInterval: 5 * time.Minute,
		RetentionPeriod: 1 * time.Hour,
	}
	hub := New(db, testLogger(), config)
	require.NoError(t, hub.Initialize())
	defer hub.Close()

	const gatewayID = "concurrent-gw"
	require.NoError(t, hub.RegisterGateway(gatewayID))

	ch, err := hub.Subscribe(gatewayID)
	require.NoError(t, err)

	const publishers = 4
	const perPublisher = 100
	var wg sync.WaitGroup
	for p := 0; p < publishers; p++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			for i := 0; i < perPublisher; i++ {
				err := hub.PublishEvent(gatewayID, Event{
					GatewayID:           gatewayID,
					OriginatedTimestamp: time.Now(),
					EventType:           EventTypeAPI,
					Action:              "UPDATE",
					EntityID:            fmt.Sprintf("entity-%d", p),
					EventID:             fmt.Sprintf("evt-%d-%04d", p, i),
					EventData:           EmptyEventData,
				})
				require.NoError(t, err)
			}
		}(p)
	}
	wg.Wait()

	_, seen := collectDelivered(ch, 1*time.Second)

	var missing []string
	for p := 0; p < publishers; p++ {
		for i := 0; i < perPublisher; i++ {
			id := fmt.Sprintf("evt-%d-%04d", p, i)
			if seen[id] == 0 {
				missing = append(missing, id)
			}
		}
	}
	require.Emptyf(t, missing, "%d of %d events were never delivered: %v", len(missing), publishers*perPublisher, missing)
}
