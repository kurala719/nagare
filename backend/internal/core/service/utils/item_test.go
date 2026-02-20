package utils

import "testing"

func TestParseItemValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		units string
		want  string
	}{
		{
			name:  "both empty",
			value: "",
			units: "",
			want:  "",
		},
		{
			name:  "empty value",
			value: "",
			units: "GB",
			want:  "",
		},
		{
			name:  "empty units",
			value: "100",
			units: "",
			want:  "100",
		},
		{
			name:  "both present",
			value: "50",
			units: "%",
			want:  "50 %",
		},
		{
			name:  "whitespace",
			value: " 20 ",
			units: " ms ",
			want:  "20 ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseItemValue(tt.value, tt.units); got != tt.want {
				t.Errorf("ParseItemValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
