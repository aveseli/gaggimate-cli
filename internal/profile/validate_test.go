package profile

import (
	"testing"
)

func TestValidateProfile_Valid(t *testing.T) {
	validJSON := `{
		"label": "9 Bar Pre-infusion [AI]",
		"type": "pro",
		"description": "Classic 9-bar with gentle pre-infusion.",
		"temperature": 92,
		"phases": [
			{
				"name": "Pre-infusion",
				"phase": "preinfusion",
				"valve": 1,
				"duration": 5,
				"temperature": 0,
				"transition": {"type": "linear", "duration": 5, "adaptive": true},
				"pump": {"target": "pressure", "pressure": 4, "flow": 0},
				"targets": []
			},
			{
				"name": "Hold",
				"phase": "brew",
				"valve": 1,
				"duration": 25,
				"temperature": 0,
				"transition": {"type": "instant", "duration": 0, "adaptive": true},
				"pump": {"target": "pressure", "pressure": 9, "flow": 0},
				"targets": [{"type": "volumetric", "operator": "gte", "value": 36}]
			}
		]
	}`

	profile, err := ValidateProfile([]byte(validJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Label != "9 Bar Pre-infusion [AI]" {
		t.Errorf("expected label '9 Bar Pre-infusion [AI]', got '%s'", profile.Label)
	}
	if profile.Type != "pro" {
		t.Errorf("expected type 'pro', got '%s'", profile.Type)
	}
	if profile.Temperature != 92 {
		t.Errorf("expected temperature 92, got %.1f", profile.Temperature)
	}
}

func TestValidateProfile_InvalidJSON(t *testing.T) {
	_, err := ValidateProfile([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateProfile_MissingLabel(t *testing.T) {
	json := `{"type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "pressure", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for missing label")
	}
}

func TestValidateProfile_InvalidType(t *testing.T) {
	json := `{"label": "Test", "type": "invalid", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "pressure", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestValidateProfile_InvalidPhase(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "invalid", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "pressure", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for invalid phase type")
	}
}

func TestValidateProfile_EmptyPhases(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": []}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for empty phases")
	}
}

func TestValidateProfile_NoPhases(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for missing phases")
	}
}

func TestValidateProfile_InvalidTransition(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "invalid", "duration": 0, "adaptive": true}, "pump": {"target": "pressure", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for invalid transition type")
	}
}

func TestValidateProfile_InvalidPumpTarget(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "invalid", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for invalid pump target")
	}
}

func TestValidateProfile_ZeroDuration(t *testing.T) {
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 0, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "pressure", "pressure": 9, "flow": 0}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestValidateProfile_FlowModeAdaptive(t *testing.T) {
	// Flow mode with -1 means adaptive (auto-adjust to puck resistance)
	json := `{"label": "Test", "type": "pro", "temperature": 93, "phases": [{"name": "Test", "phase": "brew", "valve": 1, "duration": 10, "temperature": 0, "transition": {"type": "instant", "duration": 0, "adaptive": true}, "pump": {"target": "flow", "pressure": 9, "flow": -1}, "targets": []}]}`
	_, err := ValidateProfile([]byte(json))
	if err != nil {
		t.Fatalf("unexpected error for adaptive flow: %v", err)
	}
}

func TestValidateProfile_MultipleErrors(t *testing.T) {
	json := `{"type": "pro", "temperature": 93, "phases": []}`
	_, err := ValidateProfile([]byte(json))
	if err == nil {
		t.Fatal("expected error")
	}
	// Should have multiple errors (missing label + empty phases)
	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error message")
	}
}
