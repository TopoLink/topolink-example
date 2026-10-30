package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://demo")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("KAFKA_BROKERS", "a:9092, b:9092")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Errorf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
}

func TestLoadRequiresSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("want error for missing DATABASE_URL / JWT_SECRET")
	}
}
