package alert

import (
	"fmt"
	"io"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Level represents the severity of a drift alert.
type Level string

const (
	LevelNone     Level = "none"
	LevelWarning  Level = "warning"
	LevelCritical Level = "critical"
)

// Alert holds a formatted drift notification for a single service.
type Alert struct {
	ServiceName string
	Level       Level
	Message     string
	Fields      []string
}

// Alerter evaluates drift results and emits alerts.
type Alerter struct {
	w               io.Writer
	criticalFields  []string
}

// New creates an Alerter that writes to w.
// criticalFields lists field names whose drift is always critical.
func New(w io.Writer, criticalFields []string) *Alerter {
	return &Alerter{w: w, criticalFields: criticalFields}
}

// Evaluate converts drift results into Alerts.
func (a *Alerter) Evaluate(results []drift.Result) []Alert {
	var alerts []Alert
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		lvl := LevelWarning
		for _, f := range r.DriftedFields {
			if a.isCritical(f) {
				lvl = LevelCritical
				break
			}
		}
		alerts = append(alerts, Alert{
			ServiceName: r.ServiceName,
			Level:       lvl,
			Message:     fmt.Sprintf("drift detected in service %q: %s", r.ServiceName, strings.Join(r.DriftedFields, ", ")),
			Fields:      r.DriftedFields,
		})
	}
	return alerts
}

// Emit writes all alerts to the configured writer.
func (a *Alerter) Emit(alerts []Alert) {
	for _, al := range alerts {
		fmt.Fprintf(a.w, "[%s] %s\n", strings.ToUpper(string(al.Level)), al.Message)
	}
}

func (a *Alerter) isCritical(field string) bool {
	for _, cf := range a.criticalFields {
		if strings.EqualFold(cf, field) {
			return true
		}
	}
	return false
}
