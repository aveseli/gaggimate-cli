// Package diag computes shot diagnostics from parsed shot data.
// Mirrors gaggimate-mcp/src/gaggimate_mcp/transformers/shot.py.
package diag

import (
	"math"
	"strings"
	"time"

	"github.com/adnan/gaggimate-cli/internal/parser"
)

// Detail levels
const (
	DetailSummary          = "summary"
	DetailPerPhase         = "per_phase"
	DetailPerPhaseDetailed = "per_phase_detailed"
)

// ─── Types ────────────────────────────────────────────────────────

type TransformedShot struct {
	ShotID          string      `json:"shot_id"`
	ProfileName     string      `json:"profile_name"`
	ProfileID       string      `json:"profile_id"`
	Timestamp       time.Time   `json:"timestamp"`
	DurationSeconds float64     `json:"duration_seconds"`
	FinalWeightG    *float64    `json:"final_weight_g,omitempty"`
	Summary         ShotSummary `json:"summary"`
	Phases          []PhaseData `json:"phases"`
	Diagnostics     interface{} `json:"diagnostics,omitempty"`
	DetailLevel     string      `json:"detail_level"`
}

type ShotSummary struct {
	Temperature TemperatureSummary `json:"temperature"`
	Pressure    PressureSummary    `json:"pressure"`
	Flow        FlowSummary        `json:"flow"`
	Extraction  ExtractionSummary  `json:"extraction"`
}

type TemperatureSummary struct {
	MinC       float64 `json:"min_c"`
	MaxC       float64 `json:"max_c"`
	AvgC       float64 `json:"avg_c"`
	TargetAvgC float64 `json:"target_avg_c"`
}

type PressureSummary struct {
	MinBar    float64 `json:"min_bar"`
	MaxBar    float64 `json:"max_bar"`
	AvgBar    float64 `json:"avg_bar"`
	PeakTimeS float64 `json:"peak_time_s"`
}

type FlowSummary struct {
	TotalVolumeML   float64  `json:"total_volume_ml"`
	AvgFlowMLS      float64  `json:"avg_flow_ml_s"`
	PeakFlowMLS     float64  `json:"peak_flow_ml_s"`
	TimeToFirstDrip *float64 `json:"time_to_first_drip_s,omitempty"`
}

type ExtractionSummary struct {
	PreinfusionTimeS    float64 `json:"preinfusion_time_s"`
	MainExtractionTimeS float64 `json:"main_extraction_time_s"`
	TotalTimeS          float64 `json:"total_time_s"`
}

type PhaseData struct {
	Name             string              `json:"name"`
	PhaseNumber      int                 `json:"phase_number"`
	StartTimeSeconds float64             `json:"start_time_seconds"`
	DurationSeconds  float64             `json:"duration_seconds"`
	SampleCount      int                 `json:"sample_count"`
	AvgTemperatureC  float64             `json:"avg_temperature_c"`
	AvgPressureBar   float64             `json:"avg_pressure_bar"`
	TotalFlowML      float64             `json:"total_flow_ml"`
	Samples          []TransformedSample `json:"samples,omitempty"`
	Diagnostics      interface{}         `json:"diagnostics,omitempty"`
}

type TransformedSample struct {
	TimeSeconds  float64 `json:"time_seconds"`
	TemperatureC float64 `json:"temperature_c"`
	PressureBar  float64 `json:"pressure_bar"`
	FlowMLS      float64 `json:"flow_ml_s"`
	WeightG      float64 `json:"weight_g"`
}

type SummaryDiagnostics struct {
	ResistanceAvg       float64           `json:"resistance_avg"`
	ResistanceSlope     float64           `json:"resistance_slope"`
	ChannelingRisk      string            `json:"channeling_risk"`
	TemperatureStabilityC float64         `json:"temperature_stability_c"`
	PressureRMSEBar     float64           `json:"pressure_rmse_bar"`
	MaxOvershootBar     float64           `json:"max_overshoot_bar"`
	FlowRMSEMLS         *float64          `json:"flow_rmse_ml_s,omitempty"`
	MaxFlowOvershootMLS *float64          `json:"max_flow_overshoot_ml_s,omitempty"`
	ScaleConnected      bool              `json:"scale_connected"`
	Annotations         map[string]string `json:"annotations"`
}

type ShotDiagnostics struct {
	Resistance        ResistanceDiagnostics `json:"resistance"`
	Channeling        ChannelingIndicators  `json:"channeling"`
	Temperature       TemperatureDiagnostics `json:"temperature"`
	Extraction        ExtractionMetrics     `json:"extraction"`
	Weight            WeightDiagnostics     `json:"weight"`
	ProfileCompliance *ProfileCompliance    `json:"profile_compliance,omitempty"`
}

type ResistanceDiagnostics struct {
	Avg           float64           `json:"avg"`
	Std           float64           `json:"std"`
	Slope         float64           `json:"slope"`
	Peak          float64           `json:"peak"`
	PeakTimingPct float64           `json:"peak_timing_pct"`
	Annotations   map[string]string `json:"annotations"`
}

