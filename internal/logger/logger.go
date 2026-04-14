package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	enabled bool
	l       *log.Logger
)

// Init enables debug logging to ~/.stripe-seeder-debug.log.
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".stripe-seeder-debug.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("não foi possível abrir log: %w", err)
	}
	l = log.New(f, "", log.Ltime|log.Lmicroseconds)
	enabled = true
	l.Println("=== stripe-seeder debug session started ===")
	return nil
}

func Enabled() bool { return enabled }

func Log(format string, args ...any) {
	if enabled {
		l.Printf(format, args...)
	}
}
