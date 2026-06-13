package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type QuotaBucket struct {
	BucketID          string  `json:"bucketId"`
	DisplayName       string  `json:"displayName"`
	Description       string  `json:"description"`
	Window            string  `json:"window"`
	RemainingFraction float64 `json:"remainingFraction"`
	ResetTime         string  `json:"resetTime"`
}

type QuotaGroup struct {
	DisplayName string        `json:"displayName"`
	Description string        `json:"description"`
	Buckets     []QuotaBucket `json:"buckets"`
}

type QuotaResponse struct {
	Response struct {
		Groups      []QuotaGroup `json:"groups"`
		Description string       `json:"description"`
	} `json:"response"`
}

func main() {
	addressFlag := flag.String("address", "", "Antigravity language server address (e.g., localhost:41475)")
	tokenFlag := flag.String("token", "", "Antigravity CSRF token")
	jsonFlag := flag.Bool("json", false, "Output raw JSON response")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf not provided via flags, address and token will be read from the ANTIGRAVITY_LS_ADDRESS and ANTIGRAVITY_CSRF_TOKEN environment variables respectively.\n")
	}
	flag.Parse()

	address := *addressFlag
	token := *tokenFlag

	if address == "" {
		address = os.Getenv("ANTIGRAVITY_LS_ADDRESS")
	}
	if token == "" {
		token = os.Getenv("ANTIGRAVITY_CSRF_TOKEN")
	}

	if address == "" || token == "" {
		detectedAddr, detectedToken, err := autodetect()
		if err == nil {
			if address == "" {
				address = detectedAddr
			}
			if token == "" {
				token = detectedToken
			}
		} else {
			if address == "" {
				fmt.Fprintf(os.Stderr, "Error: Antigravity language server address must be specified via --address, ANTIGRAVITY_LS_ADDRESS env, or autodetected (autodetect failed: %v).\n", err)
				os.Exit(1)
			}
			if token == "" {
				fmt.Fprintf(os.Stderr, "Error: Antigravity CSRF token must be specified via --token, ANTIGRAVITY_CSRF_TOKEN env, or autodetected (autodetect failed: %v).\n", err)
				os.Exit(1)
			}
		}
	}

	// Ensure address has http/https prefix or construct url
	url := address
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	url = strings.TrimSuffix(url, "/") + "/exa.language_server_pb.LanguageServerService/RetrieveUserQuotaSummary"

	// Make request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-codeium-csrf-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Antigravity: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: Antigravity returned status %s\n", resp.Status)
		if len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "Response: %s\n", string(bodyBytes))
		}
		os.Exit(1)
	}

	if *jsonFlag {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(bodyBytes))
		}
		return
	}

	var quotaResp QuotaResponse
	if err := json.Unmarshal(bodyBytes, &quotaResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON response: %v\n", err)
		os.Exit(1)
	}

	// Print formatted summary
	if len(quotaResp.Response.Groups) == 0 {
		fmt.Println("No quota groups found.")
		return
	}

	for _, group := range quotaResp.Response.Groups {
		fmt.Printf("=== %s ===\n", group.DisplayName)
		if group.Description != "" {
			fmt.Printf("Description: %s\n", group.Description)
		}
		fmt.Println()

		for _, bucket := range group.Buckets {
			fmt.Printf("  • %s\n", bucket.DisplayName)
			fmt.Printf("    Remaining:   %.1f%%\n", bucket.RemainingFraction*100)
			if bucket.Description != "" {
				fmt.Printf("    Info:        %s\n", bucket.Description)
			}
			if bucket.ResetTime != "" {
				fmt.Printf("    Reset Time:  %s\n", bucket.ResetTime)
			}
			fmt.Println()
		}
	}

	if quotaResp.Response.Description != "" {
		fmt.Printf("Note: %s\n", quotaResp.Response.Description)
	}
}

func autodetect() (string, string, error) {
	// 1. Try to find the CSRF token from running processes by scanning /proc
	token := ""
	files, err := os.ReadDir("/proc")
	if err == nil {
		for _, f := range files {
			if !f.IsDir() {
				continue
			}
			pid := f.Name()
			isNumeric := true
			for _, r := range pid {
				if r < '0' || r > '9' {
					isNumeric = false
					break
				}
			}
			if !isNumeric {
				continue
			}

			cmdlinePath := filepath.Join("/proc", pid, "cmdline")
			content, err := os.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}

			args := strings.Split(string(content), "\x00")
			isLanguageServer := false
			for _, arg := range args {
				if strings.Contains(arg, "language_server") {
					isLanguageServer = true
					break
				}
			}

			if isLanguageServer {
				for i, arg := range args {
					if arg == "--csrf_token" && i+1 < len(args) {
						token = args[i+1]
						break
					}
				}
			}
			if token != "" {
				break
			}
		}
	}

	// 2. Try to find the HTTP port from the language server log file
	port := ""
	var logPath string
	configDir, err := os.UserConfigDir()
	if err == nil {
		logPath = filepath.Join(configDir, "Antigravity", "logs", "language_server.log")
	} else {
		if u, err := user.Current(); err == nil {
			logPath = filepath.Join(u.HomeDir, ".config", "Antigravity", "logs", "language_server.log")
		}
	}

	if logPath != "" {
		logContent, err := os.ReadFile(logPath)
		if err == nil {
			lines := strings.Split(string(logContent), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				line := lines[i]
				if strings.Contains(line, "Language server listening on random port at") && strings.Contains(line, "for HTTP") {
					parts := strings.Split(line, "random port at ")
					if len(parts) > 1 {
						portParts := strings.Split(parts[1], " ")
						if len(portParts) > 0 {
							port = portParts[0]
							break
						}
					}
				}
			}
		}
	}

	if token == "" && port == "" {
		return "", "", fmt.Errorf("could not find language server process or log file")
	}
	if token == "" {
		return "", "", fmt.Errorf("found port %s, but could not find CSRF token from running processes", port)
	}
	if port == "" {
		return "", "", fmt.Errorf("found CSRF token, but could not find port from logs")
	}

	return "localhost:" + port, token, nil
}
