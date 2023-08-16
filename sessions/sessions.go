package sessions

import (
	"github.com/gorilla/sessions"
)

// cookie store
var Store = sessions.NewCookieStore([]byte("t0p-s3cr3t"))
