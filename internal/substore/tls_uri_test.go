package substore

import (
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeCertSHA256(t *testing.T) {
	in := "E8:E2:D3:87:FD:BF:EB:38"
	want := "e8e2d387fdbfeb38"
	if got := NormalizeCertSHA256(in); got != want {
		t.Fatalf("NormalizeCertSHA256 = %q, want %q", got, want)
	}
}

func TestApplySkipCertVerifyURI_NoAllowInsecure(t *testing.T) {
	params := url.Values{}
	proxy := Proxy{
		"skip-cert-verify": true,
		"tls":              true,
	}
	applySkipCertVerifyURI(params, proxy)
	if params.Get("allowInsecure") != "" {
		t.Fatalf("expected no allowInsecure, got %q", params.Get("allowInsecure"))
	}
}

func TestApplySkipCertVerifyURI_PinSHA256(t *testing.T) {
	params := url.Values{}
	proxy := Proxy{
		"skip-cert-verify": true,
		"tls-fingerprint":  "AA:BB:CC",
	}
	applySkipCertVerifyURI(params, proxy)
	if params.Get("allowInsecure") != "" {
		t.Fatalf("expected no allowInsecure")
	}
	if params.Get("pinSHA256") != "aabbcc" {
		t.Fatalf("pinSHA256 = %q, want aabbcc", params.Get("pinSHA256"))
	}
}

func TestEncodeVLESS_SkipVerifyWithoutAllowInsecure(t *testing.T) {
	p := NewURIProducer()
	uri, err := p.ProduceOne(Proxy{
		"name":              "test",
		"type":              "vless",
		"server":            "example.com",
		"port":              443,
		"uuid":              "00000000-0000-0000-0000-000000000001",
		"tls":               true,
		"skip-cert-verify":  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uri, "allowInsecure=1") {
		t.Fatalf("vless URI must not contain allowInsecure=1: %s", uri)
	}
}

func TestEncodeVLESS_SkipVerifyWithFingerprint(t *testing.T) {
	p := NewURIProducer()
	uri, err := p.ProduceOne(Proxy{
		"name":              "test",
		"type":              "vless",
		"server":            "example.com",
		"port":              443,
		"uuid":              "00000000-0000-0000-0000-000000000001",
		"tls":               true,
		"skip-cert-verify":  true,
		"tls-fingerprint":   "AABBCC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uri, "allowInsecure=1") {
		t.Fatalf("vless URI must not contain allowInsecure=1: %s", uri)
	}
	if !strings.Contains(uri, "pinSHA256=aabbcc") {
		t.Fatalf("vless URI should contain pinSHA256: %s", uri)
	}
}

func TestSingboxTLS_SkipVerifyUsesCertPin(t *testing.T) {
	producer := NewSingboxProducer()
	proxy := Proxy{
		"name":             "ld-node",
		"type":             "trojan",
		"server":           "ld.example.com",
		"port":             443,
		"password":         "secret",
		"skip-cert-verify": true,
		"tls-fingerprint":  "DEADBEEF",
	}
	parsed, err := producer.trojanParser(proxy)
	if err != nil {
		t.Fatal(err)
	}
	tls, ok := parsed["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tls map, got %T", parsed["tls"])
	}
	if tls["insecure"] == true {
		t.Fatal("expected certificate pin instead of insecure")
	}
	pins, ok := tls["certificate_public_key_sha256"].([]string)
	if !ok || len(pins) == 0 || pins[0] != "deadbeef" {
		t.Fatalf("unexpected certificate_public_key_sha256: %v", tls["certificate_public_key_sha256"])
	}
}
