package metrics

import "testing"

func TestTrackerSnapshot(t *testing.T) {
	tracker := NewTracker()
	tracker.RecordAttempt()
	tracker.RecordAttempt()
	tracker.RecordSuccess()
	tracker.RecordBlocked("cooldown")
	snap := tracker.Snapshot()
	if snap.ExecutionAttempts != 2 || snap.ExecutionSuccess != 1 || snap.ExecutionBlocked != 1 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
	if snap.SuccessRate != 0.5 || snap.BlockedRate != 0.5 {
		t.Fatalf("unexpected rate values: %+v", snap)
	}
	if snap.BlockReasons["cooldown"] != 1 {
		t.Fatalf("expected cooldown block reason count, got %+v", snap.BlockReasons)
	}
}
