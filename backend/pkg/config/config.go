package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name        string
		Port        int
		Env         string
		FrontendURL string `mapstructure:"frontend_url"`
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

	AppsScript struct {
		URL               string
		Timeout           int
		PollIntervalMs    int `mapstructure:"poll_interval_ms"`
		PollMaxRetries    int `mapstructure:"poll_max_retries"`
		PythonScriptPath  string `mapstructure:"python_script_path"`
	}

	NvidiaApiKey string `mapstructure:"nvidia_api_key"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