type ChannelingIndicators struct {
	FlowJitterMLS               float64           `json:"flow_jitter_ml_s"`
	FlowVsTargetResidualMLS     *float64          `json:"flow_vs_target_residual_ml_s,omitempty"`
	PressureMaxDropRateBarS     float64           `json:"pressure_max_drop_rate_bar_s"`
	FlowAccelerationLateMLS2    float64           `json:"flow_acceleration_late_ml_s2"`
	FlowSpreadMLS               float64           `json:"flow_spread_ml_s"`
	PressureJitterBar           float64           `json:"pressure_jitter_bar"`
	ChannelingRisk              string            `json:"channeling_risk"`
	Annotations                 map[string]string `json:"annotations"`
}

type TemperatureDiagnostics struct {
	OvershootC    float64           `json:"overshoot_c"`
	UndershootC   float64           `json:"undershoot_c"`
	StabilityStdC float64           `json:"stability_std_c"`
	Annotations   map[string]string `json:"annotations"`
}

type ExtractionMetrics struct {
	PressureAUCBarS       float64           `json:"pressure_auc_bar_s"`
	PressureSlopeBrewBarS float64           `json:"pressure_slope_brew_bar_s"`
	FlowSlopeBrewMLS2     float64           `json:"flow_slope_brew_ml_s2"`
	FlowAvgBrewMLS        float64           `json:"flow_avg_brew_ml_s"`
	Annotations           map[string]string `json:"annotations"`
}

type WeightDiagnostics struct {
	RateAvgGS      *float64          `json:"rate_avg_g_s,omitempty"`
	RateStdGS      *float64          `json:"rate_std_g_s,omitempty"`
	ScaleConnected bool              `json:"scale_connected"`
	Annotations    map[string]string `json:"annotations"`
}

type ProfileCompliance struct {
	PressureRMSEBar          float64           `json:"pressure_rmse_bar"`
	FlowRMSEMLS              *float64          `json:"flow_rmse_ml_s,omitempty"`
	MaxPressureOvershootBar  float64           `json:"max_pressure_overshoot_bar"`
	MaxPressureUndershootBar float64           `json:"max_pressure_undershoot_bar"`
	MaxFlowOvershootMLS      *float64          `json:"max_flow_overshoot_ml_s,omitempty"`
	MaxFlowUndershootMLS     *float64          `json:"max_flow_undershoot_ml_s,omitempty"`
	Annotations              map[string]string `json:"annotations"`
}

// ─── Math Helpers ─────────────────────────────────────────────────

func r2(v float64) float64 { return math.Round(v*100) / 100 }

func safeMean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func safeStd(v []float64) float64 {
	if len(v) < 2 {
		return 0
	}
	m := safeMean(v)
	var var_ float64
	for _, x := range v {
		d := x - m
		var_ += d * d
	}
	return math.Sqrt(var_ / float64(len(v)))
}

func jitterStd(v []float64) float64 {
	if len(v) < 3 {
		return 0
	}
	diffs := make([]float64, len(v)-1)
	for i := 1; i < len(v); i++ {
		diffs[i-1] = v[i] - v[i-1]
	}
	return safeStd(diffs)
}

func linearSlope(v []float64, dt float64) float64 {
	n := len(v)
	if n < 2 || dt <= 0 {
		return 0
	}
	xMean := float64(n-1) / 2.0 * dt
	yMean := safeMean(v)
	var num, den float64
	for i, y := range v {
		x := float64(i) * dt
		num += (x - xMean) * (y - yMean)
		den += (x - xMean) * (x - xMean)
	}
	if den == 0 {
		return 0
	}
	return num / den
}

func computeRMse(a, t []float64) float64 {
	if len(a) == 0 || len(a) != len(t) {
		return 0
	}
	sse := 0.0
	for i := range a {
		d := a[i] - t[i]
		sse += d * d
	}
	return math.Sqrt(sse / float64(len(a)))
}

func minSlice(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v {
		if x < m {
			m = x
		}
	}
	return m
}

func maxSlice(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v {
		if x > m {
			m = x
		}
	}
	return m
}

func iMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func iMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ─── Annotation Bands ─────────────────────────────────────────────

type band struct {
	upper float64
	lower float64
	label string
}

func annotateAscending(value float64, bands []band) string {
	for _, b := range bands {
		if value < b.upper {
			return b.label
		}
	}
	return bands[len(bands)-1].label
}

func annotateDescending(value float64, bands []band) string {
	for _, b := range bands {
		if value >= b.lower {
			return b.label
		}
	}
	return bands[len(bands)-1].label
}

