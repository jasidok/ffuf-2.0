package ffuf

import (
	"context"
)

type Config struct {
	AuditLog                  string                `json:"auditlog"`
	AutoCalibration           bool                  `json:"autocalibration"`
	AutoCalibrationKeyword    string                `json:"autocalibration_keyword"`
	AutoCalibrationPerHost    bool                  `json:"autocalibration_perhost"`
	AutoCalibrationStrategies []string              `json:"autocalibration_strategies"`
	AutoCalibrationStrings    []string              `json:"autocalibration_strings"`
	Cancel                    context.CancelFunc    `json:"-"`
	Colors                    bool                  `json:"colors"`
	CommandKeywords           []string              `json:"-"`
	CommandLine               string                `json:"cmdline"`
	ConfigFile                string                `json:"configfile"`
	Context                   context.Context       `json:"-"`
	Data                      string                `json:"postdata"`
	Debuglog                  string                `json:"debuglog"`
	Delay                     optRange              `json:"delay"`
	DirSearchCompat           bool                  `json:"dirsearch_compatibility"`
	Encoders                  []string              `json:"encoders"`
	Extensions                []string              `json:"extensions"`
	FilterMode                string                `json:"fmode"`
	FollowRedirects           bool                  `json:"follow_redirects"`
	Headers                   map[string]string     `json:"headers"`
	IgnoreBody                bool                  `json:"ignorebody"`
	IgnoreWordlistComments    bool                  `json:"ignore_wordlist_comments"`
	InputMode                 string                `json:"inputmode"`
	InputNum                  int                   `json:"cmd_inputnum"`
	InputProviders            []InputProviderConfig `json:"inputproviders"`
	InputShell                string                `json:"inputshell"`
	Json                      bool                  `json:"json"`
	MatcherManager            MatcherManager        `json:"matchers"`
	MatcherMode               string                `json:"mmode"`
	MaxTime                   int                   `json:"maxtime"`
	MaxTimeJob                int                   `json:"maxtime_job"`
	Method                    string                `json:"method"`
	Noninteractive            bool                  `json:"noninteractive"`
	OutputDirectory           string                `json:"outputdirectory"`
	OutputFile                string                `json:"outputfile"`
	OutputFormat              string                `json:"outputformat"`
	OutputSkipEmptyFile       bool                  `json:"OutputSkipEmptyFile"`
	ProgressFrequency         int                   `json:"-"`
	ProxyURL                  string                `json:"proxyurl"`
	Quiet                     bool                  `json:"quiet"`
	Rate                      int64                 `json:"rate"`
	Raw                       bool                  `json:"raw"`
	Recursion                 bool                  `json:"recursion"`
	RecursionDepth            int                   `json:"recursion_depth"`
	RecursionStrategy         string                `json:"recursion_strategy"`
	ReplayProxyURL            string                `json:"replayproxyurl"`
	RequestFile               string                `json:"requestfile"`
	RequestProto              string                `json:"requestproto"`
	ScraperFile               string                `json:"scraperfile"`
	Scrapers                  string                `json:"scrapers"`
	SNI                       string                `json:"sni"`
	StopOn403                 bool                  `json:"stop_403"`
	StopOnAll                 bool                  `json:"stop_all"`
	StopOnErrors              bool                  `json:"stop_errors"`
	Threads                   int                   `json:"threads"`
	Timeout                   int                   `json:"timeout"`
	Url                       string                `json:"url"`
	Verbose                   bool                  `json:"verbose"`
	Wordlists                 []string              `json:"wordlists"`
	Http2                     bool                  `json:"http2"`
	ClientCert                string                `json:"client-cert"`
	ClientKey                 string                `json:"client-key"`
	// API-specific options
	APIMode                   bool                  `json:"api_mode"`
	APIWordlistPath           string                `json:"api_wordlist_path"`
	APIWordlistCategory       string                `json:"api_wordlist_category"`
	APIAuthType               string                `json:"api_auth_type"`
	APIAuthUsername           string                `json:"api_auth_username"`
	APIAuthPassword           string                `json:"api_auth_password"`
	APIAuthToken              string                `json:"api_auth_token"`
	APIAuthAPIKey             string                `json:"api_auth_api_key"`
	APIAuthAPIKeyName         string                `json:"api_auth_api_key_name"`
	APIAuthAPIKeyLoc          string                `json:"api_auth_api_key_loc"`
	APIPayloadFormat          string                `json:"api_payload_format"`
	APIPayloadTemplate        string                `json:"api_payload_template"`
	APIPayloadPath            string                `json:"api_payload_path"`
	APIParseResponseBody      bool                  `json:"api_parse_response_body"`
	APIExtractEndpoints       bool                  `json:"api_extract_endpoints"`
	APIOutputFormat           bool                  `json:"api_output_format"`
}

