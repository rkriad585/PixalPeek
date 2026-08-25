package qrengine

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateDecodeRoundTrip(t *testing.T) {
	payload := "https://github.com/rkriad585/PixalPeek"
	data, err := Generate(EncodeOptions{Content: payload, Size: 512, ECC: "M", Format: FormatPNG})
	if err != nil {
		t.Fatalf("generate png: %v", err)
	}
	res := DecodeBytes(data, "test.png", false)
	if !res.Success {
		t.Fatalf("decode failed: %s", res.Error)
	}
	if res.Results[0].Content != payload {
		t.Fatalf("content mismatch: %q", res.Results[0].Content)
	}
	if res.Results[0].ContentType != TypeURL {
		t.Fatalf("type mismatch: %q", res.Results[0].ContentType)
	}
}

func TestGenerateAllFormats(t *testing.T) {
	cases := map[string]func([]byte) bool{
		FormatPNG: func(b []byte) bool { return len(b) > 4 && b[0] == 0x89 && b[1] == 'P' },
		FormatJPG: func(b []byte) bool { return len(b) > 3 && b[0] == 0xFF && b[1] == 0xD8 },
		FormatSVG: func(b []byte) bool { return strings.Contains(string(b), "<svg") },
		FormatPDF: func(b []byte) bool { return strings.HasPrefix(string(b), "%PDF-1.4") },
	}
	for format, check := range cases {
		data, err := Generate(EncodeOptions{Content: "FORMAT-" + format, Size: 256, ECC: "H", Format: format})
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		if !check(data) {
			t.Fatalf("%s: bad output header (%d bytes)", format, len(data))
		}
	}
}

func TestShapesDecode(t *testing.T) {
	for _, shape := range AllShapes {
		data, err := Generate(EncodeOptions{Content: "SHAPE-" + shape, Size: 480, ECC: "Q", Shape: shape})
		if err != nil {
			t.Fatalf("%s generate: %v", shape, err)
		}
		res := DecodeBytes(data, shape+".png", false)
		if !res.Success || res.Results[0].Content != "SHAPE-"+shape {
			t.Fatalf("%s decode failed: %+v", shape, res.Error)
		}
	}
}

func TestLogoOverlayStillScans(t *testing.T) {
	logoPNG := solidTestPNG(t)
	b64 := "data:image/png;base64," + mustB64(t, logoPNG)
	data, err := Generate(EncodeOptions{Content: "LOGO-OVERLAY-TEST", Size: 600, ECC: "H", LogoB64: b64})
	if err != nil {
		t.Fatalf("logo generate: %v", err)
	}
	res := DecodeBytes(data, "logo.png", false)
	if !res.Success || res.Results[0].Content != "LOGO-OVERLAY-TEST" {
		t.Fatalf("logo decode failed: %v / %q", res.Error, first(t, res))
	}
}

func TestMultiDecodeFindsAll(t *testing.T) {
	a, _ := Generate(EncodeOptions{Content: "MULTI-FIRST-AAAA", Size: 300, ECC: "M"})
	b, _ := Generate(EncodeOptions{Content: "MULTI-SECOND-BBBB", Size: 300, ECC: "M"})
	combo, err := composeSideBySide(a, b)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	res := DecodeBytes(combo, "combo.png", true)
	if !res.Success {
		t.Fatalf("multi decode failed: %s", res.Error)
	}
	found := map[string]bool{}
	for _, r := range res.Results {
		found[r.Content] = true
	}
	if !found["MULTI-FIRST-AAAA"] || !found["MULTI-SECOND-BBBB"] {
		t.Fatalf("missing codes in multi result: %v", found)
	}
}

