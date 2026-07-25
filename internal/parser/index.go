// Package parser implements binary parsers for Gaggimate shot data files.
// Mirrors the Python parsers in gaggimate-mcp/src/gaggimate_mcp/parsers/.
package parser

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// Index file constants (must match shot_log_format.h)
const (
	indexMagic      uint32 = 0x58444953 // 'SIDX'
	indexHeaderSize        = 32
	indexEntrySize         = 128
	weightScale            = 10.0
)

// Index entry flags
const (
	shotFlagCompleted = 0x01
	shotFlagDeleted   = 0x02
	shotFlagHasNotes  = 0x04
)

// IndexHeader represents the header of an index.bin file.
type IndexHeader struct {
	Magic      uint32
	Version    uint16
	EntrySize  uint16
	EntryCount uint32
	NextID     uint32
}

// IndexEntry represents a single shot entry in the index.
type IndexEntry struct {
	ID          int
	Timestamp   uint32
	Duration    uint32
	Volume      *float64 // nil if not available
	Rating      int
	Flags       uint8
	ProfileID   string
	ProfileName string
	Completed   bool
	Deleted     bool
	HasNotes    bool
	Incomplete  bool
}

// IndexData represents the complete parsed index.
type IndexData struct {
	Header  IndexHeader
	Entries []IndexEntry
}

// decodeCString extracts a null-terminated C string from bytes.
func decodeCString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

// ParseBinaryIndex parses a binary index.bin file.
func ParseBinaryIndex(data []byte) (*IndexData, error) {
	if len(data) < indexHeaderSize {
		return nil, errors.New("index file too small")
	}

	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != indexMagic {
		return nil, fmt.Errorf("invalid index magic: 0x%08x (expected 0x%08x)", magic, indexMagic)
	}

	version := binary.LittleEndian.Uint16(data[4:6])
	entrySize := binary.LittleEndian.Uint16(data[6:8])
	entryCount := binary.LittleEndian.Uint32(data[8:12])
	nextID := binary.LittleEndian.Uint32(data[12:16])

	if entrySize != indexEntrySize {
		return nil, fmt.Errorf("unsupported entry size %d (expected %d)", entrySize, indexEntrySize)
	}

	expectedSize := indexHeaderSize + int(entryCount)*indexEntrySize
	if len(data) < expectedSize {
		return nil, fmt.Errorf("index file truncated: %d bytes (expected %d)", len(data), expectedSize)
	}

	header := IndexHeader{
		Magic:      magic,
		Version:    version,
		EntrySize:  entrySize,
		EntryCount: entryCount,
		NextID:     nextID,
	}

	entries := make([]IndexEntry, 0, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		base := indexHeaderSize + int(i)*indexEntrySize

		entryID := int(binary.LittleEndian.Uint32(data[base+0 : base+4]))
		timestamp := binary.LittleEndian.Uint32(data[base+4 : base+8])
		duration := binary.LittleEndian.Uint32(data[base+8 : base+12])
		volumeRaw := binary.LittleEndian.Uint16(data[base+12 : base+14])
		rating := int(data[base+14])
		flags := data[base+15]

		profileID := decodeCString(data[base+16 : base+48])
		profileName := decodeCString(data[base+48 : base+96])

		var volume *float64
		if volumeRaw > 0 {
			v := float64(volumeRaw) / weightScale
			volume = &v
		}

		entries = append(entries, IndexEntry{
			ID:          entryID,
			Timestamp:   timestamp,
			Duration:    duration,
			Volume:      volume,
			Rating:      rating,
			Flags:       flags,
			ProfileID:   profileID,
			ProfileName: profileName,
			Completed:   flags&shotFlagCompleted != 0,
			Deleted:     flags&shotFlagDeleted != 0,
			HasNotes:    flags&shotFlagHasNotes != 0,
			Incomplete:  flags&shotFlagCompleted == 0,
		})
	}

	return &IndexData{Header: header, Entries: entries}, nil
}

// ShotListItem is a filtered, sorted shot from the index.
type ShotListItem struct {
	ID          string    `json:"id"`
	Profile     string    `json:"profile"`
	ProfileID   string    `json:"profile_id"`
	Timestamp   time.Time `json:"timestamp"`
	DurationSec float64   `json:"duration_seconds"`
	Volume      *float64  `json:"volume,omitempty"`
	Rating      *int      `json:"rating,omitempty"`
	Incomplete  bool      `json:"incomplete"`
	HasNotes    bool      `json:"has_notes"`
}