var (
	resistanceLevelBands    = []band{{0.5, 0, "VERY_LOW"}, {1.5, 0, "LOW"}, {3.0, 0, "MODERATE"}, {5.0, 0, "HIGH"}, {math.Inf(1), 0, "VERY_HIGH"}}
	resistanceStabilityBands = []band{{0.2, 0, "VERY_STABLE"}, {0.5, 0, "STABLE"}, {1.0, 0, "MODERATE"}, {math.Inf(1), 0, "VOLATILE"}}
	resistanceSlopeBands    = []band{{0, 0.05, "INCREASING"}, {0, -0.02, "FLAT"}, {0, -0.08, "GRADUAL_DECLINE"}, {0, -0.15, "MODERATE_DECLINE"}, {0, math.Inf(-1), "STEEP_DECLINE"}}
	resistancePeakTimingBands = []band{{0.15, 0, "EARLY"}, {0.35, 0, "GOOD_TIMING"}, {0.60, 0, "MID_SHOT"}, {math.Inf(1), 0, "LATE"}}

	flowJitterBands     = []band{{0.025, 0, "VERY_STABLE"}, {0.050, 0, "STABLE"}, {0.100, 0, "MODERATE_JITTER"}, {0.200, 0, "JITTERY"}, {math.Inf(1), 0, "VOLATILE"}}
	flowVsTargetBands   = []band{{0.15, 0, "WITHIN_TOLERANCE"}, {0.35, 0, "MINOR_DEVIATION"}, {0.70, 0, "NOTABLE_DEVIATION"}, {math.Inf(1), 0, "SEVERE_DEVIATION"}}
	pressureJitterBands = []band{{0.05, 0, "VERY_STABLE"}, {0.10, 0, "STABLE"}, {0.20, 0, "MODERATE_JITTER"}, {0.40, 0, "JITTERY"}, {math.Inf(1), 0, "VOLATILE"}}
	pressureDropBands   = []band{{0, -1.0, "NORMAL"}, {0, -2.5, "MODERATE_DROP"}, {0, -5.0, "STEEP_DROP"}, {0, math.Inf(-1), "CLIFF"}}
	flowAccelBands      = []band{{0.02, 0, "STABLE"}, {0.05, 0, "SLIGHT_ACCELERATION"}, {0.10, 0, "MODERATE_ACCELERATION"}, {math.Inf(1), 0, "RAPID_ACCELERATION"}}

	tempOvershootBands   = []band{{0.5, 0, "MINIMAL"}, {1.0, 0, "SLIGHT"}, {2.0, 0, "MODERATE"}, {math.Inf(1), 0, "SIGNIFICANT"}}
	tempStabilityBands   = []band{{0.3, 0, "VERY_STABLE"}, {0.8, 0, "STABLE"}, {1.5, 0, "MODERATE"}, {math.Inf(1), 0, "UNSTABLE"}}
	profileAdherenceBands = []band{{0.3, 0, "EXCELLENT"}, {0.8, 0, "GOOD"}, {1.5, 0, "FAIR"}, {math.Inf(1), 0, "POOR"}}
	pressureOvershootBands = []band{{0.25, 0, "WITHIN_TOLERANCE"}, {0.5, 0, "MINOR_OVERSHOOT"}, {1.0, 0, "NOTABLE_OVERSHOOT"}, {math.Inf(1), 0, "SEVERE_OVERSHOOT"}}
	flowDeviationBands   = []band{{0.3, 0, "WITHIN_TOLERANCE"}, {0.7, 0, "MINOR_DEVIATION"}, {1.5, 0, "NOTABLE_DEVIATION"}, {math.Inf(1), 0, "SEVERE_DEVIATION"}}
)

// ─── Sample Helpers ───────────────────────────────────────────────

func getFloat(s parser.Sample, key string) float64 {
	if v, ok := s[key].(float64); ok {
		return v
	}
	return 0
}

func extractFloats(samples []parser.Sample, key string) []float64 {
	r := make([]float64, 0, len(samples))
	for _, s := range samples {
		if v, ok := s[key].(float64); ok {
			r = append(r, v)
		}
	}
	return r
}

type paired struct {
	actual float64
	target float64
}

func extractPaired(samples []parser.Sample, aKey, tKey string) []paired {
	r := make([]paired, 0)
	for _, s := range samples {
		a, okA := s[aKey].(float64)
		t, okT := s[tKey].(float64)
		if okA && okT {
			r = append(r, paired{a, t})
		}
	}
	return r
}

func calculateTotalVolume(samples []parser.Sample, intervalMs int) float64 {
	total := 0.0
	dt := float64(intervalMs) / 1000.0
	for _, s := range samples {
		total += getFloat(s, "pf") * dt
	}
	return math.Round(total*10) / 10
}

// ─── Classification ───────────────────────────────────────────────

func classifyPhaseByName(name string) string {
	lower := strings.ToLower(name)
	preinfKeywords := []string{"preinfusion", "pre-infusion", "pi", "soak", "bloom", "fill", "preinfuse"}
	for _, kw := range preinfKeywords {
		if strings.Contains(lower, kw) {
			return "preinfusion"
		}
	}
	declineKeywords := []string{"decline", "taper", "ramp-down", "ramp down", "cool down", "cooldown"}
	for _, kw := range declineKeywords {
		if strings.Contains(lower, kw) {
			return "decline"
		}
	}
	return "brew"
}

