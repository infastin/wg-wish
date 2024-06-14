package app

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/infastin/go-validation"
	isint "github.com/infastin/go-validation/is/int"
	isstr "github.com/infastin/go-validation/is/str"
	"github.com/rs/zerolog"
)

type Config struct {
	Logger    LoggerConfig    `env-prefix:"LOGGER_" yaml:"logger"`
	Database  DatabaseConfig  `env-prefix:"DB_" yaml:"db"`
	WireGuard WireGuardConfig `env-prefix:"WG_" yaml:"wireguard"`
	SSH       SSHConfig       `env-prefix:"SSH_" yaml:"ssh"`
}

func (cfg *Config) Default() {
	cfg.Logger.Default()
	cfg.Database.Default()
	cfg.WireGuard.Default()
	cfg.SSH.Default()
}

func (cfg *Config) Validate() error {
	return validation.All(
		validation.Ptr(&cfg.Logger, "logger").With(validation.Custom),
		validation.Ptr(&cfg.WireGuard, "wg").With(validation.Custom),
		validation.Ptr(&cfg.SSH, "ssh").With(validation.Custom),
	)
}

type LoggerConfig struct {
	Level      string `env:"LEVEL" yaml:"level"`
	Directory  string `env:"DIRECTORY" yaml:"directory"`
	MaxSize    int    `env:"MAX_SIZE" yaml:"max_size"`
	MaxAge     int    `env:"MAX_AGE" yaml:"max_age"`
	MaxBackups int    `env:"MAX_BACKUPS" yaml:"max_backups"`
}

func (cfg *LoggerConfig) Default() {
	if cfg.Level == "" {
		cfg.Level = zerolog.LevelInfoValue
	}
}

func (cfg *LoggerConfig) Validate() error {
	return validation.All(
		validation.Comparable(cfg.Level, "level").In(
			zerolog.LevelTraceValue, zerolog.LevelDebugValue, zerolog.LevelInfoValue,
			zerolog.LevelWarnValue, zerolog.LevelErrorValue,
		),
		validation.Number(cfg.MaxSize, "max_size").If(cfg.Directory != "").GreaterEqual(0).EndIf(),
		validation.Number(cfg.MaxAge, "max_age").If(cfg.Directory != "").GreaterEqual(0).EndIf(),
		validation.Number(cfg.MaxBackups, "max_backups").If(cfg.Directory != "").GreaterEqual(0).EndIf(),
	)
}

type DatabaseConfig struct {
	Path string `env:"PATH" yaml:"path"`
}

func (cfg *DatabaseConfig) Default() {
	if cfg.Path == "" {
		cfg.Path = "/var/lib/wg-wish/wg-wish.db"
	}
}

type WireGuardConfig struct {
	Host                string   `env:"HOST" yaml:"host"`
	Path                string   `env:"PATH" yaml:"path"`
	Address             string   `env:"ADDRESS" yaml:"address"`
	Port                int      `env:"PORT" yaml:"port"`
	Device              string   `env:"DEVICE" yaml:"device"`
	AllowedIPs          []string `env:"ALLOWED_IPS" yaml:"allowed_ips"`
	PersistentKeepalive int      `env:"PERSISTENT_KEEPALIVE" yaml:"persistent_keepalive"`
	DNS                 []string `env:"DNS" yaml:"dns"`
}

func (cfg *WireGuardConfig) Default() {
	if cfg.Path == "" {
		cfg.Path = "/var/lib/wg-wish/wg0.conf"
	}

	if cfg.Address == "" {
		cfg.Address = "10.9.8.1/24"
	}

	if cfg.Port == 0 {
		cfg.Port = 51820
	}

	if cfg.Device == "" {
		cfg.Device = "eth0"
	}

	if len(cfg.AllowedIPs) == 0 {
		cfg.AllowedIPs = []string{"0.0.0.0/0"}
	}

	if cfg.PersistentKeepalive == 0 {
		cfg.PersistentKeepalive = 25
	}

	if len(cfg.DNS) == 0 {
		cfg.DNS = []string{"1.1.1.1", "8.8.8.8"}
	}
}

func (cfg *WireGuardConfig) Validate() error {
	return validation.All(
		validation.String(cfg.Host, "host").Required(true).With(isstr.Host),
		validation.String(cfg.Path, "path").Required(true),
		validation.String(cfg.Address, "address").Required(true).With(isstr.CIDR),
		validation.Number(cfg.Port, "port").Required(true).With(isint.Port),
		validation.String(cfg.Device, "device").Required(true),
		validation.Slice(cfg.AllowedIPs, "allowed_ips").Required(true).ValuesWith(isstr.CIDR),
		validation.Number(cfg.PersistentKeepalive, "persistent_keepalive").GreaterEqual(0),
		validation.Slice(cfg.DNS, "dns").Required(true).ValuesWith(isstr.IP),
	)
}

type SSHConfig struct {
	Port        int      `env:"PORT" yaml:"port"`
	HostKeyPath string   `env:"HOST_KEY_PATH" yaml:"host_key_path"`
	AdminKeys   []string `env:"ADMIN_KEYS" yaml:"admin_keys"`
}

func (cfg *SSHConfig) Default() {
	if cfg.Port == 0 {
		cfg.Port = 51822
	}

	if cfg.HostKeyPath == "" {
		cfg.HostKeyPath = "/var/lib/wg-wish/host_ed25519"
	}
}

func (cfg *SSHConfig) Validate() error {
	return validation.All(
		validation.Number(cfg.Port, "port").Required(true).With(isint.Port),
		validation.String(cfg.HostKeyPath, "host_key_path").Required(true),
	)
}

func NewConfig(configPath string) (cfg Config, err error) {
	if configPath != "" {
		err = cleanenv.ReadConfig(configPath, &cfg)
		if err != nil {
			return Config{}, fmt.Errorf("failed to read config: %w", err)
		}
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read envvars: %w", err)
	}

	cfg.Default()

	err = cfg.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
