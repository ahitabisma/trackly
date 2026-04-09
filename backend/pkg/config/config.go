package config

import "github.com/spf13/viper"

type Config struct {
	App struct {
		Name string
		Port int
		Env  string
	}

	Database struct {
		Host            string
		Port            int
		Database        string
		User            string
		Password        string
		MaxOpenConns    int `mapstructure:"max_open_conns"`
		MaxIdleConns    int `mapstructure:"max_idle_conns"`
		ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	}

	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}

	Jwt struct {
		Secret    string
		ExpiresIn string `mapstructure:"expires_in"`
		Audience  string
	}

	RateLimit struct {
		Requests int
		Window   int
	}
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