func getBrewPhaseSamples(shot *parser.ShotData) []parser.Sample {
	if len(shot.Samples) == 0 {
		return nil
	}
	if len(shot.Phases) > 0 {
		brew := make([]parser.Sample, 0)
		for i, phase := range shot.Phases {
			if classifyPhaseByName(phase.PhaseName) == "preinfusion" {
				continue
			}
			start := phase.SampleIndex
			end := len(shot.Samples)
			if i+1 < len(shot.Phases) {
				end = shot.Phases[i+1].SampleIndex
			}
			brew = append(brew, shot.Samples[start:end]...)
		}
		if len(brew) > 0 {
			return brew
		}
	}
	// Fallback: skip before 50% peak pressure
	pressures := extractFloats(shot.Samples, "cp")
	peak := maxSlice(pressures)
	if peak > 0 {
		thr := peak * 0.5
		for i, p := range pressures {
			if p >= thr {
				return shot.Samples[i:]
			}
		}
	}
	return shot.Samples
}

// ─── Channeling Assessment ────────────────────────────────────────

const minSteadyStateSamples = 5

func trimRampUp(p, f []float64, thrPct float64) ([]float64, []float64) {
	if len(p) == 0 {
		return p, f
	}
	peak := maxSlice(p)
	if peak <= 0 {
		return p, f
	}
	tgt := peak * thrPct
	for i, v := range p {
		if v >= tgt {
			return p[i:], f[i:]
		}
	}
	return p, f
}

func stripFlowEdges(p, f []float64, thr float64) ([]float64, []float64) {
	n := len(f)
	i := 0
	for i < n && f[i] < thr {
		i++
	}
	if i == n {
		return nil, nil
	}
	j := n - 1
	for j > i && f[j] < thr {
		j--
	}
	return p[i : j+1], f[i : j+1]
}

func windowConfidence(n int) string {
	if n < minSteadyStateSamples {
		return "INSUFFICIENT"
	}
	if n < 8 {
		return "LOW"
	}
	if n < 15 {
		return "MEDIUM"
	}
	return "HIGH"
}

func flowShapeLabel(flows []float64, dt float64) string {
	if len(flows) < 2 {
		return "FLAT"
	}
	s := linearSlope(flows, dt)
	if s > 0.03 {
		return "RAMPING_UP"
	}
	if s < -0.03 {
		return "RAMPING_DOWN"
	}
	return "FLAT"
}

func residualStdVsTarget(samples []parser.Sample) *float64 {
	pairs := make([]float64, 0)
	for _, s := range samples {
		tf := getFloat(s, "tf")
		if tf <= 0 {
			continue
		}
		pf := getFloat(s, "pf")
		pairs = append(pairs, pf-tf)
	}
	if len(pairs) < 3 {
		return nil
	}
	v := r2(safeStd(pairs))
	return &v
}

func assessChannelingRisk(flowJitter float64, flowVsTarget *float64, pMaxDrop, fAccelLate, pJitter float64) string {
	score := 0
	if flowJitter >= 0.05 {
		score++
	}
	if flowJitter >= 0.10 {
		score++
	}
	if flowVsTarget != nil {
		if *flowVsTarget >= 0.35 {
			score++
		}
		if *flowVsTarget >= 0.70 {
			score++
		}
	} else {
		if pJitter >= 0.10 {
			score++
		}
		if pJitter >= 0.20 {
			score++
		}
	}
	if pMaxDrop <= -1.5 {
		score++
	}
	if pMaxDrop <= -3.0 {
		score++
	}
	if fAccelLate >= 0.05 {
		score++
	}
	if fAccelLate >= 0.10 {
		score++
	}
	switch {
	case score <= 1:
		return "LOW"
	case score <= 3:
		return "MODERATE"
	case score <= 5:
		return "HIGH"
	default:
		return "VERY_HIGH"
	}
}

func channelingPrimarySignal(flowJitter float64, flowVsTarget *float64, pMaxDrop, fAccelLate, pJitter float64) string {
	s := make([]string, 0)
	if flowJitter >= 0.05 {
		s = append(s, "flow_jitter")
	}
	if flowVsTarget != nil && *flowVsTarget >= 0.35 {
		s = append(s, "flow_vs_target")
	} else if flowVsTarget == nil && pJitter >= 0.10 {
		s = append(s, "pressure_jitter_fallback")
	}
	if pMaxDrop <= -1.5 {
		s = append(s, "pressure_cliff")
	}
	if fAccelLate >= 0.05 {
		s = append(s, "late_flow_runaway")
	}
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ",")
}

func lateFlowRunaway(flows []float64, dt float64) float64 {
	n := len(flows)
	if n < 6 {
		return 0
	}
	lateStart := int(float64(n) * 0.6)
	return linearSlope(flows[lateStart:], dt) - linearSlope(flows, dt)
}

// ─── Summary Diagnostics ──────────────────────────────────────────

