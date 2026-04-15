package baseline

import "fmt"

// Deviation describes a single field difference between a baseline and live state.
type Deviation struct {
	Field    string
	Baseline interface{}
	Live     interface{}
}

// CompareResult holds the outcome of comparing a baseline to live fields.
type CompareResult struct {
	ServiceName string
	Deviations  []Deviation
	InSync      bool
}

// Summary returns a human-readable string describing the compare result.
// If the result is in sync, it reports no deviations. Otherwise it lists
// the number of deviations found for the service.
func (r CompareResult) Summary() string {
	if r.InSync {
		return fmt.Sprintf("%s: in sync", r.ServiceName)
	}
	return fmt.Sprintf("%s: %d deviation(s) found", r.ServiceName, len(r.Deviations))
}

// Compare checks live against the stored baseline entry.
// Fields present in the baseline but missing from live are flagged.
// Extra fields in live that are not in the baseline are ignored.
func Compare(entry Entry, live map[string]interface{}) CompareResult {
	result := CompareResult{
		ServiceName: entry.ServiceName,
		InSync:      true,
	}

	for key, baseVal := range entry.Fields {
		liveVal, ok := live[key]
		if !ok {
			result.Deviations = append(result.Deviations, Deviation{
				Field:    key,
				Baseline: baseVal,
				Live:     nil,
			})
			result.InSync = false
			continue
		}
		if fmt.Sprintf("%v", baseVal) != fmt.Sprintf("%v", liveVal) {
			result.Deviations = append(result.Deviations, Deviation{
				Field:    key,
				Baseline: baseVal,
				Live:     liveVal,
			})
			result.InSync = false
		}
	}
	return result
}
