package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Application struct {
	Output     string
	Target     string
	Method     string
	Json       bool
	DisableSSL bool
	Headers    map[string]string
}

type JSONOutputWithDomain struct {
	Domain  string              `json:"domain"`
	Results map[string]struct{} `json:"results"`
}

type headers []string

var buildNumber string
var buildVersion string
var silent bool

func (i *headers) String() string {
	return ""
}

func (i *headers) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	a := &Application{
		Output:     "",
		Target:     "",
		Method:     "GET",
		Json:       false,
		DisableSSL: false,
	}

	flag.CommandLine.StringVar(&a.Target, "target", a.Target, "Target to analyze")
	flag.CommandLine.StringVar(&a.Output, "output", a.Output, "Output file")
	flag.CommandLine.StringVar(&a.Method, "method", a.Method, "Request method")
	flag.CommandLine.BoolVar(&a.Json, "json", a.Json, "Json output format")
	flag.CommandLine.BoolVar(&a.DisableSSL, "disable-ssl", a.DisableSSL, "Don't verify the site's SSL certificate")

	h := headers{
		"user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
	}
	flag.Var(&h, "header", "Set additional request headers")

	sv := flag.Bool("version", false, "Show version and exit")
	nc := flag.Bool("no-color", false, "Disable color output")
	s := flag.Bool("silent", false, "Don't display any output")
	flag.Parse()

	silent = *s
	if *nc {
		color.NoColor = true // disables colorized output
	}

	if *sv {
		fmt.Printf("version: %s\nbuild number: %s\n", color.CyanString(buildVersion), color.CyanString(buildNumber))
		os.Exit(0)
	}

	if a.Target == "" {
		if silent == false {
			color.Red("[error] no target specified.")
		}
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: a.DisableSSL},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(a.Method, a.Target, nil)
	handleError(err) // Note: handleError exits on error

	for _, header := range h {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], parts[1])
		} else {
			handleError(errors.New("invalid header provided: " + header))
		}
	}

	resp, err := client.Do(req)
	handleError(err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	handleError(err)

	wappalyzerClient, err := wappalyzer.New()
	handleError(err)

	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)

	result := ""
	if a.Json {
		parsedURL, parseErr := url.Parse(a.Target)
		domain := ""
		if parseErr != nil {
			if silent == false {
				// Using fmt.Fprintf for warnings to stderr, as color.Red is used for fatal errors by handleError
				fmt.Fprintf(os.Stderr, "%s [warning] Could not parse hostname from target URL '%s': %v. Domain will be empty in JSON output.\n", color.YellowString("!"), a.Target, parseErr)
			}
		} else if parsedURL != nil {
			domain = parsedURL.Hostname()
			if domain == "" { // Handle cases like "localhost" or file URLs if they ever occur and parse without error but yield empty hostname
				if silent == false {
					fmt.Fprintf(os.Stderr, "%s [warning] Extracted hostname is empty for target URL '%s'. Using original target as domain in JSON output.\n", color.YellowString("!"), a.Target)
				}
				// Fallback to original target if hostname is empty but URL itself was parsable.
				// Or keep domain = "" if that's preferred for an empty hostname.
				// For consistency with "domain" field, an empty string might be better than full URL.
				// domain = a.Target
			}
		}

		outputData := JSONOutputWithDomain{
			Domain:  domain,
			Results: fingerprints,
		}
		d, err := json.MarshalIndent(outputData, "", "  ") // Use MarshalIndent for pretty JSON
		// Original was json.Marshal(fingerprints)
		handleError(err) // If marshaling fails, it will exit

		result = string(d)
	} else {
		lines := make([]string, 0)
		// The `values` in `projectdiscovery/wappalyzergo` is `struct{}`, so %v gives `{}`
		for name, _ := range fingerprints { // Changed `values` to `_` as it's not used
			lines = append(lines, fmt.Sprintf("%s: {}", color.CyanString(name)))
		}
		result = strings.Join(lines, "\n")
	}

	if a.Output != "" {
		// Ensure the result string includes a newline if it's multi-line or if user expects it
		// For JSON, it's usually self-contained. For pretty print, a final newline might be good.
		// The current behavior is to write `result` as is.
		// If `result` from json.MarshalIndent already has a newline, this is fine.
		// If not, and a newline is desired at the end of the file, add it: []byte(result + "\n")
		err = ioutil.WriteFile(a.Output, []byte(result), 0644)
		handleError(err)
	}

	if silent == false {
		// For JSON, MarshalIndent adds a final newline. For pretty print, it does not.
		// To ensure consistent printing with a newline:
		if strings.HasSuffix(result, "\n") {
			fmt.Print(result)
		} else {
			fmt.Println(result)
		}
	}
}

func handleError(err error) {
	if err != nil {
		if silent == false {
			color.Red("[error] %s", err.Error())
		}
		os.Exit(1)
	}
}