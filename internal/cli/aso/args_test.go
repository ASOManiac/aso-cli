package aso

import (
	"flag"
	"strings"
	"testing"
)

func TestResolveArgsInterspersedFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "")

	args := []string{"photo editor", "--storefront", "GB"}
	keywords := resolveArgs(fs, args, true)

	if len(keywords) != 1 || keywords[0] != "photo editor" {
		t.Errorf("keywords = %v, want [photo editor]", keywords)
	}
	if *storefront != "GB" {
		t.Errorf("storefront = %q, want GB", *storefront)
	}
}

func TestResolveArgsCommaExpansion(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String("storefront", "US", "")

	args := []string{"photo editor,camera,vpn"}
	keywords := resolveArgs(fs, args, true)

	want := []string{"photo editor", "camera", "vpn"}
	if len(keywords) != len(want) {
		t.Fatalf("keywords = %v, want %v", keywords, want)
	}
	for i, kw := range keywords {
		if kw != want[i] {
			t.Errorf("keywords[%d] = %q, want %q", i, kw, want[i])
		}
	}
}

func TestResolveArgsMixedCommaAndFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "")
	fields := fs.String("fields", "", "")

	args := []string{"photo editor,camera", "--storefront", "GB", "vpn", "--fields", "popularity,difficulty"}
	keywords := resolveArgs(fs, args, true)

	if *storefront != "GB" {
		t.Errorf("storefront = %q, want GB", *storefront)
	}
	if *fields != "popularity,difficulty" {
		t.Errorf("fields = %q, want popularity,difficulty", *fields)
	}
	want := []string{"photo editor", "camera", "vpn"}
	if len(keywords) != len(want) {
		t.Fatalf("keywords = %v, want %v", keywords, want)
	}
	for i, kw := range keywords {
		if kw != want[i] {
			t.Errorf("keywords[%d] = %q, want %q", i, kw, want[i])
		}
	}
}

func TestResolveArgsNoExpand(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "")

	args := []string{"123456789", "--storefront", "GB"}
	positional := resolveArgs(fs, args, false)

	if len(positional) != 1 || positional[0] != "123456789" {
		t.Errorf("positional = %v, want [123456789]", positional)
	}
	if *storefront != "GB" {
		t.Errorf("storefront = %q, want GB", *storefront)
	}
}

func TestResolveArgsFlagEqualsStyle(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "")

	args := []string{"camera", "--storefront=GB"}
	keywords := resolveArgs(fs, args, true)

	if len(keywords) != 1 || keywords[0] != "camera" {
		t.Errorf("keywords = %v, want [camera]", keywords)
	}
	if *storefront != "GB" {
		t.Errorf("storefront = %q, want GB", *storefront)
	}
}

func TestResolveArgsFlagsBeforePositional(t *testing.T) {
	// When flags come first (normal Go flag behavior), resolveArgs should
	// still work correctly — the flag values are already parsed by the
	// FlagSet, so resolveArgs just passes through the positional args.
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String("storefront", "US", "")

	// Simulate: flags already parsed, only positional args remain.
	args := []string{"camera", "photo"}
	keywords := resolveArgs(fs, args, true)

	if len(keywords) != 2 || keywords[0] != "camera" || keywords[1] != "photo" {
		t.Errorf("keywords = %v, want [camera photo]", keywords)
	}
}

func TestResolveArgsEmptyCommas(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	args := []string{"camera,,photo,"}
	keywords := resolveArgs(fs, args, true)

	if len(keywords) != 2 {
		t.Fatalf("keywords = %v, want [camera photo]", keywords)
	}
	if keywords[0] != "camera" || keywords[1] != "photo" {
		t.Errorf("keywords = %v, want [camera photo]", keywords)
	}
}

func TestResolveArgsUnknownFlagTreatedAsPositional(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String("storefront", "US", "")

	// --unknown is not a registered flag, should be treated as positional.
	args := []string{"camera", "--unknown", "value"}
	keywords := resolveArgs(fs, args, true)

	if len(keywords) != 3 {
		t.Fatalf("keywords = %v, want [camera --unknown value]", keywords)
	}
	if !strings.HasPrefix(keywords[1], "--unknown") {
		t.Errorf("keywords[1] = %q, want --unknown", keywords[1])
	}
}
