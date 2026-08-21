package oprom

// HTTPClientConfig is the subset of prometheus/common/config.HTTPClientConfig
// that Oodle notifiers need. It is redeclared here with a plain string
// credential because the upstream Secret type marshals as "<secret>" unless the
// package global config.MarshalSecretValue is set.
type HTTPClientConfig struct {
	Authorization   *Authorization `yaml:"authorization,omitempty" json:"authorization,omitempty"`
	FollowRedirects bool           `yaml:"follow_redirects" json:"follow_redirects"`
	EnableHTTP2     bool           `yaml:"enable_http2" json:"enable_http2"`
}

// Authorization holds the credentials sent in the Authorization header.
type Authorization struct {
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	Credentials string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
}
