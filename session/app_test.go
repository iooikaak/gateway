package session

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	uid := "1"
	token := "123456"
	timestamp := "1634554282"
	k := fmt.Sprintf("warhorse:%s:TOKEN:%s:CT:%s", uid, token, timestamp)
	s := sha1.New()
	io.WriteString(s, k)
	t.Logf("token is %s", strings.ToUpper(hex.EncodeToString(s.Sum(nil))))
}
