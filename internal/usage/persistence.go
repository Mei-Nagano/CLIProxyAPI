package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const defaultPersistenceInterval = 30 * time.Second

type statisticsPersistencePayload struct {
	Version int                `json:"version"`
	SavedAt time.Time          `json:"saved_at"`
	Usage   StatisticsSnapshot `json:"usage"`
}

// StartStatisticsPersistence restores usage statistics from disk and keeps
// periodically flushing in-memory state to the provided file path.
func StartStatisticsPersistence(ctx context.Context, path string, interval time.Duration) {
	if ctx == nil {
		return
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	if interval <= 0 {
		interval = defaultPersistenceInterval
	}

	stats := GetRequestStatistics()
	if stats == nil {
		return
	}

	if err := loadSnapshotFromFile(stats, path); err != nil {
		log.Warnf("usage: failed to restore statistics from %s: %v", path, err)
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		lastSavedTotalRequests := stats.Snapshot().TotalRequests
		for {
			select {
			case <-ctx.Done():
				if err := saveSnapshotToFile(stats, path); err != nil {
					log.Warnf("usage: failed to flush statistics to %s on shutdown: %v", path, err)
				}
				return
			case <-ticker.C:
				snapshot := stats.Snapshot()
				if snapshot.TotalRequests == lastSavedTotalRequests {
					continue
				}
				if err := writeSnapshotToFile(path, snapshot); err != nil {
					log.Warnf("usage: failed to persist statistics to %s: %v", path, err)
					continue
				}
				lastSavedTotalRequests = snapshot.TotalRequests
			}
		}
	}()
}

func loadSnapshotFromFile(stats *RequestStatistics, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err == nil {
		if _, ok := root["usage"]; ok {
			var payload statisticsPersistencePayload
			if err := json.Unmarshal(data, &payload); err != nil {
				return err
			}
			if payload.Version != 0 && payload.Version != 1 {
				return fmt.Errorf("unsupported usage statistics snapshot version: %d", payload.Version)
			}
			result := stats.MergeSnapshot(payload.Usage)
			if result.Added > 0 || result.Skipped > 0 {
				log.Infof("usage: restored snapshot from %s (added=%d, skipped=%d)", path, result.Added, result.Skipped)
			}
			return nil
		}
	}

	// Backward compatibility: accept raw StatisticsSnapshot JSON payloads.
	var snapshot StatisticsSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	result := stats.MergeSnapshot(snapshot)
	if result.Added > 0 || result.Skipped > 0 {
		log.Infof("usage: restored legacy snapshot from %s (added=%d, skipped=%d)", path, result.Added, result.Skipped)
	}
	return nil
}

func saveSnapshotToFile(stats *RequestStatistics, path string) error {
	if stats == nil {
		return nil
	}
	return writeSnapshotToFile(path, stats.Snapshot())
}

func writeSnapshotToFile(path string, snapshot StatisticsSnapshot) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	payload := statisticsPersistencePayload{
		Version: 1,
		SavedAt: time.Now().UTC(),
		Usage:   snapshot,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		// Windows rename may fail if destination exists.
		if errWrite := os.WriteFile(path, data, 0o600); errWrite != nil {
			return errWrite
		}
	}
	return nil
}
