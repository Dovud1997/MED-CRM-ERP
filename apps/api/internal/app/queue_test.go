package app

import (
	"strings"
	"testing"
)

func TestDisplayTokenIsRandomAndNotStoredAsPlaintext(t *testing.T) {
	first, second := newDisplayToken(), newDisplayToken()
	if len(first) != 64 || len(second) != 64 { t.Fatalf("expected 256-bit hex tokens") }
	if first == second { t.Fatalf("tokens must be random") }
	hash := tokenHash(first)
	if len(hash) != 64 || hash == first || strings.Contains(hash, first) { t.Fatalf("token must be stored only as an irreversible hash") }
}

func TestMediaExtensionsAreDerivedFromDetectedContentType(t *testing.T) {
	cases := map[string]string{"audio/mpeg":".mp3","audio/wav":".wav","video/mp4":".mp4","video/webm":".webm","image/jpeg":".jpg","image/png":".png","image/webp":".webp"}
	for contentType, expected := range cases { if got:=mediaExtension(contentType);got!=expected { t.Errorf("%s: got %s, want %s",contentType,got,expected) } }
	if got:=mediaExtension("text/html");got!=".bin" { t.Fatalf("unsafe unknown type must not keep client extension") }
}

func TestQueueMediaAllowListDoesNotPermitExecutableOrHTML(t *testing.T) {
	for kind, allowed := range queueMediaTypes { if allowed["text/html"] || allowed["application/x-msdownload"] { t.Fatalf("unsafe content allowed for %s",kind) } }
}
