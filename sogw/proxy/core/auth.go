package core

import (
	"encoding/base64"
	"strings"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

type Auth struct {
	Meta *meta.Auth
	Fn   func(Config map[string]string, ctx *RequestContext) bool
}

var authFnMap = map[meta.AuthKind]func(Config map[string]string, ctx *RequestContext) bool{
	meta.AuthKind_HttpBasic: doHttpBasic,
	meta.AuthKind_OAuth2:    doOAuth2,
}

func NewAuth(m *meta.Auth) *Auth {
	if m.Config == nil {
		m.Config = make(map[string]string)
	}
	a := &Auth{Meta: m}
	if fn, ok := authFnMap[m.Kind]; ok {
		a.Fn = fn
	}
	return a
}

func doHttpBasic(cfg map[string]string, ctx *RequestContext) bool {
	auth := reflectx.BytesToString(ctx.ReqCtx.Request.Header.Peek("Authentication"))
	if strings.HasPrefix(auth, "Basic ") || strings.HasPrefix(auth, "basic ") {
		auth = auth[6:]
		buf := make([]byte, base64.URLEncoding.DecodedLen(len(auth)))
		n, err := base64.URLEncoding.Decode(buf, reflectx.StringToBytes(auth))
		if err != nil {
			if cobrax.Flags.Debug {
				log.Error("fail to decode http basic value : ", err)
			}
			return false
		}
		auth = reflectx.BytesToString(buf[:n])
		user := auth
		var passwd string
		idx := strings.IndexByte(auth, ':')
		if idx >= 0 {
			user = auth[:idx]
			passwd = auth[idx+1:]
		}
		if pwd, ok := cfg[user]; ok && pwd == passwd {
			return true
		}
	}
	return false
}

func doOAuth2(Config map[string]string, ctx *RequestContext) bool {
	return true
}
