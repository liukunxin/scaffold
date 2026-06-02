package main

import "testing"

func TestParseInitFeaturesArg_EmptyMeansNone(t *testing.T) {
	flags, withLLM, err := parseInitFeaturesArg("")
	if err != nil {
		t.Fatal(err)
	}
	if withLLM {
		t.Fatal("expected llm false")
	}
	for name, enabled := range flags {
		if enabled {
			t.Fatalf("expected %s false by default", name)
		}
	}
}

func TestParseInitFeaturesArg_Selective(t *testing.T) {
	flags, withLLM, err := parseInitFeaturesArg("mysql,redis,llm")
	if err != nil {
		t.Fatal(err)
	}
	if !withLLM || !flags["mysql"] || !flags["redis"] || flags["metrics"] {
		t.Fatalf("unexpected flags: %+v llm=%v", flags, withLLM)
	}
}

func TestParseConfigFeaturesArg_RejectsLLM(t *testing.T) {
	_, err := parseConfigFeaturesArg("llm")
	if err == nil {
		t.Fatal("expected error for llm")
	}
}
