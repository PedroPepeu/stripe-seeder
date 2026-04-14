package stripe

import (
	"strings"
	"testing"
)

func TestRandomProductName_Format(t *testing.T) {
	for i := 0; i < 50; i++ {
		name := randomProductName()
		parts := strings.SplitN(name, " ", 2)
		if len(parts) != 2 {
			t.Fatalf("expected 'Adj Noun', got %q", name)
		}
		if parts[0] == "" || parts[1] == "" {
			t.Fatalf("empty part in name %q", name)
		}
	}
}

func TestRandomProductName_UsesKnownWords(t *testing.T) {
	adjSet := make(map[string]bool, len(productAdjectives))
	for _, a := range productAdjectives {
		adjSet[a] = true
	}
	nounSet := make(map[string]bool, len(productNouns))
	for _, n := range productNouns {
		nounSet[n] = true
	}

	for i := 0; i < 100; i++ {
		name := randomProductName()
		parts := strings.SplitN(name, " ", 2)
		if !adjSet[parts[0]] {
			t.Errorf("unknown adjective %q in name %q", parts[0], name)
		}
		if !nounSet[parts[1]] {
			t.Errorf("unknown noun %q in name %q", parts[1], name)
		}
	}
}

func TestRandomDescription_NonEmpty(t *testing.T) {
	for i := 0; i < 20; i++ {
		d := randomDescription()
		if d == "" {
			t.Fatal("got empty description")
		}
	}
}

func TestRandomEmail_Format(t *testing.T) {
	cases := []struct{ first, last string }{
		{"Alice", "Silva"},
		{"João", "Costa"},
		{"Maria", "Oliveira"},
	}
	for _, c := range cases {
		email := randomEmail(c.first, c.last)
		if !strings.Contains(email, "@") {
			t.Errorf("no @ in email %q", email)
		}
		parts := strings.Split(email, "@")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			t.Errorf("malformed email %q", email)
		}
		if !strings.Contains(parts[0], ".") {
			t.Errorf("local part missing dot in %q", email)
		}
		found := false
		for _, d := range domains {
			if parts[1] == d {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unknown domain %q in email %q", parts[1], email)
		}
	}
}

func TestRandomPrice_InRange(t *testing.T) {
	tests := []struct {
		min, max int64
	}{
		{100, 100},
		{100, 10000},
		{500, 50000},
	}
	for _, tt := range tests {
		for i := 0; i < 100; i++ {
			p := randomPrice(tt.min, tt.max)
			if p < tt.min || p > tt.max {
				t.Errorf("price %d out of [%d, %d]", p, tt.min, tt.max)
			}
		}
	}
}

func TestSeedResult_ZeroValue(t *testing.T) {
	var r SeedResult
	if r.Created != 0 || r.Errors != 0 || r.Details != nil {
		t.Errorf("unexpected zero value: %+v", r)
	}
}
