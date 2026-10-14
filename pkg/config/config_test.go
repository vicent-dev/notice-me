package config

import (
	"testing"
)

func TestExpandEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envKey   string
		envValue string
		want     string
	}{
		{
			name:     "simple reference uses env value",
			input:    "${FOO}",
			envKey:   "FOO",
			envValue: "bar",
			want:     "bar",
		},
		{
			name:  "simple reference unset expands to empty",
			input: "${FOO}",
			want:  "",
		},
		{
			name:  "default used when unset",
			input: "${FOO:-baz}",
			want:  "baz",
		},
		{
			name:     "env wins over default",
			input:    "${FOO:-baz}",
			envKey:   "FOO",
			envValue: "bar",
			want:     "bar",
		},
		{
			name:     "empty env treated as unset",
			input:    "${FOO:-baz}",
			envKey:   "FOO",
			envValue: "",
			want:     "baz",
		},
		{
			name:  "empty default expands to empty",
			input: "${FOO:-}",
			want:  "",
		},
		{
			name:     "multiple references",
			input:    "${DB_HOST:-127.0.0.1}:${DB_PORT:-3306}",
			envKey:   "DB_HOST",
			envValue: "10.0.0.5",
			want:     "10.0.0.5:3306",
		},
		{
			name:  "plain variable without braces is unchanged",
			input: "$FOO",
			want:  "$FOO",
		},
		{
			name:  "invalid variable name is unchanged",
			input: "${1BAD}",
			want:  "${1BAD}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}
			if got := expandEnv(tt.input); got != tt.want {
				t.Fatalf("expandEnv(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithDefaults(t *testing.T) {
	c := &Config{}
	c.withDefaults()

	if c.Server.Port != "8080" {
		t.Fatalf("Server.Port = %q, want %q", c.Server.Port, "8080")
	}
	if c.Server.Env != "development" {
		t.Fatalf("Server.Env = %q, want %q", c.Server.Env, "development")
	}
	if c.Db.Port != "3306" {
		t.Fatalf("Db.Port = %q, want %q", c.Db.Port, "3306")
	}
	if c.Rabbit.Port != "5672" {
		t.Fatalf("Rabbit.Port = %q, want %q", c.Rabbit.Port, "5672")
	}
}

func TestWithDefaultsPreservesValues(t *testing.T) {
	c := &Config{}
	c.Server.Port = "9090"
	c.Server.Env = "production"
	c.Db.Port = "13306"
	c.Rabbit.Port = "15672"

	c.withDefaults()

	if c.Server.Port != "9090" {
		t.Fatalf("Server.Port = %q, want %q", c.Server.Port, "9090")
	}
	if c.Server.Env != "production" {
		t.Fatalf("Server.Env = %q, want %q", c.Server.Env, "production")
	}
	if c.Db.Port != "13306" {
		t.Fatalf("Db.Port = %q, want %q", c.Db.Port, "13306")
	}
	if c.Rabbit.Port != "15672" {
		t.Fatalf("Rabbit.Port = %q, want %q", c.Rabbit.Port, "15672")
	}
}

func TestLoadConfig(t *testing.T) {
	c := LoadConfig()

	if c == nil {
		t.Fatal("LoadConfig returned nil")
	}
	if c.Server.Host != "localhost" {
		t.Fatalf("Server.Host = %q, want %q", c.Server.Host, "localhost")
	}
	if c.Server.Port != "8080" {
		t.Fatalf("Server.Port = %q, want %q", c.Server.Port, "8080")
	}
	if c.Db.Port != "3306" {
		t.Fatalf("Db.Port = %q, want %q", c.Db.Port, "3306")
	}
	if c.Rabbit.Port != "5672" {
		t.Fatalf("Rabbit.Port = %q, want %q", c.Rabbit.Port, "5672")
	}
	if len(c.Rabbit.Queues) != 2 {
		t.Fatalf("len(Rabbit.Queues) = %d, want %d", len(c.Rabbit.Queues), 2)
	}
}

func TestLoadConfigParsesQueues(t *testing.T) {
	c := LoadConfig()

	names := map[string]bool{}
	for _, q := range c.Rabbit.Queues {
		names[q.Name] = true
		if q.Exchange == "" {
			t.Errorf("queue %q has an empty exchange", q.Name)
		}
		if !q.Durable {
			t.Errorf("queue %q should be durable", q.Name)
		}
	}

	if !names["notification.create"] || !names["notification.notify"] {
		t.Fatalf("expected create/notify queues, got %v", names)
	}
}
