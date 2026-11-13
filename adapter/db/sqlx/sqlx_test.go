package sqlx

import "testing"

func TestNew_EmptyDSN_ReturnsError(t *testing.T) {
	_, err := New("", Config{}, nil)
	if err == nil {
		t.Fatal("expected error for empty dsn")
	}
}

func TestAdapter_Close_NilSafe(t *testing.T) {
	var nilAdapter *adapter
	if err := nilAdapter.Close(); err != nil {
		t.Fatalf("nil adapter close returned error: %v", err)
	}

	empty := &adapter{}
	if err := empty.Close(); err != nil {
		t.Fatalf("empty adapter close returned error: %v", err)
	}
}

func TestAdapter_Readiness_NilSafe(t *testing.T) {
	var nilAdapter *adapter
	if nilAdapter.Readiness(nil) {
		t.Fatal("expected nil adapter readiness to be false")
	}

	empty := &adapter{}
	if empty.Readiness(nil) {
		t.Fatal("expected empty adapter readiness to be false")
	}
}

func TestSanitizeDSN_MasksPassword(t *testing.T) {
	maskedURL := sanitizeDSN("postgresql://user:secret@localhost:5432/postgres?sslmode=disable")
	if maskedURL == "postgresql://user:secret@localhost:5432/postgres?sslmode=disable" {
		t.Fatalf("expected url password masked, got %q", maskedURL)
	}

	maskedKV := sanitizeDSN("host=localhost user=postgres password=secret dbname=postgres sslmode=disable")
	if maskedKV == "host=localhost user=postgres password=secret dbname=postgres sslmode=disable" {
		t.Fatalf("expected kv password masked, got %q", maskedKV)
	}
}
