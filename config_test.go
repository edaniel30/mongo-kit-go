package mongo_kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "mongodb://localhost:27017", cfg.URI)
	assert.Equal(t, "default", cfg.Database)
	assert.Equal(t, uint64(100), cfg.MaxPoolSize)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Nil(t, cfg.ClientOptions)
}

func TestConfigOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(t *testing.T, cfg Config)
	}{
		{
			name:   "WithURI sets URI",
			option: WithURI("mongodb://prod:27017"),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "mongodb://prod:27017", cfg.URI)
			},
		},
		{
			name:   "WithURI with credentials",
			option: WithURI("mongodb://user:pass@localhost:27017"),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "mongodb://user:pass@localhost:27017", cfg.URI)
			},
		},
		{
			name:   "WithURI empty",
			option: WithURI(""),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "", cfg.URI)
			},
		},
		{
			name:   "WithDatabase sets database",
			option: WithDatabase("myapp"),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "myapp", cfg.Database)
			},
		},
		{
			name:   "WithMaxPoolSize sets pool size",
			option: WithMaxPoolSize(200),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, uint64(200), cfg.MaxPoolSize)
			},
		},
		{
			name:   "WithTimeout sets timeout",
			option: WithTimeout(30 * time.Second),
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, 30*time.Second, cfg.Timeout)
			},
		},
		{
			name:   "WithClientOptions sets options",
			option: WithClientOptions(options.Client().SetMinPoolSize(10)),
			validate: func(t *testing.T, cfg Config) {
				require.NotNil(t, cfg.ClientOptions)
			},
		},
		{
			name:   "WithClientOptions nil",
			option: WithClientOptions(nil),
			validate: func(t *testing.T, cfg Config) {
				assert.Nil(t, cfg.ClientOptions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.option(&cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestOptionChaining(t *testing.T) {
	cfg := DefaultConfig()
	WithURI("mongodb://prod:27017")(&cfg)
	WithDatabase("production")(&cfg)
	WithMaxPoolSize(500)(&cfg)
	WithTimeout(30 * time.Second)(&cfg)

	assert.Equal(t, "mongodb://prod:27017", cfg.URI)
	assert.Equal(t, "production", cfg.Database)
	assert.Equal(t, uint64(500), cfg.MaxPoolSize)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorField  string
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "valid custom config",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Database:    "testdb",
				MaxPoolSize: 50,
				Timeout:     5 * time.Second,
			},
			expectError: false,
		},
		{
			name: "empty URI",
			config: Config{
				URI:         "",
				Database:    "testdb",
				MaxPoolSize: 100,
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorField:  "URI",
			errorMsg:    "is required",
		},
		{
			name: "empty database",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Database:    "",
				MaxPoolSize: 100,
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorField:  "Database",
			errorMsg:    "is required",
		},
		{
			name: "zero max pool size",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Database:    "testdb",
				MaxPoolSize: 0,
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorField:  "MaxPoolSize",
			errorMsg:    "must be greater than 0",
		},
		{
			name: "zero timeout",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Database:    "testdb",
				MaxPoolSize: 100,
				Timeout:     0,
			},
			expectError: true,
			errorField:  "Timeout",
			errorMsg:    "must be greater than 0",
		},
		{
			name: "negative timeout",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Database:    "testdb",
				MaxPoolSize: 100,
				Timeout:     -1 * time.Second,
			},
			expectError: true,
			errorField:  "Timeout",
			errorMsg:    "must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.expectError {
				require.Error(t, err)
				var configErr *ConfigError
				require.ErrorAs(t, err, &configErr)
				assert.Equal(t, tt.errorField, configErr.Field)
				assert.Contains(t, configErr.Message, tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
