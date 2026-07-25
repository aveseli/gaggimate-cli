// gaggimate-cli: CLI tool for interacting with a Gaggimate espresso machine.
//
// Provides shot history browsing, shot analysis with diagnostics, profile
// management, and shot notes — all from the command line. Designed for
// agents and humans alike.
//
// Environment variables:
//
//	GAGGIMATE_HOST       Device hostname or IP (default: gaggimate.local)
//	GAGGIMATE_PROTOCOL   ws or wss (default: ws)
//	GAGGIMATE_TIMEOUT    Request timeout in seconds (default: 15)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adnan/gaggimate-cli/internal/api"
	"github.com/adnan/gaggimate-cli/internal/diag"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	host := envOrDefault("GAGGIMATE_HOST", "gaggimate.local")
	protocol := envOrDefault("GAGGIMATE_PROTOCOL", "ws")
	useHTTPS := protocol == "wss"
	timeoutSec := envOrDefaultInt("GAGGIMATE_TIMEOUT", 15)

	httpClient := api.NewHTTPClient(host, useHTTPS)
	httpClient.Timeout = time.Duration(timeoutSec) * time.Second
	wsClient := api.NewWebSocketClient(host, useHTTPS)
	wsClient.Timeout = time.Duration(timeoutSec) * time.Second

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "shots":
		cmdShots(httpClient, wsClient, args)
	case "profiles":
		cmdProfiles(wsClient, args)
	case "notes":
		cmdNotes(wsClient, args)
	case "diagnose":
		cmdDiagnose(httpClient, host)
	case "version", "--version", "-v":
		fmt.Printf("gaggimate-cli %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// ─── Usage ────────────────────────────────────────────────────────

func printUsage() {
	fmt.Print(`gaggimate-cli — interact with your Gaggimate espresso machine

USAGE
    gaggimate-cli <command> [subcommand] [flags] [args]

COMMANDS
    shots list [--limit N]                 List recent shots
    shots analyze <ID> [--detail LEVEL]    Analyze shot with diagnostics
    shots get <ID>                         Get raw shot data (JSON)
    profiles list                          List all profiles
    profiles get <ID>                      Get a profile by ID
    profiles select <ID>                   Select active profile
    profiles delete <ID>                   Delete an AI-created profile
    notes get <SHOT_ID>                    Get shot notes
    notes set <SHOT_ID> [flags]            Set shot notes
    diagnose                               Check device connectivity
    version                                Print version

ANALYSIS DETAIL LEVELS (--detail)
    summary            Key indicators only (default)
    per_phase          Full diagnostics + per-phase breakdown
    per_phase_detailed Everything + ~5 samples per phase

SHOT ANALYSIS FIELDS
    The analyze command computes:
      • Puck resistance (P/F²) — master diagnostic metric
      • Channeling risk — 4 independent indicators scored 0-8
      • Temperature stability vs target
      • Profile compliance — RMSE vs target pressure/flow
      • Per-phase breakdown (at per_phase level)

    Each numeric metric has a band annotation (e.g., MODERATE, STABLE)
    so you can interpret values without knowing thresholds.

NOTES FLAGS
    --rating N         Star rating 0-5
    --notes TEXT        Tasting notes
    --balance TEXT      Taste: bitter, balanced, or sour
    --grind TEXT        Grinder setting
    --dose-in G         Coffee dose in grams
    --dose-out G        Espresso output in grams

ENVIRONMENT
    GAGGIMATE_HOST       Device hostname (default: gaggimate.local)
    GAGGIMATE_PROTOCOL   ws or wss (default: ws)
    GAGGIMATE_TIMEOUT    Timeout in seconds (default: 15)

EXAMPLES
    gaggimate-cli shots list
    gaggimate-cli shots analyze 196
    gaggimate-cli shots analyze 196 --detail per_phase
    gaggimate-cli profiles list
    gaggimate-cli profiles select abc-def-123
    gaggimate-cli notes get 196
    gaggimate-cli notes set 196 --rating 4 --notes "sweet, balanced" --balance balanced
    gaggimate-cli diagnose
`)
}

// ─── Shots Commands ───────────────────────────────────────────────

func cmdShots(http *api.HTTPClient, ws *api.WebSocketClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli shots <list|analyze|get> [args]")
		os.Exit(1)
	}

	sub := args[0]
	switch sub {
	case "list":
		cmdShotsList(http, args[1:])
	case "analyze":
		cmdShotsAnalyze(http, ws, args[1:])
	case "get":
		cmdShotsGet(http, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown shots subcommand: %s\n", sub)
		os.Exit(1)
	}
}

func cmdShotsList(http *api.HTTPClient, args []string) {
	limit := 10
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--limit" {
			if v, err := strconv.Atoi(args[i+1]); err == nil {
				limit = v
			}
			i++
		}
	}

	shots, err := http.FetchShotIndex()
	if err != nil {
		fatal("fetching shots: %v", err)
	}

	if limit > len(shots) {
		limit = len(shots)
	}
	shots = shots[:limit]

	printJSON(shots)
}

