// Package profile provides profile validation and utilities.
package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Profile represents a Gaggimate brewing profile for validation purposes.
type Profile struct {
	Label       string          `json:"label"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Temperature float64         `json:"temperature"`
	Phases      json.RawMessage `json:"phases"`
}

// Phase represents a single phase in a profile.
type Phase struct {
	Name        string          `json:"name"`
	Phase       string          `json:"phase"`
	Valve       int             `json:"valve"`
	Duration    float64         `json:"duration"`
	Temperature float64         `json:"temperature"`
	Transition  Transition      `json:"transition"`
	Pump        Pump            `json:"pump"`
	Targets     []Target        `json:"targets"`
	TargetsRaw  json.RawMessage `json:"-"`
}

// Transition defines how a phase transitions from the previous phase.
type Transition struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
	Adaptive bool    `json:"adaptive"`
}

// Pump defines the pump settings for a phase.
type Pump struct {
	Target   string  `json:"target"`
	Pressure float64 `json:"pressure"`
	Flow     float64 `json:"flow"`
}

// Target defines a stop condition for a phase.
type Target struct {
	Type     string      `json:"type"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ValidationErrors collects validation errors.
type ValidationErrors struct {
	Errors []string
}

func (v *ValidationErrors) Add(format string, args ...interface{}) {
	v.Errors = append(v.Errors, fmt.Sprintf(format, args...))
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *ValidationErrors) Error() string {
	return strings.Join(v.Errors, "\n")
}

// ValidateProfile validates a profile JSON and returns any errors.
func ValidateProfile(data []byte) (*Profile, error) {
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	errs := &ValidationErrors{}

	// Validate required fields
	if strings.TrimSpace(profile.Label) == "" {
		errs.Add("label is required")
	}

	if strings.TrimSpace(profile.Type) == "" {
		errs.Add("type is required")
	} else if profile.Type != "simple" && profile.Type != "pro" {
		errs.Add("type must be 'simple' or 'pro', got '%s'", profile.Type)
	}

	if profile.Temperature < 0 || profile.Temperature > 100 {
		errs.Add("temperature must be between 0 and 100, got %.1f", profile.Temperature)
	}

	// Validate phases
	var phases []Phase
	if err := json.Unmarshal(profile.Phases, &phases); err != nil {
		errs.Add("invalid phases array: %v", err)
	} else {
		if len(phases) == 0 {
			errs.Add("at least one phase is required")
		}

		validPhases := map[string]bool{
			"preinfusion": true,
			"brew":        true,
			"decline":     true,
		}

		validTransitions := map[string]bool{
			"instant":     true,
			"linear":      true,
			"ease-in":     true,
			"ease-out":    true,
			"ease-in-out": true,
		}

		validPumpTargets := map[string]bool{
			"pressure": true,
			"flow":     true,
			"power":    true,
		}

		for i, phase := range phases {
			prefix := fmt.Sprintf("phases[%d]", i)

			if strings.TrimSpace(phase.Name) == "" {
				errs.Add("%s.name is required", prefix)
			}

			if !validPhases[phase.Phase] {
				errs.Add("%s.phase must be one of: preinfusion, brew, decline (got '%s')", prefix, phase.Phase)
			}

			if phase.Duration <= 0 {
				errs.Add("%s.duration must be > 0", prefix)
			}

			if !validTransitions[phase.Transition.Type] {
				errs.Add("%s.transition.type must be one of: instant, linear, ease-in, ease-out, ease-in-out (got '%s')", prefix, phase.Transition.Type)
			}

			if !validPumpTargets[phase.Pump.Target] {
				errs.Add("%s.pump.target must be one of: pressure, flow, power (got '%s')", prefix, phase.Pump.Target)
			}

			if phase.Pump.Target == "pressure" && phase.Pump.Pressure <= 0 {
				errs.Add("%s.pump.pressure must be > 0 for pressure target", prefix)
			}

			if phase.Pump.Target == "flow" && phase.Pump.Flow <= 0 && phase.Pump.Flow != -1 {
				errs.Add("%s.pump.flow must be > 0 (or -1 for adaptive) for flow target", prefix)
			}
		}
	}

	if errs.HasErrors() {
		return nil, errs
	}

	return &profile, nil
}

// LoadProfileFromFile loads and validates a profile from a JSON file.
// If path is "-", reads from stdin.
func LoadProfileFromFile(path string) ([]byte, error) {
	var data []byte
	var err error

	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading from stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
	}

	// Validate the JSON
	if _, err := ValidateProfile(data); err != nil {
		return nil, err
	}

	return data, nil
}
