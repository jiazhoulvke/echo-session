package session

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jiazhoulvke/gocache"
	"github.com/jiazhoulvke/goutil"
	"github.com/jiazhoulvke/seqsvr"
	"github.com/labstack/echo"
)

//Session session
type Session struct {
	SessionID string
	//UserID 用户ID
	UserID int64
	//UserName 用户名
	UserName string
	//UserType 用户类型
	UserType string
	//ExpireAt 过期时间,Unix时间戳格式
	ExpireAt int
	//Data 数据
	Data map[string]interface{}
}

var (
	//CookieKey session key
	CookieKey = "_SESSION_ID"
	//HTTPKey session key
	HTTPKey = "_SESSION_ID"
	//SessionIDPrefix session id 前缀
	SessionIDPrefix = "SESSIONID_"
	//MaxAge 过期时间，默认为一周
	MaxAge = 60 * 60 * 24 * 7

	//storer session存储器
	storer gocache.Storer
	//maker id生成器
	maker = seqsvr.NewMaker("session")

	//ErrStorerNotInit 存储器未初始化
	ErrStorerNotInit = fmt.Errorf("存储器未初始化")
	//ErrSessionNotFound session不存在
	ErrSessionNotFound = fmt.Errorf("session不存在")
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
		ExpireAt:  int(time.Now().Unix()) + MaxAge,
		Data:      make(map[string]interface{}),
	}
	if err := storer.Remember(sess.SessionID, sess, MaxAge); err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		HttpOnly: true,
		MaxAge:   MaxAge,
		Name:     CookieKey,
		Value:    sess.SessionID,
	}
	c.SetCookie(&cookie)
	return sess, nil
}

//FindSession 获取session，如果不存在则直接报错
func FindSession(c echo.Context) (*Session, error) {
	if storer == nil {
		return nil, ErrStorerNotInit
	}
	var sess *Session
	sessionID := ID(c)
	if sessionID == "" {
		return sess, ErrSessionNotFound
	}
	if err := storer.Get(sessionID, &sess); err != nil {
		return sess, err
	}
	return sess, nil
}

//GetSession 获取session，如果不存在则新建
func GetSession(c echo.Context) (*Session, error) {
	if storer == nil {
		return nil, ErrStorerNotInit
	}
	var sess *Session
	var err error
	//先检查session是否存在，如果存在则直接返回
	if sess, err = FindSession(c); err == nil {
		return sess, nil
	}
	return New(c)
}

//Save save session
func (s *Session) Save() error {
	if storer == nil {
		return ErrStorerNotInit
	}
	return storer.Set(s.SessionID, s)
}

//Get 获取session中存储的数据
func (s *Session) Get(key string) (interface{}, bool) {
	v, exists := s.Data[key]
	return v, exists
}

//Float64 获取float64型值
func (s *Session) Float64(key string) (float64, bool) {
	v, exists := s.Get(key)
	if !exists {
		return 0, false
	}
	value, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return value, true
}

//Float32 获取float32型值
func (s *Session) Float32(key string) (float32, bool) {
	v, ok := s.Float64(key)
	return float32(v), ok
}

//Int8 获取int型值
func (s *Session) Int8(key string) (int8, bool) {
	v, ok := s.Float64(key)
	return int8(v), ok
}

//Uint8 获取uint8型值
func (s *Session) Uint8(key string) (uint8, bool) {
	v, ok := s.Float64(key)
	return uint8(v), ok
}

//Int16 获取int16型值
func (s *Session) Int16(key string) (int16, bool) {
	v, ok := s.Float64(key)
	return int16(v), ok
}

//Uint16 获取uint16型值
func (s *Session) Uint16(key string) (uint16, bool) {
	v, ok := s.Float64(key)
	return uint16(v), ok
}

//Int 获取int型值
func (s *Session) Int(key string) (int, bool) {
	v, ok := s.Float64(key)
	return int(v), ok
}

//Uint 获取uint型值
func (s *Session) Uint(key string) (uint, bool) {
	v, ok := s.Float64(key)
	return uint(v), ok
}

//IntSlice 获取[]int型值
func (s *Session) IntSlice(key string) ([]int, bool) {
	v, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]int)
	return value, ok
}

//Int32 获取int32型值
func (s *Session) Int32(key string) (int32, bool) {
	v, ok := s.Float64(key)
	return int32(v), ok
}

//Uint32 获取uint32型值
func (s *Session) Uint32(key string) (uint32, bool) {
	v, ok := s.Float64(key)
	return uint32(v), ok
}

//Int64 获取int64型值
func (s *Session) Int64(key string) (int64, bool) {
	v, ok := s.Float64(key)
	return int64(v), ok
}

//Uint64 获取uint64型值
func (s *Session) Uint64(key string) (uint64, bool) {
	v, ok := s.Float64(key)
	return uint64(v), ok
}

//Int64Slice 获取[]int64型值
func (s *Session) Int64Slice(key string) ([]int64, bool) {
	v, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]int64)
	return value, ok
}

//Byte 获取byte型值
func (s *Session) Byte(key string) (byte, bool) {
	v, ok := s.Float64(key)
	return byte(v), ok
}

//ByteSlice 获取[]byte型值
func (s *Session) ByteSlice(key string) ([]byte, bool) {
	v, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]byte)
	return value, ok
}

//String 获取string型值
func (s *Session) String(key string) (string, bool) {
	v, exists := s.Get(key)
	if !exists {
		return "", false
	}
	value, ok := v.(string)
	return value, ok
}

//StringSlice 获取[]string型值
func (s *Session) StringSlice(key string) ([]string, bool) {
	v, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]string)
	return value, ok
}

//InterfaceSlice 获取[]interface{}型值
func (s *Session) InterfaceSlice(key string) ([]interface{}, bool) {
	v, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]interface{})
	return value, ok
}

//Bool 获取bool型值
func (s *Session) Bool(key string) (value bool, ok bool) {
	v, exists := s.Get(key)
	if !exists {
		return false, false
	}
	value, ok = v.(bool)
	return
}

//Set 设置session数据
func (s *Session) Set(key string, value interface{}) {
	switch v := value.(type) {
	case uint8:
		s.Data[key] = float64(v)
	case int8:
		s.Data[key] = float64(v)
	case uint16:
		s.Data[key] = float64(v)
	case int16:
		s.Data[key] = float64(v)
	case uint32:
		s.Data[key] = float64(v)
	case int32:
		s.Data[key] = float64(v)
	case uint:
		s.Data[key] = float64(v)
	case int:
		s.Data[key] = float64(v)
	case uint64:
		s.Data[key] = float64(v)
	case int64:
		s.Data[key] = float64(v)
	case float32:
		s.Data[key] = float64(v)
	default:
		s.Data[key] = value
	}
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
	cookie, err := c.Cookie(CookieKey)
	if err == nil {
		sessionID = cookie.Value
	} else {
		sessionID = c.FormValue(HTTPKey)
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
	if _, err := FindSession(c); err != nil {
		return false
	}
	return true
}
