package conf

type ServiceCfg struct {
	MySql struct {
		Dsn    string `yaml:"dsn"`
		Driver string `yaml:"driver"`
	} `yaml:"mysql"`

	Redis struct {
		Address string `yaml:"address"`
	} `yaml:"redis"`

	Address string `yaml:"address"`

	Tls struct {
		Server struct {
			Pem string `yaml:"pem"`
			Key string `yaml:"key"`
		} `yaml:"server"`
		Client struct {
			Pem string `yaml:"pem"`
			Key string `yaml:"key"`
		} `yaml:"client"`
	} `yaml:"tls"`
}
