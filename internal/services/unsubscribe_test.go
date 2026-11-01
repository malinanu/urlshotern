package services

import "testing"

func TestUnsubscribeTokenRoundTrip(t *testing.T) {
	token := UnsubscribeToken("secret", "  Jane.Doe@Example.com ")
	addr, err := ParseUnsubscribeToken("secret", token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "jane.doe@example.com" {
		t.Fatalf("got %q, want normalized address", addr)
	}
}

func TestUnsubscribeTokenRejectsTampering(t *testing.T) {
	good := UnsubscribeToken("secret", "jane@example.com")
	other := UnsubscribeToken("secret", "victim@example.com")

	cases := map[string]string{
		"wrong secret":      UnsubscribeToken("other-secret", "jane@example.com"),
		"swapped signature": other[:len(other)-64] + good[len(good)-64:],
		"truncated":         good[:len(good)-2],
		"no separator":      "amFuZUBleGFtcGxlLmNvbQ",
		"empty":             "",
		"garbage base64":    "!!!." + good[len(good)-64:],
	}
	for name, token := range cases {
		if _, err := ParseUnsubscribeToken("secret", token); err == nil {
			t.Errorf("%s: expected token to be rejected", name)
		}
	}
}