type InputProviderConfig struct {
	Name     string `json:"name"`
	Keyword  string `json:"keyword"`
	Value    string `json:"value"`
	Encoders string `json:"encoders"`
	Template string `json:"template"` // the templating string used for sniper mode (usually "§")
}

func NewConfig(ctx context.Context, cancel context.CancelFunc) Config {
	var conf Config
	conf.AutoCalibrationKeyword = "FUZZ"
	conf.AutoCalibrationStrategies = []string{"basic"}
	conf.AutoCalibrationStrings = make([]string, 0)
	conf.CommandKeywords = make([]string, 0)
	conf.Context = ctx
	conf.Cancel = cancel
	conf.Data = ""
	conf.Debuglog = ""
	conf.Delay = optRange{0, 0, false, false}
	conf.DirSearchCompat = false
	conf.Encoders = make([]string, 0)
	conf.Extensions = make([]string, 0)
	conf.FilterMode = "or"
	conf.FollowRedirects = false
	conf.Headers = make(map[string]string)
	conf.IgnoreWordlistComments = false
	conf.InputMode = "clusterbomb"
	conf.InputNum = 0
	conf.InputShell = ""
	conf.InputProviders = make([]InputProviderConfig, 0)
	conf.Json = false
	conf.MatcherMode = "or"
	conf.MaxTime = 0
	conf.MaxTimeJob = 0
	conf.Method = "GET"
	conf.Noninteractive = false
	conf.ProgressFrequency = 125
	conf.ProxyURL = ""
	conf.Quiet = false
	conf.Rate = 0
	conf.Raw = false
	conf.Recursion = false
	conf.RecursionDepth = 0
	conf.RecursionStrategy = "default"
	conf.RequestFile = ""
	conf.RequestProto = "https"
	conf.SNI = ""
	conf.ScraperFile = ""
	conf.Scrapers = "all"
	conf.StopOn403 = false
	conf.StopOnAll = false
	conf.StopOnErrors = false
	conf.Timeout = 10
	conf.Url = ""
	conf.Verbose = false
	conf.Wordlists = []string{}
	conf.Http2 = false

	// Initialize API-specific options
	conf.APIMode = false
	conf.APIWordlistPath = ""
	conf.APIWordlistCategory = ""
	conf.APIAuthType = ""
	conf.APIAuthUsername = ""
	conf.APIAuthPassword = ""
	conf.APIAuthToken = ""
	conf.APIAuthAPIKey = ""
	conf.APIAuthAPIKeyName = ""
	conf.APIAuthAPIKeyLoc = "header"
	conf.APIPayloadFormat = "json"
	conf.APIPayloadTemplate = ""
	conf.APIPayloadPath = ""
	conf.APIParseResponseBody = false
	conf.APIExtractEndpoints = false
	conf.APIOutputFormat = false

	return conf
}

func (c *Config) SetContext(ctx context.Context, cancel context.CancelFunc) {
	c.Context = ctx
	c.Cancel = cancel
}
