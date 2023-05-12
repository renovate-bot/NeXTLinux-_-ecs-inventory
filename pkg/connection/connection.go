package connection

// Information for posting in-use image details to Anchore (or any URL for that matter)
type AnchoreInfo struct {
	URL      string     `mapstructure:"url"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Account  string     `mapstructure:"account"`
	HTTP     HTTPConfig `mapstructure:"http"`
}

// Configurations for the HTTP Client itself (net/http)
type HTTPConfig struct {
	Insecure       bool `mapstructure:"insecure"`
	TimeoutSeconds int  `mapstructure:"timeout-seconds"`
}

// Return whether or not AnchoreDetails are specified
func (nextlinux *AnchoreInfo) IsValid() bool {
	return nextlinux.URL != "" &&
		nextlinux.User != "" &&
		nextlinux.Password != ""
}
