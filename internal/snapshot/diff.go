package snapshot

import "fmt"

// FieldDiff describes a single field change between a snapshot and the current live state.
type FieldDiff struct {
	Field    string
	Snapshot interface{}
	Live     interface{}
}

// String returns a human-readable representation of the diff.
func (d FieldDiff) String() string {
	return fmt.Sprintf("field %q: snapshot=%v live=%v", d.Field, d.Snapshot, d.Live)
}

// DiffResult holds the comparison outcome for one service.
type DiffResult struct {
	ServiceName string
	Diffs       []FieldDiff
}

// HasDrift reports whether any fields changed since the snapshot.
func (r DiffResult) HasDrift() bool {
	return len(r.Diffs) > 0
}

// Compare compares a previously saved snapshot against a current live field map.
// Fields present in the snapshot but missing from live are flagged as drift.
func Compare(snap Snapshot, live map[string]interface{}) DiffResult {
	result := DiffResult{ServiceName: snap.ServiceName}

	for key, snapVal := range snap.Fields {
		liveVal, ok := live[key]
		if !ok {
			result.Diffs = append(result.Diffs, FieldDiff{
				Field:    key,
				Snapshot: snapVal,
				Live:     nil,
			})
			continue
		}
		if fmt.Sprintf("%v", snapVal) != fmt.Sprintf("%v", liveVal) {
			result.Diffs = append(result.Diffs, FieldDiff{
				Field:    key,
				Snapshot: snapVal,
				Live:     liveVal,
			})
		}
	}

	return result
}