func computeSummaryDiagnostics(shot *parser.ShotData) *SummaryDiagnostics {
	samples := shot.Samples
	if len(samples) < 5 {
		return nil
	}
	dt := float64(shot.SampleInterval) / 1000.0
	brewSamples := getBrewPhaseSamples(shot)
	if len(brewSamples) < 3 {
		return nil
	}

	brewP := extractFloats(brewSamples, "cp")
	brewF := extractFloats(brewSamples, "pf")
	brewT := extractFloats(brewSamples, "ct")

	// Resistance
	rVals := make([]float64, 0)
	for i := range brewP {
		if brewF[i] > 0.1 {
			rVals = append(rVals, brewP[i]/(brewF[i]*brewF[i]))
		}
	}
	rAvg := r2(safeMean(rVals))
	rSlope := r2(linearSlope(rVals, dt))

	// Channeling
	fj := jitterStd(brewF)
	pj := jitterStd(brewP)
	fvt := residualStdVsTarget(brewSamples)
	pDeriv := make([]float64, 0)
	for i := 1; i < len(brewP); i++ {
		pDeriv = append(pDeriv, (brewP[i]-brewP[i-1])/dt)
	}
	pDrop := 0.0
	if len(pDeriv) > 0 {
		pDrop = minSlice(pDeriv)
	}
	fAccel := lateFlowRunaway(brewF, dt)
	risk := assessChannelingRisk(fj, fvt, pDrop, fAccel, pj)

	tStd := r2(safeStd(brewT))

	// Profile compliance
	pRMSE := 0.0
	maxOvershoot := 0.0
	pPairs := extractPaired(brewSamples, "cp", "tp")
	if len(pPairs) >= 3 {
		a := make([]float64, len(pPairs))
		t := make([]float64, len(pPairs))
		for i, p := range pPairs {
			a[i] = p.actual
			t[i] = p.target
		}
		pRMSE = r2(computeRMse(a, t))
		for i := range a {
			d := a[i] - t[i]
			if d > maxOvershoot {
				maxOvershoot = d
			}
		}
	}

	var fRMSE, maxFlowOvershoot *float64
	fPairs := extractPaired(brewSamples, "pf", "tf")
	if len(fPairs) >= 3 {
		a := make([]float64, len(fPairs))
		t := make([]float64, len(fPairs))
		for i, p := range fPairs {
			a[i] = p.actual
			t[i] = p.target
		}
		v := r2(computeRMse(a, t))
		fRMSE = &v
		mf := 0.0
		for i := range a {
			d := a[i] - t[i]
			if d > mf {
				mf = d
			}
		}
		maxFlowOvershoot = &mf
	}

	hasScale := false
	for _, w := range extractFloats(brewSamples, "v") {
		if w > 0 {
			hasScale = true
			break
		}
	}

	ann := map[string]string{
		"resistance_level":     annotateAscending(rAvg, resistanceLevelBands),
		"resistance_erosion":   annotateDescending(rSlope, resistanceSlopeBands),
		"channeling_risk":      risk,
		"temperature_stability": annotateAscending(tStd, tempStabilityBands),
		"pressure_adherence":   annotateAscending(pRMSE, profileAdherenceBands),
		"pressure_overshoot":   annotateAscending(maxOvershoot, pressureOvershootBands),
	}
	if fRMSE != nil {
		ann["flow_adherence"] = annotateAscending(*fRMSE, profileAdherenceBands)
	}
	if maxFlowOvershoot != nil {
		ann["flow_overshoot"] = annotateAscending(*maxFlowOvershoot, flowDeviationBands)
	}

	return &SummaryDiagnostics{
		ResistanceAvg:         rAvg,
		ResistanceSlope:       rSlope,
		ChannelingRisk:        risk,
		TemperatureStabilityC: tStd,
		PressureRMSEBar:       r2(pRMSE),
		MaxOvershootBar:       r2(maxOvershoot),
		FlowRMSEMLS:           fRMSE,
		MaxFlowOvershootMLS:   maxFlowOvershoot,
		ScaleConnected:        hasScale,
		Annotations:           ann,
	}
}

// ─── Full Diagnostics ─────────────────────────────────────────────

