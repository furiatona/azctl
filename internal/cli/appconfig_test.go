package cli

import (
	"testing"
)

func TestFormatAsEnv(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		expected string
	}{
		{
			name:     "empty data",
			data:     map[string]string{},
			expected: "",
		},
		{
			name: "single variable",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			expected: "export DATABASE_URL='postgres://localhost:5432/db'",
		},
		{
			name: "multiple variables sorted",
			data: map[string]string{
				"REDIS_URL":    "redis://localhost:6379",
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			expected: "export DATABASE_URL='postgres://localhost:5432/db'\nexport REDIS_URL='redis://localhost:6379'",
		},
		{
			name: "value with single quotes",
			data: map[string]string{
				"API_KEY": "key'with'quotes",
			},
			expected: "export API_KEY='key'\\''with'\\''quotes'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsEnv(tt.data)
			if result != tt.expected {
				t.Errorf("formatAsEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatAsJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    map[string]string{},
			wantErr: false,
		},
		{
			name: "single variable",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			wantErr: false,
		},
		{
			name: "multiple variables",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
				"REDIS_URL":    "redis://localhost:6379",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatAsJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Errorf("formatAsJSON() returned empty string")
			}
		})
	}
}

func TestFormatAsYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    map[string]string{},
			wantErr: false,
		},
		{
			name: "single variable",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			wantErr: false,
		},
		{
			name: "multiple variables",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
				"REDIS_URL":    "redis://localhost:6379",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatAsYAML(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatAsYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Errorf("formatAsYAML() returned empty string")
			}
		})
	}
}

func TestFormatAsDotEnv(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		expected string
	}{
		{
			name:     "empty data",
			data:     map[string]string{},
			expected: "",
		},
		{
			name: "simple values",
			data: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			expected: "DATABASE_URL=postgres://localhost:5432/db",
		},
		{
			name: "value with spaces",
			data: map[string]string{
				"API_KEY": "key with spaces",
			},
			expected: "API_KEY=\"key with spaces\"",
		},
		{
			name: "value with quotes",
			data: map[string]string{
				"PASSWORD": "pass\"word",
			},
			expected: "PASSWORD=\"pass\\\"word\"",
		},
		{
			name: "multiple variables sorted",
			data: map[string]string{
				"REDIS_URL":    "redis://localhost:6379",
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			expected: "DATABASE_URL=postgres://localhost:5432/db\nREDIS_URL=redis://localhost:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsDotEnv(tt.data)
			if result != tt.expected {
				t.Errorf("formatAsDotEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