func cmdShotsAnalyze(http *api.HTTPClient, ws *api.WebSocketClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli shots analyze <SHOT_ID> [--detail summary|per_phase|per_phase_detailed]")
		os.Exit(1)
	}

	shotID := args[0]
	detail := diag.DetailSummary
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "--detail" {
			detail = args[i+1]
			i++
		}
	}

	shot, err := http.FetchShot(shotID)
	if err != nil {
		fatal("fetching shot %s: %v", shotID, err)
	}

	result := diag.TransformShotForAI(shot, detail)

	// Try to get notes from the device
	normalizedID := normalizeShotID(shotID)
	notes, err := ws.GetShotNotes(normalizedID)
	if err == nil && notes != nil {
		type annotatedResult struct {
			*diag.TransformedShot
			Notes *api.ShotNotes `json:"notes,omitempty"`
		}
		ar := annotatedResult{TransformedShot: result, Notes: notes}
		printJSON(ar)
	} else {
		printJSON(result)
	}
}

func cmdShotsGet(http *api.HTTPClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli shots get <SHOT_ID>")
		os.Exit(1)
	}

	shot, err := http.FetchShot(args[0])
	if err != nil {
		fatal("fetching shot %s: %v", args[0], err)
	}
	printJSON(shot)
}

// ─── Profiles Commands ────────────────────────────────────────────

func cmdProfiles(ws *api.WebSocketClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli profiles <list|get|select|delete> [args]")
		os.Exit(1)
	}

	sub := args[0]
	switch sub {
	case "list":
		profiles, err := ws.ListProfiles()
		if err != nil {
			fatal("listing profiles: %v", err)
		}
		printJSON(profiles)

	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli profiles get <PROFILE_ID>")
			os.Exit(1)
		}
		profile, err := ws.GetProfile(args[1])
		if err != nil {
			fatal("getting profile %s: %v", args[1], err)
		}
		printJSON(profile)

	case "select":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli profiles select <PROFILE_ID>")
			os.Exit(1)
		}
		if err := ws.SelectProfile(args[1]); err != nil {
			fatal("selecting profile: %v", err)
		}
		fmt.Printf("Profile %s selected\n", args[1])

	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli profiles delete <PROFILE_ID>")
			os.Exit(1)
		}
		// Safety: verify it's an AI profile
		profile, err := ws.GetProfile(args[1])
		if err != nil {
			fatal("getting profile: %v", err)
		}
		if !strings.HasSuffix(profile.Label, " [AI]") {
			fatal("cannot delete profile '%s': only AI-created profiles (ending with ' [AI]') can be deleted", profile.Label)
		}
		if err := ws.DeleteProfile(args[1]); err != nil {
			fatal("deleting profile: %v", err)
		}
		fmt.Printf("Profile '%s' deleted\n", profile.Label)

	default:
		fmt.Fprintf(os.Stderr, "Unknown profiles subcommand: %s\n", sub)
		os.Exit(1)
	}
}

// ─── Notes Commands ───────────────────────────────────────────────

func cmdNotes(ws *api.WebSocketClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli notes <get|set> [args]")
		os.Exit(1)
	}

	sub := args[0]
	switch sub {
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli notes get <SHOT_ID>")
			os.Exit(1)
		}
		notes, err := ws.GetShotNotes(args[1])
		if err != nil {
			fatal("getting notes: %v", err)
		}
		if notes == nil {
			fmt.Println("No notes for this shot")
		} else {
			printJSON(notes)
		}

	case "set":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gaggimate-cli notes set <SHOT_ID> [--rating N] [--notes TEXT] ...")
			os.Exit(1)
		}
		shotID := args[1]
		notes := api.ShotNotes{}

		for i := 2; i < len(args)-1; i++ {
			switch args[i] {
			case "--rating":
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					notes.Rating = &v
				}
				i++
			case "--notes":
				notes.Notes = args[i+1]
				i++
			case "--balance":
				notes.BalanceTaste = args[i+1]
				i++
			case "--grind":
				notes.GrindSetting = args[i+1]
				i++
			case "--dose-in":
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					notes.DoseIn = &v
				}
				i++
			case "--dose-out":
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					notes.DoseOut = &v
				}
				i++
			}
		}

		if err := ws.SaveShotNotes(shotID, notes); err != nil {
			fatal("saving notes: %v", err)
		}
		fmt.Println("Notes saved")

	default:
		fmt.Fprintf(os.Stderr, "Unknown notes subcommand: %s\n", sub)
		os.Exit(1)
	}
}

// ─── Diagnose ─────────────────────────────────────────────────────

func cmdDiagnose(http *api.HTTPClient, host string) {
	fmt.Printf("Diagnosing connection to %s...\n\n", host)

	// Test 1: Fetch shot index
	fmt.Print("  HTTP API (/api/history/index.bin)... ")
	start := time.Now()
	shots, err := http.FetchShotIndex()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("FAIL (%v)\n", elapsed.Round(time.Millisecond))
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("OK (%d shots, %v)\n", len(shots), elapsed.Round(time.Millisecond))
	}

	// Test 2: If we have shots, try fetching one
	if len(shots) > 0 {
		latestID := shots[0].ID
		fmt.Printf("  Fetching shot %s... ", latestID)
		start = time.Now()
		_, err := http.FetchShot(latestID)
		elapsed = time.Since(start)
		if err != nil {
			fmt.Printf("FAIL (%v)\n", elapsed.Round(time.Millisecond))
			fmt.Printf("    Error: %v\n", err)
		} else {
			fmt.Printf("OK (%v)\n", elapsed.Round(time.Millisecond))
		}
	}

	fmt.Println()
}

// ─── Helpers ──────────────────────────────────────────────────────

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fatal("encoding JSON: %v", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrDefaultInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func normalizeShotID(id string) string {
	n := 0
	for _, c := range id {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return strconv.Itoa(n)
}