func computeShotDiagnostics(shot *parser.ShotData) *ShotDiagnostics {
	samples := shot.Samples
	if len(samples) < 5 {
		return nil
	}
	dt := float64(shot.SampleInterval) / 1000.0
	brewSamples := getBrewPhaseSamples(shot)
	if len(brewSamples) < 3 {
		return nil
	}

	brewP := extractFloats(brewSamples, "cp")
	brewF := extractFloats(brewSamples, "pf")
	brewT := extractFloats(brewSamples, "ct")
	brewTT := extractFloats(brewSamples, "tt")
	allP := extractFloats(samples, "cp")
	brewW := extractFloats(brewSamples, "v")

	// Resistance
	rVals := make([]float64, 0)
	for i := range brewP {
		if brewF[i] > 0.1 {
			rVals = append(rVals, brewP[i]/(brewF[i]*brewF[i]))
		}
	}
	rAvg := r2(safeMean(rVals))
	rStd := r2(safeStd(rVals))
	rSlope := r2(linearSlope(rVals, dt))
	rPeak := 0.0
	rPeakTiming := 0.0
	if len(rVals) > 0 {
		rPeak = rVals[0]
		idx := 0
		for i, v := range rVals {
			if v > rPeak {
				rPeak = v
				idx = i
			}
		}
		rPeak = r2(rPeak)
		rPeakTiming = r2(float64(idx) / float64(len(rVals)))
	}

	resistance := ResistanceDiagnostics{
		Avg: rAvg, Std: rStd, Slope: rSlope, Peak: rPeak, PeakTimingPct: rPeakTiming,
		Annotations: map[string]string{
			"level":      annotateAscending(rAvg, resistanceLevelBands),
			"stability":  annotateAscending(rStd, resistanceStabilityBands),
			"erosion":    annotateDescending(rSlope, resistanceSlopeBands),
			"saturation": annotateAscending(rPeakTiming, resistancePeakTimingBands),
		},
	}

	// Channeling
	fj := jitterStd(brewF)
	pj := jitterStd(brewP)
	fvt := residualStdVsTarget(brewSamples)
	pDeriv := make([]float64, 0)
	for i := 1; i < len(brewP); i++ {
		pDeriv = append(pDeriv, (brewP[i]-brewP[i-1])/dt)
	}
	pDrop := 0.0
	if len(pDeriv) > 0 {
		pDrop = minSlice(pDeriv)
	}
	fAccel := lateFlowRunaway(brewF, dt)

	ssP, ssF := trimRampUp(brewP, brewF, 0.90)
	ssP, ssF = stripFlowEdges(ssP, ssF, 0.1)
	n := len(ssP)
	conf := windowConfidence(n)
	risk := assessChannelingRisk(fj, fvt, pDrop, fAccel, pj)
	primary := channelingPrimarySignal(fj, fvt, pDrop, fAccel, pj)
	fSpread := r2(safeStd(ssF))
	fShape := flowShapeLabel(ssF, dt)

	chAnn := map[string]string{
		"flow_jitter":       annotateAscending(r2(fj), flowJitterBands),
		"pressure_jitter":   annotateAscending(r2(pj), pressureJitterBands),
		"pressure_drop":     annotateDescending(r2(pDrop), pressureDropBands),
		"late_flow_trend":   annotateAscending(r2(fAccel), flowAccelBands),
		"flow_shape":        fShape,
		"window_confidence": conf,
		"primary_signal":    primary,
	}
	if fvt != nil {
		chAnn["flow_vs_target"] = annotateAscending(*fvt, flowVsTargetBands)
	} else {
		chAnn["flow_vs_target"] = "N/A"
	}

	channeling := ChannelingIndicators{
		FlowJitterMLS:            r2(fj),
		FlowVsTargetResidualMLS:  fvt,
		PressureMaxDropRateBarS:  r2(pDrop),
		FlowAccelerationLateMLS2: r2(fAccel),
		FlowSpreadMLS:            fSpread,
		PressureJitterBar:        r2(pj),
		ChannelingRisk:           risk,
		Annotations:              chAnn,
	}

	// Temperature
	tDev := make([]float64, 0)
	for i := range brewT {
		if i < len(brewTT) && brewTT[i] > 0 {
			tDev = append(tDev, brewT[i]-brewTT[i])
		}
	}
	tOvershoot := 0.0
	tUndershoot := 0.0
	if len(tDev) > 0 {
		tOvershoot = math.Max(0, maxSlice(tDev))
		tUndershoot = math.Max(0, -minSlice(tDev))
	}

	temperature := TemperatureDiagnostics{
		OvershootC:    r2(tOvershoot),
		UndershootC:   r2(tUndershoot),
		StabilityStdC: r2(safeStd(brewT)),
		Annotations: map[string]string{
			"overshoot":  annotateAscending(tOvershoot, tempOvershootBands),
			"undershoot": annotateAscending(tUndershoot, tempOvershootBands),
			"stability":  annotateAscending(r2(safeStd(brewT)), tempStabilityBands),
		},
	}

	// Extraction
	pAUC := 0.0
	for _, p := range allP {
		pAUC += p * dt
	}
	extraction := ExtractionMetrics{
		PressureAUCBarS:       r2(pAUC),
		PressureSlopeBrewBarS: r2(linearSlope(brewP, dt)),
		FlowSlopeBrewMLS2:     r2(linearSlope(brewF, dt)),
		FlowAvgBrewMLS:        r2(safeMean(brewF)),
		Annotations: map[string]string{
			"pressure_trend": annotateDescending(r2(linearSlope(brewP, dt)), resistanceSlopeBands),
			"flow_trend":     flowTrendLabel(r2(linearSlope(brewF, dt))),
		},
	}

	// Weight
	hasScale := false
	for _, w := range brewW {
		if w > 0 {
			hasScale = true
			break
		}
	}
	var wAvg, wStd *float64
	wAnn := map[string]string{}
	if hasScale {
		rates := make([]float64, 0)
		for i := 1; i < len(brewW); i++ {
			r := (brewW[i] - brewW[i-1]) / dt
			if r >= 0 {
				rates = append(rates, r)
			}
		}
		if len(rates) > 0 {
			v1 := r2(safeMean(rates))
			v2 := r2(safeStd(rates))
			wAvg = &v1
			wStd = &v2
			wAnn["rate_stability"] = annotateAscending(v2, flowJitterBands)
		}
	} else {
		wAnn["note"] = "No scale data available"
	}

	weight := WeightDiagnostics{
		RateAvgGS:      wAvg,
		RateStdGS:      wStd,
		ScaleConnected: hasScale,
		Annotations:    wAnn,
	}

	// Profile compliance
	pc := computeProfileCompliance(samples, dt)

	return &ShotDiagnostics{
		Resistance:        resistance,
		Channeling:        channeling,
		Temperature:       temperature,
		Extraction:        extraction,
		Weight:            weight,
		ProfileCompliance: pc,
	}
}

