package envloader

import (
	"os"
	"testing"
)

// gopathDefs is used only in tests; GOPATH is not in DefaultDefs.
var gopathDefs = []EnvDef{{Name: "GOPATH", Type: TypeString, Required: true}}

func TestLoad_EmptyDefs_ReturnsConfig(t *testing.T) {
	os.Clearenv()
	cfg, msgs := Load()
	if cfg == nil {
		t.Fatal("expected config when DefaultDefs empty")
	}
	if len(cfg.Values) != 0 {
		t.Errorf("expected empty Values, got %v", cfg.Values)
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages, got %v", msgs)
	}
}

func TestLoadWith_GOPATHMissing_ReturnsNilAndMessages(t *testing.T) {
	os.Clearenv()
	cfg, msgs := LoadWith(gopathDefs)
	if cfg != nil {
		t.Error("expected nil config when GOPATH missing")
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Key != "GOPATH" || msgs[0].Kind != "missing" || msgs[0].Level != LevelError {
		t.Errorf("expected GOPATH missing error, got %v", msgs[0])
	}
}

func TestLoadWith_GOPATHEmpty_ReturnsNilAndMessages(t *testing.T) {
	os.Clearenv()
	os.Setenv("GOPATH", "")
	cfg, msgs := LoadWith(gopathDefs)
	if cfg != nil {
		t.Error("expected nil config when GOPATH empty")
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Kind != "empty" || msgs[0].Level != LevelError {
		t.Errorf("expected empty error, got %v", msgs[0])
	}
}

func TestLoadWith_GOPATHSet_ReturnsConfig(t *testing.T) {
	os.Clearenv()
	os.Setenv("GOPATH", "/home/user/go")
	cfg, msgs := LoadWith(gopathDefs)
	if cfg == nil {
		t.Fatal("expected config when GOPATH set")
	}
	if cfg.Values["GOPATH"] != "/home/user/go" {
		t.Errorf("GOPATH = %v, want /home/user/go", cfg.Values["GOPATH"])
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages, got %v", msgs)
	}
}

func TestLoadWith_CustomDefs(t *testing.T) {
	os.Clearenv()
	os.Setenv("FOO", "bar")
	os.Setenv("NUM", "42")
	os.Setenv("FLAG", "true")
	defs := []EnvDef{
		{Name: "FOO", Type: TypeString, Required: true},
		{Name: "NUM", Type: TypeInt, Required: false},
		{Name: "FLAG", Type: TypeBool, Required: false},
	}
	cfg, msgs := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Values["FOO"] != "bar" {
		t.Errorf("FOO = %v, want bar", cfg.Values["FOO"])
	}
	if cfg.Values["NUM"] != 42 {
		t.Errorf("NUM = %v, want 42", cfg.Values["NUM"])
	}
	if cfg.Values["FLAG"] != true {
		t.Errorf("FLAG = %v, want true", cfg.Values["FLAG"])
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages, got %v", msgs)
	}
}

func TestLoadWith_OptionalMissing_StillOk(t *testing.T) {
	os.Clearenv()
	os.Setenv("GOPATH", "/go")
	defs := append(gopathDefs, EnvDef{Name: "OPTIONAL", Type: TypeString, Required: false})
	cfg, msgs := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages for optional missing, got %v", msgs)
	}
}

func TestExists_LookupBeforeGetenv(t *testing.T) {
	os.Clearenv()
	if Exists("GOPATH") {
		t.Error("GOPATH should not exist")
	}
	os.Setenv("GOPATH", "/go")
	if !Exists("GOPATH") {
		t.Error("GOPATH should exist")
	}
}

func TestRead_AfterExists(t *testing.T) {
	os.Clearenv()
	os.Setenv("GOPATH", "/home/go")
	v := Read("GOPATH")
	if v != "/home/go" {
		t.Errorf("Read = %q, want /home/go", v)
	}
}

func TestEnvMessage_String(t *testing.T) {
	m := EnvMessage{Key: "FOO", Kind: KindMissing, Message: "FOO missing"}
	s := m.String()
	if s == "" || len(s) < 10 {
		t.Errorf("String() = %q", s)
	}
}

func TestLoadWith_OptionalWithDefault_UsesDefault(t *testing.T) {
	os.Clearenv()
	defs := []EnvDef{
		{Name: "APP_ENV", Type: TypeString, Required: false, Default: "development"},
	}
	cfg, msgs := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages, got %v", msgs)
	}
	if cfg.GetString("APP_ENV") != "development" {
		t.Errorf("APP_ENV = %q, want development", cfg.GetString("APP_ENV"))
	}
}

func TestLoadWith_OptionalEmpty_UsesDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "")
	defs := []EnvDef{
		{Name: "APP_ENV", Type: TypeString, Required: false, Default: "development"},
	}
	cfg, _ := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.GetString("APP_ENV") != "development" {
		t.Errorf("APP_ENV = %q, want development", cfg.GetString("APP_ENV"))
	}
}

func TestLoadWith_OptionalInvalid_UsesDefaultAndWarning(t *testing.T) {
	os.Clearenv()
	os.Setenv("NUM", "not-a-number")
	defs := []EnvDef{
		{Name: "NUM", Type: TypeInt, Required: false, Default: 42},
	}
	cfg, msgs := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Values["NUM"] != 42 {
		t.Errorf("NUM = %v, want 42", cfg.Values["NUM"])
	}
	var foundWarning bool
	for _, m := range msgs {
		if m.Level == LevelWarning {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("expected warning for optional invalid")
	}
}

func TestLoadWith_OptionalInvalidBool_UsesDefaultAndWarning(t *testing.T) {
	os.Clearenv()
	os.Setenv("FEATURE", "not-bool")
	defs := []EnvDef{
		{Name: "FEATURE", Type: TypeBool, Required: false, Default: true},
	}
	cfg, msgs := LoadWith(defs)
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Values["FEATURE"] != true {
		t.Errorf("FEATURE = %v, want true", cfg.Values["FEATURE"])
	}
	var foundWarning bool
	for _, m := range msgs {
		if m.Level == LevelWarning {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("expected warning for optional invalid bool")
	}
}
