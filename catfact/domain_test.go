package catfact

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "catfact" {
		t.Errorf("Scheme = %q, want catfact", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "catfact" {
		t.Errorf("Identity.Binary = %q, want catfact", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify empty string should return error")
	}

	typ, id, err := Domain{}.Classify("some-fact")
	if err != nil {
		t.Errorf("Classify: unexpected error: %v", err)
	}
	if typ != "fact" {
		t.Errorf("Classify type = %q, want fact", typ)
	}
	if id != "some-fact" {
		t.Errorf("Classify id = %q, want some-fact", id)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("fact", "anything")
	if err != nil {
		t.Fatalf("Locate: unexpected error: %v", err)
	}
	want := "https://catfact.ninja/fact"
	if got != want {
		t.Errorf("Locate = %q, want %q", got, want)
	}

	_, err = Domain{}.Locate("unknown", "x")
	if err == nil {
		t.Error("Locate with unknown type should return error")
	}
}
