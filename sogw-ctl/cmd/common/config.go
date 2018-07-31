package common

type config struct {
	Publish struct {
		Mapper struct {
			PathPattern string `mapstructure:"pathPattern"`
			Services    map[string]*struct {
				Servers []*struct {
					Addr string `mapstructure:"addr"`
				} `mapstructure:"servers"`
				Config *struct {
					Status     string `mapstructure:"status"`
					LoadBlance string `mapstructure:"loadBlance"`
				} `mapstructure:"config"`
			} `mapstructure:"services"`
		} `mapstructure:"mapper"`
		Swagger struct {
			File string `mapstructure:"file"`
		} `mapstructure:"swagger"`
	} `mapstructure:"publish"`

	Store struct {
		Url     string                 `mapstructure:"url"`
		Options map[string]interface{} `mapstructure:"options"`
	} `mapstructure:"store"`
}

var Config *config = &config{}
