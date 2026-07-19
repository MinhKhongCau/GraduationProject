package database

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", ".."))
}

func TestSharedPostgresInitScriptIsLinuxExecutable(t *testing.T) {
	t.Helper()
	path := filepath.Join(repositoryRoot(t), "init-db.sh")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(content, []byte("#!/bin/bash\n")) {
		t.Fatal("init-db.sh must use an LF-only Linux shebang")
	}
	if bytes.Contains(content, []byte{'\r', '\n'}) {
		t.Fatal("init-db.sh must not contain CRLF line endings")
	}
	for _, databaseName := range [][]byte{
		[]byte(`CREATE DATABASE "${PAYMENT_DB_NAME:-payment_db}"`),
		[]byte(`CREATE DATABASE "${BOOKING_DB_NAME:-booking_db}"`),
	} {
		if !bytes.Contains(content, databaseName) {
			t.Fatalf("init-db.sh does not create required database %q", databaseName)
		}
	}
}

func TestPaymentComposeUsesContainerInfrastructureHosts(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repositoryRoot(t), "docker-compose.dev.huy.yml"))
	if err != nil {
		t.Fatal(err)
	}
	compose := string(content)
	for _, setting := range []string{
		"REDIS_HOST=redis",
		"REDIS_PORT=6379",
		"RABBITMQ_HOST=rabbitmq",
		"RABBITMQ_PORT=5672",
	} {
		if !strings.Contains(compose, setting) {
			t.Fatalf("payment-service Compose configuration is missing %q", setting)
		}
	}
}
