package configs

type Config struct {
	BindAddr     string   `toml:"bind_addr"`
	TlsCertFile  string   `toml:"tls_cert"`
	TlsKeyFile   string   `toml:"tls_key"`
	Token        string   `toml:"token"`
	IpSetFile    string   `toml:"ipset_file"`
	Fail2banLogs string   `toml:"fail2ban_logs"`
	NeverIpsFile string   `toml:"neverips_file"`
	LogLevel     string   `toml:"log_level"`
	ServerAddr   []string `toml:"server_addr"`
	AbuseipdbKey string   `toml:"abuseipdb_key"`
}

func NewConfig() *Config {
	return &Config{}
}
