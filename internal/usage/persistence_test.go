package usage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sampleSnapshot() StatisticsSnapshot {
	return StatisticsSnapshot{
		APIs: map[string]APISnapshot{
			"test-api": {
				Models: map[string]ModelSnapshot{
					"gpt-test": {
						Details: []RequestDetail{
							{
								Timestamp: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
								Source:    "openai",
								AuthIndex: "0",
								Failed:    false,
								Tokens: TokenStats{
									InputTokens:  10,
									OutputTokens: 20,
									TotalTokens:  30,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestWriteAndLoadSnapshot(t *testing.T) {
	stats := NewRequestStatistics()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "usage-statistics.json")

	if err := writeSnapshotToFile(path, sampleSnapshot()); err != nil {
		t.Fatalf("writeSnapshotToFile failed: %v", err)
	}

	if err := loadSnapshotFromFile(stats, path); err != nil {
		t.Fatalf("loadSnapshotFromFile failed: %v", err)
	}

	snapshot := stats.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Fatalf("expected total_requests=1, got %d", snapshot.TotalRequests)
	}
	if snapshot.TotalTokens != 30 {
		t.Fatalf("expected total_tokens=30, got %d", snapshot.TotalTokens)
	}
	api := snapshot.APIs["test-api"]
	model := api.Models["gpt-test"]
	if model.TotalRequests != 1 {
		t.Fatalf("expected model total_requests=1, got %d", model.TotalRequests)
	}
}

func TestLoadLegacyRawSnapshot(t *testing.T) {
	stats := NewRequestStatistics()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "legacy-snapshot.json")
	data, err := json.Marshal(sampleSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot failed: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write legacy snapshot failed: %v", err)
	}

	if err := loadSnapshotFromFile(stats, path); err != nil {
		t.Fatalf("loadSnapshotFromFile failed: %v", err)
	}

	snapshot := stats.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Fatalf("expected total_requests=1, got %d", snapshot.TotalRequests)
	}
	if snapshot.SuccessCount != 1 {
		t.Fatalf("expected success_count=1, got %d", snapshot.SuccessCount)
	}
}
