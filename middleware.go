package kocha

import (
	"strconv"
)

// Middleware is the interface that middleware.
type Middleware interface {
	Before(c *Controller)
	After(c *Controller)
}

var (
	// Default middlewares.
	DefaultMiddlewares = []Middleware{
		&ResponseContentTypeMiddleware{},
	}
)

// Middleware that set Content-Type header.
type ResponseContentTypeMiddleware struct{}

func (m *ResponseContentTypeMiddleware) Before(c *Controller) {
	// do nothing
}

func (m *ResponseContentTypeMiddleware) After(c *Controller) {
	res := c.Response
	res.Header().Set("Content-Type", res.ContentType)
}

// Session processing middleware.
type SessionMiddleware struct{}

func (m *SessionMiddleware) Before(c *Controller) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case ErrSession:
				Log.Error("%v", err)
			case ErrSessionExpected:
				Log.Info("%v", err)
			default:
				panic(err)
			}
			c.Session = make(Session)
		}
	}()
	cookie, err := c.Request.Cookie(appConfig.Session.Name)
	if err != nil {
		panic(NewErrSessionExpected("new session"))
	}
	sess := appConfig.Session.Store.Load(cookie.Value)
	expiresStr, ok := sess[SessionExpiresKey]
	if !ok {
		panic(NewErrSession("expires value not found"))
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		panic(NewErrSession(err.Error()))
	}
	if expires < Now().Unix() {
		panic(NewErrSessionExpected("session has been expired"))
	}
	c.Session = sess
}

func (m *SessionMiddleware) After(c *Controller) {
	expires, _ := expiresFromDuration(appConfig.Session.SessionExpires)
	c.Session[SessionExpiresKey] = strconv.FormatInt(expires.Unix(), 10)
	cookie := newSessionCookie(c)
	cookie.Value = appConfig.Session.Store.Save(c.Session)
	c.Response.SetCookie(cookie)
}

// Request logging middleware.
type RequestLoggingMiddleware struct{}

func (m *RequestLoggingMiddleware) Before(c *Controller) {
	// do nothing.
}

func (m *RequestLoggingMiddleware) After(c *Controller) {
	Log.Info(`"%v %v %v" %v`, c.Request.Method, c.Request.RequestURI, c.Request.Proto, c.Response.StatusCode)
}
