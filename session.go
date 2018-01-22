package session

import (
	"fmt"
	"net/http"

	"github.com/jiazhoulvke/gocache"
	"github.com/jiazhoulvke/goutil"
	"github.com/jiazhoulvke/seqsvr"
	"github.com/labstack/echo"
)

//Session session
type Session struct {
	SessionID string
	UserID    int64
	UserName  string
	UserType  string
	Data      map[string]interface{}
}

var (
	//CookieSessionKey session key
	CookieSessionKey = "_SESSION_ID"
	//HTTPSessionKey session key
	HTTPSessionKey = "_SESSION_ID"
	//SessionIDPrefix session id 前缀
	SessionIDPrefix = "SESSIONID_"
	//CookieMaxAge cookie过期时间
	CookieMaxAge = 60 * 60 * 24 * 7 //一周

	//storer session存储器
	storer gocache.Storer
	//maker id生成器
	maker = seqsvr.NewMaker("session")

	//ErrStorerNotInit 存储器未初始化
	ErrStorerNotInit = fmt.Errorf("存储器未初始化")
)

//Init 初始化session存储器
func Init(s gocache.Storer) {
	storer = s
}

//New 新建session
func New(c echo.Context) (*Session, error) {
	if storer == nil {
		return nil, ErrStorerNotInit
	}
	sess := &Session{
		SessionID: fmt.Sprintf("%s%x_%s", SessionIDPrefix, maker.SequenceID(), goutil.RandomString(8)),
		Data:      make(map[string]interface{}),
	}
	if err := storer.Set(sess.SessionID, sess); err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		HttpOnly: true,
		MaxAge:   CookieMaxAge,
		Name:     CookieSessionKey,
		Value:    sess.SessionID,
	}
	c.SetCookie(&cookie)
	return sess, nil
}

//GetSession 获取session
func GetSession(c echo.Context) (*Session, error) {
	if storer == nil {
		return nil, ErrStorerNotInit
	}
	var sess *Session
	var err error
	sessionID := ID(c)
	if sessionID == "" { //如果session不存在则创建
		sess, err = New(c)
		if err != nil {
			return sess, nil
		}
	}
	if err := storer.Get(sessionID, &sess); err != nil {
		sess, err = New(c)
		if err != nil {
			return sess, nil
		}
	}
	return sess, nil
}

//Save save session
func (s *Session) Save() error {
	if storer == nil {
		return ErrStorerNotInit
	}
	return storer.Set(s.SessionID, s)
}

//Get 获取session中存储的数据
func (s *Session) Get(key string) interface{} {
	return s.Data[key]
}

//Set 设置session数据
func (s *Session) Set(key string, value interface{}) {
	s.Data[key] = value
}

//Delete delete session
func (s *Session) Delete() error {
	if storer == nil {
		return ErrStorerNotInit
	}
	return storer.Delete(s.SessionID)
}

//ID 获取session id
func ID(c echo.Context) string {
	var sessionID string
	cookie, err := c.Cookie(CookieSessionKey)
	if err == nil {
		sessionID = cookie.Value
	} else {
		sessionID = c.FormValue(HTTPSessionKey)
	}
	return sessionID
}

//IsLogOn 是否已经登录
func IsLogOn(c echo.Context) bool {
	if storer == nil {
		return false
	}
	sessionID := ID(c)
	if sessionID == "" {
		return false
	}
	if _, err := GetSession(c); err != nil {
		return false
	}
	return true
}
