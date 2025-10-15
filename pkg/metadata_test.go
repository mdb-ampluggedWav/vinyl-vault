package pkg

import "testing"

func TestDuration_String(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected string
	}{
		{"30 seconds", Duration(30), "0:30"},
		{"1 minute", Duration(60), "1:00"},
		{"3 minutes 45 seconds", Duration(225), "3:45"},
		{"10 minutes 5 seconds", Duration(605), "10:05"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.duration.String()
			if result != tt.expected {
				t.Errorf("Duration.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDuration_StringLong(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected string
	}{
		{"30 seconds", Duration(30), "0:30"},
		{"1 hour", Duration(3600), "1:00:00"},
		{"1 hour 5 min 30 sec", Duration(3930), "1:05:30"},
		{"2 hours 15 min", Duration(8100), "2:15:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.duration.StringLong()
			if result != tt.expected {
				t.Errorf("Duration.StringLong() = %v, want %v", result, tt.expected)
			}
		})
	}
}
