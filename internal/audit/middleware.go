package audit

import (
	"fmt"

	"driftwatch/internal/drift"
)

// DetectHook wraps a drift detection result and records an audit event.
// It is intended to be called after a detection run completes.
func DetectHook(log *Log, results []drift.Result) error {
	if log == nil {
		return nil
	}
	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
			meta := map[string]string{
				"drifted_fields": fmt.Sprintf("%d", len(r.Deltas)),
			}
			if err := log.Record(EventDetect, r.Service, "drift detected", meta); err != nil {
				return err
			}
		}
	}
	if drifted == 0 {
		if err := log.Record(EventDetect, "", "all services in sync", nil); err != nil {
			return err
		}
	}
	return nil
}

// BaselineHook records an audit event when a baseline is saved or deleted.
func BaselineHook(log *Log, service, action string) error {
	if log == nil {
		return nil
	}
	meta := map[string]string{"action": action}
	return log.Record(EventBaseline, service, fmt.Sprintf("baseline %s", action), meta)
}
