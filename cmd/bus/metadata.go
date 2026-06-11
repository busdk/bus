package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type openCLIDocument struct {
	OpenCLI  string                 `json:"opencli"`
	Info     openCLIInfo            `json:"info"`
	Commands []openCLICommand       `json:"commands,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type openCLIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

type openCLICommand struct {
	Name      string            `json:"name"`
	Summary   string            `json:"summary,omitempty"`
	Usage     string            `json:"usage,omitempty"`
	Options   []openCLIOption   `json:"options,omitempty"`
	Examples  []openCLIExample  `json:"examples,omitempty"`
	ExitCodes []openCLIExitCode `json:"exitCodes,omitempty"`
}

type openCLIOption struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases,omitempty"`
	Description string   `json:"description,omitempty"`
	ValueName   string   `json:"valueName,omitempty"`
}

type openCLIExample struct {
	Summary string `json:"summary,omitempty"`
	Command string `json:"command"`
}

type openCLIExitCode struct {
	Code        int    `json:"code"`
	Description string `json:"description,omitempty"`
}

func handleMetadataHelp(args []string, stdout io.Writer, stderr io.Writer) (bool, int) {
	if len(args) == 0 || args[0] != "help" {
		return false, 0
	}
	format := "text"
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--format" || arg == "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "help: missing value for "+arg)
				return true, 2
			}
			i++
			format = args[i]
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		case arg == "--help" || arg == "-h":
			format = "text"
		default:
			fmt.Fprintln(stderr, "help: unknown help argument: "+arg)
			return true, 2
		}
	}
	switch format {
	case "text", "":
		_, _ = io.WriteString(stdout, metadataTextHelp())
		return true, 0
	case "opencli", "json":
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(metadataDocument()); err != nil {
			fmt.Fprintln(stderr, "help: "+err.Error())
			return true, 1
		}
		return true, 0
	default:
		fmt.Fprintln(stderr, "help: unsupported format: "+format)
		return true, 2
	}
}

func metadataDocument() openCLIDocument {
	return openCLIDocument{
		OpenCLI: "0.1.0",
		Info: openCLIInfo{
			Title:       "bus",
			Version:     "dev",
			Summary:     "Bus module dispatcher.",
			Description: "Dispatches bus subcommands to module binaries without importing private module internals.",
		},
		Commands: []openCLICommand{{
			Name:    "dispatch",
			Summary: "Dispatch to a Bus module.",
			Usage:   "bus [global options] MODULE [ARGS...]",
			Options: []openCLIOption{
				{Name: "--help", Aliases: []string{"-h"}, Description: "Show help and exit."},
				{Name: "--version", Aliases: []string{"-V"}, Description: "Print version information and exit."},
				{Name: "--verbose", Aliases: []string{"-v"}, Description: "Increase diagnostics to DEBUG; repeat for TRACE."},
				{Name: "--trace", Description: "Enable TRACE diagnostics; equivalent to -vv."},
				{Name: "--quiet", Aliases: []string{"-q"}, Description: "Suppress non-ERROR diagnostics."},
				{Name: "--chdir", Aliases: []string{"-C"}, ValueName: "dir", Description: "Set effective working directory."},
				{Name: "--format", Aliases: []string{"-f"}, ValueName: "format", Description: "Select output format where supported."},
				{Name: "--perf", Description: "Emit dispatcher timing diagnostics at INFO level unless quiet mode is active."},
			},
			Examples: []openCLIExample{
				{Summary: "Show dispatcher metadata.", Command: "bus help --format opencli"},
				{Summary: "Dispatch to module help.", Command: "bus journal help --format opencli"},
			},
			ExitCodes: []openCLIExitCode{
				{Code: 0, Description: "Success."},
				{Code: 1, Description: "Runtime error."},
				{Code: 2, Description: "Usage error."},
			},
		}},
		Metadata: map[string]interface{}{
			"io.busdk.profile": map[string]interface{}{"version": "0.1", "module": "bus"},
			"io.busdk.environment": map[string]interface{}{
				"version":      "0.1",
				"sourceModule": "bus",
				"precedence":   []string{"process environment", ".env", "dispatcher defaults"},
				"dotenv":       []map[string]string{{"path": ".env", "description": "Workspace environment loaded before dispatch."}},
				"variables": []map[string]interface{}{
					envVar("BUS_BUSFILE", "Busfile path used for dispatcher batch execution."),
					envVar("BUS_PERF", "Enable dispatcher performance timing output."),
					envVar("BUS_TRANSACTION_PROVIDER", "Filesystem transaction provider for Busfile execution."),
					envVar("BUS_PREFERENCES_PATH", "Override Bus preferences path used by dispatcher workflows."),
					envVarDefault("BUS_EVENTS_API_URL", "Local Bus Events API URL supplied to child commands when unset.", "http://127.0.0.1:8081/local/v1"),
					envVarDefault("BUS_EVENTS_TOKEN_FILE", "Local Bus Events token file path supplied to child commands when unset.", ".bus/tokens/local-events.jwt"),
					envVarDefault("BUS_WORKERS_API_URL", "Local Bus Workers API URL supplied to child commands when unset.", "http://127.0.0.1:8090/local/v1"),
					envVarDefault("BUS_WORKERS_API_TOKEN_FILE", "Local Bus Workers token file path supplied to child commands when unset.", ".bus/tokens/local-events.jwt"),
				},
			},
		},
	}
}

func envVar(name string, description string) map[string]interface{} {
	return map[string]interface{}{
		"name":        name,
		"description": description,
		"schema":      map[string]interface{}{"type": "string"},
		"safeHandling": map[string]interface{}{
			"printable":     true,
			"storeInDotenv": true,
			"redactInLogs":  false,
		},
		"affects": []string{"dispatch"},
		"scope":   "workspace",
	}
}

// envVarDefault describes a dispatcher-provided default that can still be overridden.
// Used by: metadataDocument for Bus client environment defaults.
func envVarDefault(name string, description string, defaultValue string) map[string]interface{} {
	item := envVar(name, description)
	item["default"] = defaultValue
	item["source"] = "dispatcher default"
	item["safeHandling"] = map[string]interface{}{
		"printable":     true,
		"storeInDotenv": false,
		"redactInLogs":  false,
	}
	return item
}

func metadataTextHelp() string {
	return "bus exposes live dispatcher metadata.\n\nUsage:\n  bus help [--format text|opencli|json]\n"
}
