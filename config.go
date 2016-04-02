package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gopkg.in/yaml.v2"
)

type accountConfig struct {
	Password string   `yaml:"password"`
	OAuth    []string `yaml:"oauth"`
}

type filetypeConfig struct {
	Name      string   `yaml:"name"`
	Mime      string   `yaml:"mimetype"`
	Pygments  string   `yaml:"pygments"`
	DisplayAs string   `yaml:"displayAs"`
	IconFile  string   `yaml:"icon"`
	Patterns  []string `yaml:"patterns"`
}

type expiryConfig struct {
	Ident    string
	Name     string
	Duration string
	Days     int
	Months   int
	Years    int
}

func (e expiryConfig) AddTo(t time.Time) (*time.Time, error) {
	if len(e.Duration) > 0 {
		d, err := time.ParseDuration(e.Duration)
		if err != nil {
			return nil, err
		}

		result := t.Add(d)

		return &result, nil
	}

	// this represents "no expiry"
	if e.Days == 0 && e.Months == 0 && e.Years == 0 {
		return nil, nil
	}

	result := t.AddDate(e.Years, e.Months, e.Days)

	return &result, nil
}

type configuration struct {
	Environment string `yaml:"environment"`

	Directories struct {
		Storage   string `yaml:"storage"`
		Resources string `yaml:"resources"`
		Www       string `yaml:"www"`
	}

	Accounts map[string]accountConfig `yaml:"accounts"`

	OAuth map[string]struct {
		ClientID     string   `yaml:"clientId"`
		ClientSecret string   `yaml:"clientSecret"`
		Scopes       []string `yaml:"scopes"`
	}

	Pastebin []struct {
		Name      string   `yaml:"name"`
		FileTypes []string `yaml:"filetypes"`
	}

	Expiries  []expiryConfig
	FileTypes map[string]filetypeConfig

	Server struct {
		Listen          string   `yaml:"listen"`
		BaseUrl         string   `yaml:"baseUrl"`
		MaxRequestSize  int      `yaml:"maxRequestSize"`
		CertificateFile string   `yaml:"certificateFile"`
		PrivateKeyFile  string   `yaml:"privateKeyFile"`
		Ciphers         []string `yaml:"ciphers"`
	} `yaml:"server"`

	Session struct {
		CookieName   string         `yaml:"cookieName"`
		Lifetime     *time.Duration `yaml:"lifetime"`
		CookieSecure bool           `yaml:"cookieSecure"`
		CookiePath   string         `yaml:"cookiePath"`
	} `yaml:"session"`
}

func loadConfiguration(filename string, hostyPath string) (*configuration, error) {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	config := configuration{}

	if yaml.Unmarshal(content, &config) != nil {
		return &config, errors.New("Could not load configuration file '" + filename + "'. Make sure all values are valid, especially the session lifetime needs to be parsable.")
	}

	// perform some sanity checks
	if config.Environment != gin.DebugMode && config.Environment != gin.ReleaseMode && config.Environment != gin.TestMode {
		return &config, errors.New("Invalid environment (" + config.Environment + ") configured.")
	}

	absfile, _ := filepath.Abs(filename)
	configDir := filepath.Dir(absfile)

	adjustPaths := func(s string) string {
		return strings.Replace(strings.Replace(s, "%config%", configDir, -1), "%hosty%", hostyPath, -1)
	}

	config.Directories.Storage = adjustPaths(config.Directories.Storage)
	config.Directories.Resources = adjustPaths(config.Directories.Resources)
	config.Directories.Www = adjustPaths(config.Directories.Www)

	return &config, nil
}

func (c *configuration) DatabaseFile() string {
	return filepath.Join(c.Directories.Storage, "database.sqlite3")
}

func (c *configuration) CipherSuites() []uint16 {
	ciphers := make([]uint16, 0)

	for _, cipher := range c.Server.Ciphers {
		var c uint16

		switch cipher {
		case "TLS_RSA_WITH_RC4_128_SHA":
			c = tls.TLS_RSA_WITH_RC4_128_SHA
		case "TLS_RSA_WITH_3DES_EDE_CBC_SHA":
			c = tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA
		case "TLS_RSA_WITH_AES_128_CBC_SHA":
			c = tls.TLS_RSA_WITH_AES_128_CBC_SHA
		case "TLS_RSA_WITH_AES_256_CBC_SHA":
			c = tls.TLS_RSA_WITH_AES_256_CBC_SHA
		case "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":
			c = tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA
		case "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":
			c = tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA
		case "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":
			c = tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
		case "TLS_ECDHE_RSA_WITH_RC4_128_SHA":
			c = tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA
		case "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":
			c = tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA
		case "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":
			c = tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
		case "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":
			c = tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			c = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
			c = tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			c = tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
		case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
			c = tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
		case "TLS_FALLBACK_SCSV":
			c = tls.TLS_FALLBACK_SCSV
		default:
			panic("Unknown cipher '" + cipher + "' configured.")
		}

		ciphers = append(ciphers, c)
	}

	return ciphers
}

func (c *configuration) Expiry(ident string) *expiryConfig {
	for _, exp := range c.Expiries {
		if exp.Ident == ident {
			return &exp
		}
	}

	return nil
}

func (c *configuration) AccountByUsername(username string) *accountConfig {
	acc, ok := c.Accounts[username]
	if ok {
		return &acc
	}

	return nil
}

func (c *configuration) UsernameByOAuthIdentity(ident string) string {
	for username, config := range c.Accounts {
		for _, i := range config.OAuth {
			if i == ident {
				return username
			}
		}
	}

	return ""
}

func (c *configuration) FileTypeIdentByPygments(pygments string) string {
	for ident, filetype := range c.FileTypes {
		if filetype.Pygments == pygments {
			return ident
		}
	}

	return ""
}

func (c *configuration) FileTypeByIdentifier(identifier string) *filetypeConfig {
	if identifier == "" {
		return c.FallbackFileType()
	}

	ft, ok := c.FileTypes[identifier]
	if ok {
		return &ft
	}

	return c.FallbackFileType()
}

func (c *configuration) FileTypeIdentByFilename(filename string) string {
	candidate := ""

	for ident, filetype := range c.FileTypes {
		for _, pattern := range filetype.Patterns {
			// exact match is a direct win
			if pattern == filename {
				return ident
			}

			if matched, _ := filepath.Match(pattern, filename); matched {
				candidate = ident
			}
		}
	}

	return candidate
}

func (c *configuration) FallbackFileType() *filetypeConfig {
	return &filetypeConfig{
		DisplayAs: "link",
		Mime:      "application/octet-stream",
		Name:      "Binary File",
		IconFile:  "blank-file",
		Patterns:  make([]string, 0),
		Pygments:  "",
	}
}
