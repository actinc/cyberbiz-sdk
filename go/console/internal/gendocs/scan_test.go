package gendocs

import "testing"

func TestScanOutput(t *testing.T) {
	bad := []string{
		"contact john@gmail.com",
		"call 0933123456 now",
		"token eyJhbGciOiJIUzI1NiJ9.eyJzaG9wIjoxfQ",
		"name: REDACTED",
		"host acme.cyberbiz.co",
		"confirmation_token=xxxxxxxxxxxxxxxxxxxx",
	}
	for _, b := range bad {
		if err := scanOutput("x", []byte(b)); err == nil {
			t.Errorf("%q should be rejected", b)
		}
	}
	good := []string{
		"customer@example.com 0912345678 https://example.cyberbiz.co",
		"price 3690.0933786519886",
		"confirmation_token=SYNTHETIC",
	}
	for _, g := range good {
		if err := scanOutput("x", []byte(g)); err != nil {
			t.Errorf("%q should pass: %v", g, err)
		}
	}
}
