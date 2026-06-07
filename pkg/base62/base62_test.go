package base62_test

import (
	"shortener/pkg/base62"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
		want string
	}{
		{"Zero value", 0, "0"},
		{"Single digit", 9, "9"},
		{"First letter a", 10, "a"},
		{"Last letter z", 35, "z"},
		{"First capital A", 36, "A"},
		{"Last capital Z", 61, "Z"},
		{"Base case 62", 62, "10"},
		{"Large ID", 123456789, "8m0Kx"},
		{"Max uint64", 18446744073709551615, "lYGhA16ahyf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := base62.Encode(tt.id); got != tt.want {
				t.Errorf("Encode(%d) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		want    uint64
		wantErr bool
	}{
		{"Zero value", "0", 0, false},
		{"Single digit", "9", 9, false},
		{"First letter a", "a", 10, false},
		{"Last letter z", "z", 35, false},
		{"First capital A", "A", 36, false},
		{"Last capital Z", "Z", 61, false},
		{"Base case 10", "10", 62, false},
		{"Large ID", "8m0Kx", 123456789, false},
		{"Max uint64", "lYGhA16ahyf", 18446744073709551615, false},
		{"Invalid character hypen", "abc-123", 0, true},
		{"Invalid character cyrillic", "abс", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base62.Decode(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode(%s) error = %v, wantErr %v", tt.encoded, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode(%s) = %v, want %v", tt.encoded, got, tt.want)
			}
		})
	}
}

func TestRoundtrip(t *testing.T) {
	var i uint64
	for i = 0; i < 100000; i++ {
		encoded := base62.Encode(i)
		decoded, err := base62.Decode(encoded)
		if err != nil {
			t.Fatalf("Ошибка декодирования на числе %d: %v", i, err)
		}
		if decoded != i {
			t.Fatalf("Несовпадение данных: исходное %d, закодированное %s, декодированное %d", i, encoded, decoded)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		base62.Encode(123456789)
	}
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = base62.Decode("8M9b1")
	}
}