// IndexToShotList filters deleted entries and converts to a sorted shot list.
func IndexToShotList(indexData *IndexData) []ShotListItem {
	shots := make([]ShotListItem, 0)
	for _, entry := range indexData.Entries {
		if entry.Deleted {
			continue
		}

		item := ShotListItem{
			ID:          fmt.Sprintf("%d", entry.ID),
			Profile:     entry.ProfileName,
			ProfileID:   entry.ProfileID,
			Timestamp:   time.Unix(int64(entry.Timestamp), 0),
			DurationSec: float64(entry.Duration) / 1000.0,
			Volume:      entry.Volume,
			Incomplete:  entry.Incomplete,
			HasNotes:    entry.HasNotes,
		}

		if entry.Rating > 0 {
			r := entry.Rating
			item.Rating = &r
		}

		shots = append(shots, item)
	}

	// Sort by timestamp descending (newest first)
	for i := 0; i < len(shots); i++ {
		for j := i + 1; j < len(shots); j++ {
			if shots[j].Timestamp.After(shots[i].Timestamp) {
				shots[i], shots[j] = shots[j], shots[i]
			}
		}
	}

	return shots
}

// countSetBits counts the number of set bits in an integer.
func countSetBits(n uint32) int {
	count := 0
	for n != 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

// fieldBit positions (must match shot_log_format.h)
const (
	fieldBitT  = 0  // tick
	fieldBitTT = 1  // target temp
	fieldBitCT = 2  // current temp
	fieldBitTP = 3  // target pressure
	fieldBitCP = 4  // current pressure
	fieldBitFL = 5  // pump flow
	fieldBitTF = 6  // target flow
	fieldBitPF = 7  // puck flow
	fieldBitVF = 8  // volumetric flow
	fieldBitV  = 9  // volumetric weight
	fieldBitEV = 10 // estimated weight
	fieldBitPR = 11 // puck resistance
	fieldBitSI = 12 // system info (v2+)
)

// Scaling factors
const (
	tempScale      = 10.0
	pressureScale  = 10.0
	flowScale      = 100.0
	shotWeightScale = 10.0
	resistanceScale = 100.0
)

// Shot header sizes
const (
	shotHeaderSizeV4 = 128
	shotHeaderSizeV5 = 512
	shotMagic        uint32 = 0x544F4853 // 'SHOT'
)

// PhaseTransition represents a phase boundary in the shot.
type PhaseTransition struct {
	SampleIndex int
	PhaseNumber int
	PhaseName   string
}

// ShotData represents a parsed shot.
type ShotData struct {
	ID              string
	Version         int
	FieldsMask      uint32
	SampleCount     int
	SampleInterval  int // milliseconds
	ProfileID       string
	ProfileName     string
	Timestamp       uint32
	Duration        uint32 // milliseconds
	Weight          *float64
	Samples         []Sample
	Phases          []PhaseTransition
	Incomplete      bool
}

// Sample returns a map of field values for a single sample.
type Sample map[string]interface{}

// IsHTMLResponse checks if data looks like HTML instead of binary shot data.
func IsHTMLResponse(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// Check for common HTML prefixes
	prefix := string(data[:min(15, len(data))])
	lower := toLower(prefix)
	return len(lower) >= 5 && (lower[:5] == "<!doc" || lower[:5] == "<html")
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseBinaryShot parses a binary .slog shot file.
func ParseBinaryShot(data []byte, shotID string) (*ShotData, error) {
	if IsHTMLResponse(data) {
		return nil, errors.New("device returned HTML instead of binary shot data (device may be overloaded)")
	}

	if len(data) < shotHeaderSizeV4 {
		return nil, fmt.Errorf("shot file too small: %d bytes", len(data))
	}

	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != shotMagic {
		return nil, fmt.Errorf("invalid shot magic: 0x%08x (expected 0x%08x)", magic, shotMagic)
	}

	version := int(data[4])
	headerSize := shotHeaderSizeV4
	if version >= 5 {
		headerSize = shotHeaderSizeV5
	}

	if len(data) < headerSize {
		return nil, fmt.Errorf("shot file too small for version %d: %d bytes", version, len(data))
	}

	sampleInterval := int(binary.LittleEndian.Uint16(data[8:10]))
	fieldsMask := binary.LittleEndian.Uint32(data[12:16])
	sampleCount := int(binary.LittleEndian.Uint32(data[16:20]))
	duration := binary.LittleEndian.Uint32(data[20:24])
	timestamp := binary.LittleEndian.Uint32(data[24:28])

	profileID := decodeCString(data[28:60])
	profileName := decodeCString(data[60:108])
	weightRaw := binary.LittleEndian.Uint16(data[108:110])

	var weight *float64
	if weightRaw > 0 {
		w := float64(weightRaw) / shotWeightScale
		weight = &w
	}

	// Parse phase transitions (V5+)
	var phases []PhaseTransition
	if version >= 5 {
		transitionCount := int(data[458])
		baseOffset := 110

		for i := 0; i < min(transitionCount, 12); i++ {
			offset := baseOffset + i*29
			sampleIdx := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
			phaseNum := int(data[offset+2])
			phaseName := decodeCString(data[offset+4 : offset+29])

			phases = append(phases, PhaseTransition{
				SampleIndex: sampleIdx,
				PhaseNumber: phaseNum,
				PhaseName:   phaseName,
			})
		}
	}

	// Calculate sample data size
	fieldsPerSample := countSetBits(fieldsMask)
	if fieldsPerSample == 0 {
		return nil, errors.New("invalid shot file: fields_mask is 0 (no fields recorded)")
	}
	sampleDataSize := fieldsPerSample * 2 // Each field is 2 bytes

	// Handle truncated files
	availableData := len(data) - headerSize
	actualSampleCount := min(sampleCount, availableData/sampleDataSize)
	incomplete := actualSampleCount < sampleCount

	// Build field order based on mask
	fieldOrder := make([]int, 0)
	for bit := 0; bit < 13; bit++ {
		if fieldsMask&(1<<uint(bit)) != 0 {
			fieldOrder = append(fieldOrder, bit)
		}
	}

	// Parse samples
	samples := make([]Sample, 0, actualSampleCount)
	for i := 0; i < actualSampleCount; i++ {
		sample := make(Sample)
		sampleOffset := headerSize + i*sampleDataSize

		// Add phase number based on transitions
		for p := len(phases) - 1; p >= 0; p-- {
			if i >= phases[p].SampleIndex {
				sample["phase"] = phases[p].PhaseNumber
				break
			}
		}

		// Parse each field
		for fieldIndex, fieldBit := range fieldOrder {
			fieldOffset := sampleOffset + fieldIndex*2
			if fieldOffset+2 > len(data) {
				break
			}

			switch fieldBit {
			case fieldBitT:
				// tick: uint16 * sample_interval
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["t"] = float64(raw) * float64(sampleInterval)
			case fieldBitTT:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["tt"] = float64(raw) / tempScale
			case fieldBitCT:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["ct"] = float64(raw) / tempScale
			case fieldBitTP:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["tp"] = float64(raw) / pressureScale
			case fieldBitCP:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["cp"] = float64(raw) / pressureScale
			case fieldBitFL:
				raw := int16(binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2]))
				sample["fl"] = float64(raw) / flowScale
			case fieldBitTF:
				raw := int16(binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2]))
				sample["tf"] = float64(raw) / flowScale
			case fieldBitPF:
				raw := int16(binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2]))
				sample["pf"] = float64(raw) / flowScale
			case fieldBitVF:
				raw := int16(binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2]))
				sample["vf"] = float64(raw) / flowScale
			case fieldBitV:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["v"] = float64(raw) / shotWeightScale
			case fieldBitEV:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["ev"] = float64(raw) / shotWeightScale
			case fieldBitPR:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["pr"] = float64(raw) / resistanceScale
			case fieldBitSI:
				raw := binary.LittleEndian.Uint16(data[fieldOffset : fieldOffset+2])
				sample["systemInfo"] = map[string]interface{}{
					"raw":                          raw,
					"shot_started_volumetric":      raw&0x0001 != 0,
					"currently_volumetric":         raw&0x0002 != 0,
					"bluetooth_scale_connected":    raw&0x0004 != 0,
					"volumetric_available":         raw&0x0008 != 0,
					"extended_recording":           raw&0x0010 != 0,
				}
			}
		}

		samples = append(samples, sample)
	}

	return &ShotData{
		ID:             shotID,
		Version:        version,
		FieldsMask:     fieldsMask,
		SampleCount:    actualSampleCount,
		SampleInterval: sampleInterval,
		ProfileID:      profileID,
		ProfileName:    profileName,
		Timestamp:      timestamp,
		Duration:       duration,
		Weight:         weight,
		Samples:        samples,
		Phases:         phases,
		Incomplete:     incomplete,
	}, nil
}

