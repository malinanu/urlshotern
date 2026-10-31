package services

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// fakeLimiter counts IncrWindow calls per key and records SetNX keys.
type fakeLimiter struct {
	counts map[string]int64
	nx     map[string]bool
	vals   map[string]string
}

func newFakeLimiter() *fakeLimiter {
	return &fakeLimiter{counts: map[string]int64{}, nx: map[string]bool{}, vals: map[string]string{}}
}

func (f *fakeLimiter) IncrWindow(key string, _ time.Duration) (int64, error) {
	f.counts[key]++
	return f.counts[key], nil
}

func (f *fakeLimiter) SetNX(key string, _ interface{}, _ time.Duration) (bool, error) {
	if f.nx[key] {
		return false, nil
	}
	f.nx[key] = true
	return true, nil
}

func (f *fakeLimiter) Get(key string) (string, error) { return f.vals[key], nil }
func (f *fakeLimiter) Del(keys ...string) error {
	for _, k := range keys {
		delete(f.vals, k)
	}
	return nil
}

type fakeMailer struct{ sent int }

func (m *fakeMailer) SendEmail(_, _, _, _ string) error { m.sent++; return nil }

func formPage(t *testing.T, kind string, enabled bool) *PublishedMicrosite {
	t.Helper()
	return &PublishedMicrosite{
		ID: 42, UserID: 7, Slug: "shop", Title: "Shop", FormsEnabled: enabled,
		Content: models.MicrositeContent{Blocks: []models.MicrositeBlock{
			{ID: "f1", Type: models.BlockTypeForm, SchemaVersion: 1, Data: mustJSON(t, models.FormData{Kind: kind, Notify: true})},
			{ID: "f2", Type: models.BlockTypeForm, SchemaVersion: 1, Hidden: true, Data: mustJSON(t, models.FormData{Kind: kind})},
		}},
	}
}

func newTestFormService(l formLimiter, m formMailer) *MicrositeFormService {
	s := NewMicrositeFormService(nil, nil, nil, "test-secret", "https://trunc.test/")
	s.limiter, s.mailer = l, m
	return s
}

func TestFormTokenLifetime(t *testing.T) {
	s := newTestFormService(nil, nil)
	rendered := time.Unix(1_800_000_000, 0)
	tok := s.Token(42, "f1", rendered)

	cases := []struct {
		name  string
		page  int64
		block string
		token string
		at    time.Time
		want  string
	}{
		{"valid after a few seconds", 42, "f1", tok, rendered.Add(10 * time.Second), ""},
		{"too fast", 42, "f1", tok, rendered.Add(time.Second), FormErrTooFast},
		{"expired", 42, "f1", tok, rendered.Add(25 * time.Hour), FormErrExpired},
		{"other page", 43, "f1", tok, rendered.Add(10 * time.Second), FormErrInvalid},
		{"other block", 42, "f2", tok, rendered.Add(10 * time.Second), FormErrInvalid},
		{"tampered time", 42, "f1", "1799999000" + tok[strings.Index(tok, "."):], rendered.Add(10 * time.Second), FormErrInvalid},
		{"garbage", 42, "f1", "nope", rendered, FormErrInvalid},
		{"empty", 42, "f1", "", rendered, FormErrInvalid},
	}
	for _, tc := range cases {
		if got := s.VerifyToken(tc.page, tc.block, tc.token, tc.at); got != tc.want {
			t.Errorf("%s: VerifyToken = %q, want %q", tc.name, got, tc.want)
		}
	}

	other := NewMicrositeFormService(nil, nil, nil, "different-secret", "")
	if got := other.VerifyToken(42, "f1", tok, rendered.Add(10*time.Second)); got != FormErrInvalid {
		t.Errorf("token from another secret accepted: %q", got)
	}
}

// Every case here is rejected (or dropped) before the database is touched,
// so the service runs with a nil DB.
func TestFormSubmitRejections(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	s := newTestFormService(newFakeLimiter(), nil)
	good := s.Token(42, "f1", now.Add(-time.Minute))
	in := func(mod func(*SubmissionInput)) SubmissionInput {
		v := SubmissionInput{BlockID: "f1", Token: good, Email: "ana@example.com", Name: "Ana", Message: "Hi", VisitorHash: "v", Now: now}
		mod(&v)
		return v
	}

	cases := []struct {
		name   string
		page   *PublishedMicrosite
		input  SubmissionInput
		code   string
		stored bool
	}{
		{"forms off for plan", formPage(t, "newsletter", false), in(func(*SubmissionInput) {}), FormErrClosed, false},
		{"unknown block", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.BlockID = "nope" }), FormErrInvalid, false},
		{"hidden block", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.BlockID = "f2" }), FormErrInvalid, false},
		{"bad token", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.Token = "1.abc" }), FormErrInvalid, false},
		{"too fast", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.Token = s.Token(42, "f1", now) }), FormErrTooFast, false},
		{"bad email", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.Email = "not-an-email" }), FormErrInvalidEmail, false},
		{"missing email", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.Email = "" }), FormErrInvalidEmail, false},
		{"header injection", formPage(t, "newsletter", true), in(func(v *SubmissionInput) { v.Email = "a@example.com\r\nBcc: x@evil.test" }), FormErrInvalidEmail, false},
		{"contact without message", formPage(t, "contact", true), in(func(v *SubmissionInput) { v.Message = "  " }), FormErrMissingMessage, false},
		{"contact message too long", formPage(t, "contact", true), in(func(v *SubmissionInput) { v.Message = strings.Repeat("x", 2001) }), FormErrInvalid, false},
	}
	for _, tc := range cases {
		code, res, err := s.Submit(tc.page, tc.input)
		if err != nil {
			t.Errorf("%s: unexpected error %v", tc.name, err)
			continue
		}
		if code != tc.code {
			t.Errorf("%s: code = %q, want %q", tc.name, code, tc.code)
		}
		if res != nil && res.Stored != tc.stored {
			t.Errorf("%s: stored = %v", tc.name, res.Stored)
		}
	}
}