func computeProfileCompliance(samples []parser.Sample, dt float64) *ProfileCompliance {
	pPairs := extractPaired(samples, "cp", "tp")
	if len(pPairs) < 3 {
		return nil
	}
	a := make([]float64, len(pPairs))
	t := make([]float64, len(pPairs))
	for i, p := range pPairs {
		a[i] = p.actual
		t[i] = p.target
	}
	pRMSE := r2(computeRMse(a, t))

	maxO := 0.0
	maxU := 0.0
	for i := range a {
		d := a[i] - t[i]
		if d > maxO {
			maxO = d
		}
		if -d > maxU {
			maxU = -d
		}
	}

	ann := map[string]string{
		"pressure_adherence": annotateAscending(pRMSE, profileAdherenceBands),
		"pressure_overshoot": annotateAscending(maxO, pressureOvershootBands),
	}

	var fRMSE, maxFO, maxFU *float64
	fPairs := extractPaired(samples, "pf", "tf")
	if len(fPairs) >= 3 {
		fa := make([]float64, len(fPairs))
		ft := make([]float64, len(fPairs))
		for i, p := range fPairs {
			fa[i] = p.actual
			ft[i] = p.target
		}
		v := r2(computeRMse(fa, ft))
		fRMSE = &v
		fo := 0.0
		fu := 0.0
		for i := range fa {
			d := fa[i] - ft[i]
			if d > fo {
				fo = d
			}
			if -d > fu {
				fu = -d
			}
		}
		maxFO = &fo
		maxFU = &fu
		ann["flow_adherence"] = annotateAscending(*fRMSE, profileAdherenceBands)
		ann["flow_overshoot"] = annotateAscending(fo, flowDeviationBands)
		ann["flow_undershoot"] = annotateAscending(fu, flowDeviationBands)
	}

	return &ProfileCompliance{
		PressureRMSEBar:          pRMSE,
		FlowRMSEMLS:              fRMSE,
		MaxPressureOvershootBar:  r2(maxO),
		MaxPressureUndershootBar: r2(maxU),
		MaxFlowOvershootMLS:      maxFO,
		MaxFlowUndershootMLS:     maxFU,
		Annotations:              ann,
	}
}

func flowTrendLabel(slope float64) string {
	if slope < -0.02 {
		return "DECLINING"
	}
	if slope < 0.02 {
		return "STABLE"
	}
	return "INCREASING"
}

// ─── Summary Calculation ──────────────────────────────────────────

func calculateSummary(shot *parser.ShotData) ShotSummary {
	samples := shot.Samples
	if len(samples) == 0 {
		return ShotSummary{}
	}

	temps := extractFloats(samples, "ct")
	tTemps := extractFloats(samples, "tt")
	pressures := extractFloats(samples, "cp")
	flows := extractFloats(samples, "pf")
	times := make([]float64, len(samples))
	for i, s := range samples {
		times[i] = getFloat(s, "t") / 1000.0
	}

	tMin, tMax := 0.0, 0.0
	if len(temps) > 0 {
		tMin = temps[0]
		tMax = temps[0]
		for _, v := range temps {
			if v < tMin {
				tMin = v
			}
			if v > tMax {
				tMax = v
			}
		}
	}

	peakP := 0.0
	peakIdx := 0
	for i, p := range pressures {
		if p > peakP {
			peakP = p
			peakIdx = i
		}
	}
	peakTime := 0.0
	if peakIdx < len(times) {
		peakTime = times[peakIdx]
	}

	totalVol := calculateTotalVolume(samples, shot.SampleInterval)
	avgFlow := r2(safeMean(flows))
	peakFlow := r2(maxSlice(flows))

	var firstDrip *float64
	for i, f := range flows {
		if f > 0.0 && i < len(times) {
			v := r2(times[i])
			firstDrip = &v
			break
		}
	}

	preinfTime := 0.0
	if peakP > 0 {
		thr := peakP * 0.5
		for i, p := range pressures {
			if p >= thr && i < len(times) {
				preinfTime = times[i]
				break
			}
		}
	}
	totalTime := float64(shot.Duration) / 1000.0
	mainExt := math.Max(0, totalTime-preinfTime)

	return ShotSummary{
		Temperature: TemperatureSummary{
			MinC: r2(tMin), MaxC: r2(tMax), AvgC: r2(safeMean(temps)), TargetAvgC: r2(safeMean(tTemps)),
		},
		Pressure: PressureSummary{
			MinBar: r2(minSlice(pressures)), MaxBar: r2(maxSlice(pressures)),
			AvgBar: r2(safeMean(pressures)), PeakTimeS: r2(peakTime),
		},
		Flow: FlowSummary{
			TotalVolumeML: totalVol, AvgFlowMLS: avgFlow, PeakFlowMLS: peakFlow, TimeToFirstDrip: firstDrip,
		},
		Extraction: ExtractionSummary{
			PreinfusionTimeS: r2(preinfTime), MainExtractionTimeS: r2(mainExt), TotalTimeS: r2(totalTime),
		},
	}
}

