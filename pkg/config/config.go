package config

// Config holds the specific configuration for the StreamGate instance.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
}

type ServerConfig struct {
	TCPPort  int `yaml:"tcp_port"`
	UDPPort  int `yaml:"udp_port"`
	HTTPPort int `yaml:"http_port"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Channel  string `yaml:"channel"` // PubSub channel name
}

// DefaultConfig returns a safe default configuration.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			TCPPort:  8081,
			UDPPort:  8082,
			HTTPPort: 8080,
		},
		Redis: RedisConfig{
			Address: "localhost:6379",
			Channel: "streamgate_config",
		},
	}
}
