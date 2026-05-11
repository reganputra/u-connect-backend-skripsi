package utils

import (
	"os"
	"strings"
)

// LoadEnvFile membaca file .env dan menyetel variabel lingkungan.
// File ini melewatkan baris kosong dan baris komentar yang dimulai dengan '#'.
// Variabel lingkungan yang ada tidak ditimpa.
func LoadEnvFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// File .env opsional, abaikan secara diam-diam jika tidak ditemukan
		return
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Lewati baris kosong dan komentar
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Hanya mengatur jika belum didefinisikan di lingkungan
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// GetEnv mengembalikan nilai variabel lingkungan yang bernama dengan key.
// Jika variabel tidak diatur atau kosong, ia mengembalikan nilai fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