// ─── Sample Selection ─────────────────────────────────────────────

func selectSamples(samples []parser.Sample) []TransformedSample {
	n := len(samples)
	if n == 0 {
		return nil
	}
	indices := make([]int, 0)
	if n <= 5 {
		for i := 0; i < n; i++ {
			indices = append(indices, i)
		}
	} else {
		seen := make(map[int]bool)
		for i := 0; i < 5; i++ {
			idx := int(float64(i)*float64(n-1)/4.0 + 0.5)
			if !seen[idx] {
				seen[idx] = true
				indices = append(indices, idx)
			}
		}
	}

	result := make([]TransformedSample, 0, len(indices))
	for _, idx := range indices {
		if idx >= len(samples) {
			continue
		}
		lo := iMax(0, idx-1)
		hi := iMin(len(samples), idx+2)
		window := samples[lo:hi]
		wn := float64(len(window))
		var sT, sC, sP, sF, sV float64
		for _, s := range window {
			sT += getFloat(s, "t")
			sC += getFloat(s, "ct")
			sP += getFloat(s, "cp")
			sF += getFloat(s, "pf")
			sV += getFloat(s, "v")
		}
		result = append(result, TransformedSample{
			TimeSeconds:  r2(getFloat(samples[idx], "t") / 1000.0),
			TemperatureC: r2(sC / wn),
			PressureBar:  r2(sP / wn),
			FlowMLS:      r2(sF / wn),
			WeightG:      r2(sV / wn),
		})
	}
	return result
}

// ─── Phase Building ───────────────────────────────────────────────

func buildPhases(shot *parser.ShotData, includeSamples, includeDiagnostics bool, dt float64) []PhaseData {
	phases := make([]PhaseData, 0)
	samples := shot.Samples
	if len(samples) == 0 {
		return phases
	}

	if len(shot.Phases) > 0 {
		for i, phase := range shot.Phases {
			startIdx := phase.SampleIndex
			endIdx := len(samples)
			if i+1 < len(shot.Phases) {
				endIdx = shot.Phases[i+1].SampleIndex
			}
			ps := samples[startIdx:endIdx]
			if len(ps) == 0 {
				continue
			}
			startTime := getFloat(ps[0], "t") / 1000.0
			endTime := getFloat(ps[len(ps)-1], "t") / 1000.0
			dur := math.Max(0, endTime-startTime+float64(shot.SampleInterval)/1000.0)

			pd := PhaseData{
				Name:             phase.PhaseName,
				PhaseNumber:      phase.PhaseNumber,
				StartTimeSeconds: r2(startTime),
				DurationSeconds:  r2(dur),
				SampleCount:      len(ps),
				AvgTemperatureC:  r2(safeMean(extractFloats(ps, "ct"))),
				AvgPressureBar:   r2(safeMean(extractFloats(ps, "cp"))),
				TotalFlowML:      calculateTotalVolume(ps, shot.SampleInterval),
			}
			if includeSamples {
				pd.Samples = selectSamples(ps)
			}
			phases = append(phases, pd)
		}
	} else {
		pd := PhaseData{
			Name:             "extraction",
			PhaseNumber:      0,
			StartTimeSeconds: 0,
			DurationSeconds:  r2(float64(shot.Duration) / 1000.0),
			SampleCount:      len(samples),
			AvgTemperatureC:  r2(safeMean(extractFloats(samples, "ct"))),
			AvgPressureBar:   r2(safeMean(extractFloats(samples, "cp"))),
			TotalFlowML:      calculateTotalVolume(samples, shot.SampleInterval),
		}
		if includeSamples {
			pd.Samples = selectSamples(samples)
		}
		phases = append(phases, pd)
	}

	return phases
}

// ─── Public API ───────────────────────────────────────────────────

// TransformShotForAI transforms raw shot data for AI analysis.
func TransformShotForAI(shot *parser.ShotData, detail string) *TransformedShot {
	if detail != DetailSummary && detail != DetailPerPhase && detail != DetailPerPhaseDetailed {
		detail = DetailSummary
	}

	summary := calculateSummary(shot)
	dt := float64(shot.SampleInterval) / 1000.0

	var phases []PhaseData
	var diagnostics interface{}

	switch detail {
	case DetailSummary:
		phases = buildPhases(shot, false, false, dt)
		diagnostics = computeSummaryDiagnostics(shot)
	case DetailPerPhase:
		phases = buildPhases(shot, false, true, dt)
		diagnostics = computeShotDiagnostics(shot)
	case DetailPerPhaseDetailed:
		phases = buildPhases(shot, true, true, dt)
		diagnostics = computeShotDiagnostics(shot)
	}

	ts := time.Unix(int64(shot.Timestamp), 0)
	durSec := r2(float64(shot.Duration) / 1000.0)

	return &TransformedShot{
		ShotID:          shot.ID,
		ProfileName:     shot.ProfileName,
		ProfileID:       shot.ProfileID,
		Timestamp:       ts,
		DurationSeconds: durSec,
		FinalWeightG:    shot.Weight,
		Summary:         summary,
		Phases:          phases,
		Diagnostics:     diagnostics,
		DetailLevel:     detail,
	}
}
