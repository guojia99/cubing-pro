package configs_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/guojia99/cubing-pro/src/configs"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfigKnownFields(t *testing.T) {
	t.Parallel()

	file, err := os.Open(filepath.Join("..", "..", "etc", "server.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	var cfg configs.Config
	if err := decoder.Decode(&cfg); err != nil {
		t.Fatalf("default server.yaml must match Config: %v", err)
	}
}

func TestConfigLoadGatewayCompatibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yaml string
		want configs.GatewayConfig
	}{
		{
			name: "gateway httpPort",
			yaml: "gateway: {httpPort: 8080, indexPath: current.html}\n",
			want: configs.GatewayConfig{HttpPort: 8080, IndexPath: "current.html"},
		},
		{
			name: "legacy apiGateway",
			yaml: "apiGateway: {httpPort: 8081, indexPath: legacy.html}\n",
			want: configs.GatewayConfig{HttpPort: 8081, IndexPath: "legacy.html"},
		},
		{
			name: "gateway port fallback",
			yaml: "gateway: {port: 8082}\n",
			want: configs.GatewayConfig{HttpPort: 8082, Port: 8082},
		},
		{
			name: "legacy apiGateway port fallback",
			yaml: "apiGateway: {port: 8083}\n",
			want: configs.GatewayConfig{HttpPort: 8083, Port: 8083},
		},
		{
			name: "httpPort takes precedence",
			yaml: "gateway: {httpPort: 8084, port: 8085}\n",
			want: configs.GatewayConfig{HttpPort: 8084, Port: 8085},
		},
		{
			name: "legacy httpPort takes precedence",
			yaml: "apiGateway: {httpPort: 8084, port: 8085}\n",
			want: configs.GatewayConfig{HttpPort: 8084, Port: 8085},
		},
		{
			name: "zero httpPort falls back to port",
			yaml: "gateway: {httpPort: 0, port: 8086}\n",
			want: configs.GatewayConfig{HttpPort: 8086, Port: 8086},
		},
		{
			name: "zero ports remain zero",
			yaml: "gateway: {httpPort: 0, port: 0}\n",
			want: configs.GatewayConfig{},
		},
		{
			name: "gateway takes precedence without merging",
			yaml: "apiGateway: {httpPort: 8087, indexPath: legacy.html}\ngateway: {port: 8088}\n",
			want: configs.GatewayConfig{HttpPort: 8088, Port: 8088},
		},
		{
			name: "empty gateway takes precedence",
			yaml: "apiGateway: {port: 8087}\ngateway: {}\n",
			want: configs.GatewayConfig{},
		},
		{
			name: "null gateway takes precedence",
			yaml: "apiGateway: {port: 8087}\ngateway: null\n",
			want: configs.GatewayConfig{},
		},
		{
			name: "unknown fields remain accepted by Load",
			yaml: "unknown: true\ngateway: {port: 8089, unknown: true}\n",
			want: configs.GatewayConfig{HttpPort: 8089, Port: 8089},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "server.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			var cfg configs.Config
			if err := cfg.Load(path); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cfg.Gateway, tt.want) {
				t.Errorf("Gateway = %+v, want %+v", cfg.Gateway, tt.want)
			}
		})
	}
}

func TestConfigLoadErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		var cfg configs.Config
		err := cfg.Load(filepath.Join(t.TempDir(), "missing.yaml"))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Load error = %v, want os.ErrNotExist", err)
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		t.Parallel()

		var cfg configs.Config
		err := cfg.Load(t.TempDir())
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			t.Fatalf("Load error = %v, want *os.PathError", err)
		}
	})

	tests := []struct {
		name      string
		yaml      string
		typeError bool
	}{
		{
			name: "invalid YAML",
			yaml: "gateway: [\n",
		},
		{
			name:      "invalid gateway field type",
			yaml:      "gateway: {httpPort: invalid}\n",
			typeError: true,
		},
		{
			name:      "invalid legacy gateway field type",
			yaml:      "apiGateway: {port: invalid}\n",
			typeError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "server.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			var cfg configs.Config
			err := cfg.Load(path)
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			if tt.typeError {
				var typeErr *yaml.TypeError
				if !errors.As(err, &typeErr) {
					t.Fatalf("Load error = %v, want *yaml.TypeError", err)
				}
			}
		})
	}
}
