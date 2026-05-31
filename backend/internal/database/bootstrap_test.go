package database

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	const unambiguous = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"

	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		pw, err := generatePassword()
		if err != nil {
			t.Fatalf("generatePassword: %v", err)
		}
		// Shape: 4 groups of 4 joined by dashes → "xxxx-xxxx-xxxx-xxxx" (19 chars).
		if pw == "" {
			t.Fatal("empty password")
		}
		groups := strings.Split(pw, "-")
		if len(groups) != 4 {
			t.Fatalf("want 4 groups, got %d in %q", len(groups), pw)
		}
		for _, g := range groups {
			if len(g) != 4 {
				t.Fatalf("group %q not 4 chars in %q", g, pw)
			}
			for _, r := range g {
				if !strings.ContainsRune(unambiguous, r) {
					t.Fatalf("char %q outside unambiguous alphabet in %q", r, pw)
				}
			}
		}
		if seen[pw] {
			t.Fatalf("duplicate password generated: %q", pw)
		}
		seen[pw] = true
	}
}

func TestSlugifyBootstrap(t *testing.T) {
	cases := map[string]string{
		"Minha Organização": "minha-organiza-o",
		"  Acme Inc.  ":     "acme-inc",
		"@@@###":            "org",
		"":                  "org",
		"Org 2024":          "org-2024",
	}
	for in, want := range cases {
		if got := slugifyBootstrap(in); got != want {
			t.Errorf("slugifyBootstrap(%q) = %q, want %q", in, got, want)
		}
	}
}