// safeMean returns the mean of a float64 slice, or 0 if empty.
func safeMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// safeStd returns the population standard deviation, or 0 if < 2 values.
func safeStd(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := safeMean(values)
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

// linearSlope calculates the linear regression slope (units per second).
func linearSlope(values []float64, dt float64) float64 {
	n := len(values)
	if n < 2 || dt <= 0 {
		return 0
	}
	xMean := float64(n-1) / 2.0
	yMean := safeMean(values)
	numerator := 0.0
	denominator := 0.0
	for i, y := range values {
		x := float64(i) * dt
		numerator += (x - xMean*dt) * (y - yMean)
		denominator += (x - xMean*dt) * (x - xMean*dt)
	}
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// jitterStd computes the population std of first-differences (noise measure).
func jitterStd(values []float64) float64 {
	if len(values) < 3 {
		return 0
	}
	diffs := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		diffs[i-1] = values[i] - values[i-1]
	}
	return safeStd(diffs)
}

// computeRMse computes root mean square error between actual and target.
func computeRMse(actual, target []float64) float64 {
	if len(actual) == 0 || len(actual) != len(target) {
		return 0
	}
	sse := 0.0
	for i := range actual {
		d := actual[i] - target[i]
		sse += d * d
	}
	return math.Sqrt(sse / float64(len(actual)))
}