func TestFormHoneypotLooksLikeSuccess(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	s := newTestFormService(newFakeLimiter(), nil)
	// Bots that fill the hidden field get a success result and nothing is stored,
	// even without a valid token.
	code, res, err := s.Submit(formPage(t, "newsletter", true), SubmissionInput{
		BlockID: "f1", Honeypot: "https://spam.test", Email: "bot@spam.test", Now: now,
	})
	if err != nil || code != "" || res == nil || res.Stored {
		t.Fatalf("honeypot: code=%q res=%+v err=%v", code, res, err)
	}
}

func TestFormRateLimit(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	lim := newFakeLimiter()
	s := newTestFormService(lim, nil)
	page := formPage(t, "newsletter", true)
	// Pretend this visitor already posted formIPLimit times in the window.
	lim.counts["microsite:form:ip:42:v"] = formIPLimit
	code, _, err := s.Submit(page, SubmissionInput{
		BlockID: "f1", Token: s.Token(42, "f1", now.Add(-time.Minute)), Email: "ana@example.com", VisitorHash: "v", Now: now,
	})
	if err != nil || code != FormErrRateLimited {
		t.Fatalf("per-visitor limit: code=%q err=%v", code, err)
	}

	lim = newFakeLimiter()
	s = newTestFormService(lim, nil)
	lim.counts["microsite:form:day:42:"+now.UTC().Format("20060102")] = formDailyLimit
	code, _, err = s.Submit(page, SubmissionInput{
		BlockID: "f1", Token: s.Token(42, "f1", now.Add(-time.Minute)), Email: "ana@example.com", VisitorHash: "w", Now: now,
	})
	if err != nil || code != FormErrRateLimited {
		t.Fatalf("daily page cap: code=%q err=%v", code, err)
	}
}

func TestFormNotifyThrottle(t *testing.T) {
	lim := newFakeLimiter()
	mail := &fakeMailer{}
	s := newTestFormService(lim, mail)
	page := formPage(t, "newsletter", true)
	res := &SubmitResult{Stored: true, Submission: models.MicrositeSubmission{BlockID: "f1"}, Form: models.FormData{Notify: true}}

	// The throttle key is taken by an earlier alert: this one is only counted.
	lim.nx["microsite:form:notify:42:f1"] = true
	if err := s.Notify(page, res); err != nil {
		t.Fatal(err)
	}
	if mail.sent != 0 || lim.counts["microsite:form:notify:42:f1:pending"] != 1 {
		t.Fatalf("throttled alert: sent=%d pending=%d", mail.sent, lim.counts["microsite:form:notify:42:f1:pending"])
	}

	// Owner switched alerts off, or the submission was a dropped honeypot.
	off := *res
	off.Form.Notify = false
	dropped := *res
	dropped.Stored = false
	for _, r := range []*SubmitResult{&off, &dropped, nil} {
		if err := s.Notify(page, r); err != nil || mail.sent != 0 {
			t.Fatalf("alert sent when it should not be: %+v", r)
		}
	}
}

func TestCSVSafe(t *testing.T) {
	cases := map[string]string{
		"=HYPERLINK(\"http://x\")": "'=HYPERLINK(\"http://x\")",
		"+1 555":                   "'+1 555",
		"-2+3":                     "'-2+3",
		"@SUM(A1)":                 "'@SUM(A1)",
		"\tcmd":                    "'\tcmd",
		"\rcmd":                    "'\rcmd",
		"ana@example.com":          "ana@example.com",
		"":                         "",
		"Hello = world":            "Hello = world",
	}
	for in, want := range cases {
		if got := CSVSafe(in); got != want {
			t.Errorf("CSVSafe(%q) = %q, want %q", in, got, want)
		}
	}
	// The guarded value still round-trips as a single CSV cell.
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{CSVSafe("=1+1,\"x\"\n")})
	w.Flush()
	rec, err := csv.NewReader(strings.NewReader(b.String())).Read()
	if err != nil || len(rec) != 1 || rec[0] != "'=1+1,\"x\"\n" {
		t.Fatalf("round trip: %q %v", rec, err)
	}
}
