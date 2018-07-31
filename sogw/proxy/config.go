package proxy

type Config struct {
	// listen address
	Addr     string `mapstructure:"addr"`
	TLSAddr  string `mapstructure:"tls_addr"`
	UnixAddr string `mapstructure:"unix_addr"`

	// k/v store
	Store StoreConfig `mapstructure:"store"`

	Jobs    map[string]interface{} `mapstructure:"jobs"`
	Filters map[string]interface{} `mapstructure:"filters"`
}

type StoreConfig struct {
	Url     string                 `mapstructure:"url"`
	Watch   bool                   `mapstructure:"watch"`
	Options map[string]interface{} `mapstructure:"options"`
}