func TestDetectorClassification(t *testing.T) {
	cases := []struct {
		in   string
		want ContentType
	}{
		{"https://example.com/x?y=1", TypeURL},
		{"http://example.com", TypeURL},
		{"www.wikipedia.org", TypeURL},
		{"WIFI:T:WPA;S:Home;P:pass123;;", TypeWiFi},
		{"BEGIN:VCARD\nVERSION:3.0\nFN:A\nEND:VCARD", TypeVCard},
		{"BEGIN:VEVENT\nSUMMARY:x\nEND:VEVENT", TypeEvent},
		{"GEO:35.6812,51.3795", TypeGeo},
		{"mailto:user@host.com", TypeEmail},
		{"MATMSG:TO:a@b.c;SUB:j;BODY:m;;", TypeEmail},
		{"user@example.org", TypeEmail},
		{"SMSTO:+8801700000000:hi", TypeSMS},
		{"tel:+1234567890", TypePhone},
		{"+880 1712 345678", TypePhone},
		{"(030) 123-4567", TypePhone},
		{"just some plain text", TypeText},
		{"Meeting notes from Tuesday", TypeText},
	}
	for _, c := range cases {
		if got := DetectContentType(c.in); got != c.want {
			t.Errorf("DetectContentType(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestBuildContentWiFi(t *testing.T) {
	got, err := BuildContent(BuildContentRequest{Type: TypeWiFi, Fields: map[string]string{
		"ssid": "My;Net", "password": "p@ss\\word", "enc": "wpa",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := `WIFI:T:WPA;S:My\;Net;P:p@ss\\word;;`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildContentEventAndGeo(t *testing.T) {
	got, err := BuildContent(BuildContentRequest{Type: TypeGeo, Fields: map[string]string{"lat": "23.81", "lng": "90.41"}})
	if err != nil || got != "geo:23.81,90.41" {
		t.Fatalf("geo: %v %q", err, got)
	}
	ev, err := BuildContent(BuildContentRequest{Type: TypeEvent, Fields: map[string]string{
		"title": "Standup, weekly", "start": "2026-08-24T09:30",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := "BEGIN:VEVENT\r\nSUMMARY:Standup\\, weekly\r\nDTSTART:20260824T093000Z\r\nEND:VEVENT"
	if ev != want {
		t.Fatalf("event got:\n%q\nwant:\n%q", ev, want)
	}
}

func TestBatchCSVAndZip(t *testing.T) {
	dir := t.TempDir()
	csvPath := dir + "/batch.csv"
	content := "content,name\nhttps://one.example/first,alpha\nhttps://two.example/second,beta\n"
	if err := writeFile(csvPath, content); err != nil {
		t.Fatal(err)
	}
	entries, err := ParseBatchFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name != "alpha" {
		t.Fatalf("parse: %+v", entries)
	}

	outDir := dir + "/out"
	opts := EncodeOptions{Size: 240, ECC: "M", Format: FormatPNG, QuietZone: 4}
	written, failures, err := BatchGenerate(entries, opts, outDir, true, func(string, ...interface{}) {})
	if err != nil {
		t.Fatal(err)
	}
	if failures != 0 || len(written) != 2 {
		t.Fatalf("written=%d failures=%d", len(written), failures)
	}
	assertFileExists(t, outDir+"/alpha.png")
	assertFileExists(t, outDir+"/beta.png")
	assertFileExists(t, outDir+".zip")
}

func mustB64(t *testing.T, data []byte) string {
	t.Helper()
	return b64encode(data)
}

func first(t *testing.T, r ScanResponse) string {
	t.Helper()
	if len(r.Results) > 0 {
		return r.Results[0].Content
	}
	return ""
}

func TestBuildContentSocial(t *testing.T) {
	cases := []struct {
		platform string
		handle   string
		want     string
	}{
		{"x", "rkriad585", "https://x.com/rkriad585"},
		{"instagram", "testuser", "https://instagram.com/testuser"},
		{"github", "octocat", "https://github.com/octocat"},
		{"linkedin", "jdoe", "https://linkedin.com/in/jdoe"},
		{"youtube", "mychannel", "https://youtube.com/@mychannel"},
		{"tiktok", "dancer", "https://tiktok.com/@dancer"},
		{"telegram", "user123", "https://t.me/user123"},
		{"whatsapp", "+15551234567", "https://wa.me/15551234567"},
	}
	for _, tc := range cases {
		got, err := BuildContent(BuildContentRequest{
			Type:   TypeSocial,
			Fields: map[string]string{"platform": tc.platform, "handle": tc.handle},
		})
		if err != nil {
			t.Errorf("social %s: %v", tc.platform, err)
			continue
		}
		if got != tc.want {
			t.Errorf("social %s = %q; want %q", tc.platform, got, tc.want)
		}
	}
}

func TestSocialGenerateAndDecode(t *testing.T) {
	payload := "https://x.com/rkriad585"
	data, err := Generate(EncodeOptions{Content: payload, Size: 256, ECC: "M", Format: FormatPNG})
	if err != nil {
		t.Fatal(err)
	}
	res := DecodeBytes(data, "social.png", false)
	if !res.Success {
		t.Fatalf("decode failed: %s", res.Error)
	}
	if res.Results[0].Content != payload {
		t.Errorf("content = %q; want %q", res.Results[0].Content, payload)
	}
}

func TestBuildContentAllTypes(t *testing.T) {
	requests := []BuildContentRequest{
		{Type: TypeText, Fields: map[string]string{"text": "hello"}},
		{Type: TypeURL, Fields: map[string]string{"url": "https://example.com"}},
		{Type: TypeWiFi, Fields: map[string]string{"ssid": "Net", "password": "Pass"}},
		{Type: TypeEmail, Fields: map[string]string{"to": "a@b.com"}},
		{Type: TypeSMS, Fields: map[string]string{"number": "+1234", "message": "hi"}},
		{Type: TypePhone, Fields: map[string]string{"number": "+1234"}},
		{Type: TypeGeo, Fields: map[string]string{"lat": "40.7", "lng": "-74.0"}},
		{Type: TypeEvent, Fields: map[string]string{"title": "Meet", "start": "2026-09-01 10:00"}},
		{Type: TypeSocial, Fields: map[string]string{"platform": "github", "handle": "test"}},
	}
	for _, req := range requests {
		_, err := BuildContent(req)
		if err != nil {
			t.Errorf("BuildContent(%s): %v", req.Type, err)
		}
	}
}

func TestURLSecurityAssessment(t *testing.T) {
	cases := []struct {
		url          string
		suspicious   bool
		minScore     int
		expectReason string
	}{
		{"https://example.com", false, 0, ""},
		{"https://bit.ly/abc123", true, 35, "shortener"},
		{"http://192.168.1.1/login", true, 50, "ip"},
		{"https://xn--e1afmapc.xn--p1ai/", true, 40, "punycode"},
		{"https://suspicious.top/download", false, 20, "top-level domain"},
		{"http://site.com/page?auth=token123", true, 45, "http scheme"},
	}
	for _, tc := range cases {
		a := AssessURLSecurity(tc.url)
		if a.IsSuspicious != tc.suspicious {
			t.Errorf("%s: IsSuspicious=%v score=%d reasons=%v", tc.url, a.IsSuspicious, a.Score, a.Reasons)
		}
		if tc.minScore > 0 && a.Score < tc.minScore {
			t.Errorf("%s: score=%d want >= %d", tc.url, a.Score, tc.minScore)
		}
		if tc.expectReason != "" && len(a.Reasons) > 0 {
			found := false
			for _, r := range a.Reasons {
				if strings.Contains(strings.ToLower(r), tc.expectReason) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: expected reason containing %q, got %v", tc.url, tc.expectReason, a.Reasons)
			}
		}
	}
}

func TestURLSecurityInDecodeResult(t *testing.T) {
	payload := "https://bit.ly/abc123"
	data, err := Generate(EncodeOptions{Content: payload, Size: 256, ECC: "M"})
	if err != nil {
		t.Fatal(err)
	}
	res := DecodeBytes(data, "shortlink.png", false)
	if !res.Success {
		t.Fatalf("decode failed: %s", res.Error)
	}
	if res.Results[0].Security == nil {
		t.Fatal("expected Security assessment on URL result")
	}
	if !res.Results[0].Security.IsSuspicious {
		t.Error("expected IsSuspicious=true for bit.ly URL")
	}
	if !res.Results[0].Security.IsShortener {
		t.Error("expected IsShortener=true for bit.ly URL")
	}
}

func TestDecodeDirectory(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		payload := "https://example.com/dir-test-" + string(rune('A'+i))
		data, err := Generate(EncodeOptions{Content: payload, Size: 200, ECC: "M", Format: FormatPNG})
		if err != nil {
			t.Fatal(err)
		}
		path := dir + "/qr_" + string(rune('A'+i)) + ".png"
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := DecodeDirectory(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalFilesScanned != 3 {
		t.Errorf("TotalFilesScanned=%d want 3", result.TotalFilesScanned)
	}
	if result.SuccessfulDecodes != 3 {
		t.Errorf("SuccessfulDecodes=%d want 3", result.SuccessfulDecodes)
	}
	if len(result.Entries) != 3 {
		t.Errorf("Entries=%d want 3", len(result.Entries))
	}
	if result.DurationMs < 0 {
		t.Error("expected non-negative DurationMs")
	}
}
