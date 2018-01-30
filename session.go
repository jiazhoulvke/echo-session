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
	//ExpireAt 过期时间,Unix时间戳格式
	ExpireAt int
	//Data 数据
	Data map[string]interface{}
}

//Sessioner session接口
type Sessioner interface {
	SetSessionID(sessionID string)
	GetSessionID() string
	SetExpireAt(expireAt int)
	GetExpireAt() int
	SetData(data map[string]interface{})
	GetData() map[string]interface{}
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
}

//Options options
type Options struct {
	HTTPOnly bool
	MaxAge   int
}

var (
	//CookieKey session key
	CookieKey = "_SESSION_ID"
	//HTTPKey session key
	HTTPKey = "_SESSION_ID"
	//SessionIDPrefix session id 前缀
	SessionIDPrefix = "SESSID_"
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
func New(c echo.Context, sess Sessioner, opts ...Options) error {
	if storer == nil {
		return ErrStorerNotInit
	}
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	} else {
		o.HTTPOnly = false
		o.MaxAge = MaxAge
	}
	sess.SetSessionID(fmt.Sprintf("%s%x_%s", SessionIDPrefix, maker.SequenceID(), goutil.RandomString(8)))
	sess.SetExpireAt(int(time.Now().Unix()) + o.MaxAge)
	sess.SetData(map[string]interface{}{})
	if err := storer.Remember(sess.GetSessionID(), sess, MaxAge); err != nil {
		return err
	}
	cookie := http.Cookie{
		HttpOnly: o.HTTPOnly,
		MaxAge:   o.MaxAge,
		Name:     CookieKey,
		Value:    sess.GetSessionID(),
	}
	c.SetCookie(&cookie)
	return nil
}

//FindSession 获取session，如果不存在则直接报错
func FindSession(c echo.Context, sess Sessioner) error {
	if storer == nil {
		return ErrStorerNotInit
	}
	sessionID := ID(c)
	if sessionID == "" {
		return ErrSessionNotFound
	}
	if err := storer.Get(sessionID, &sess); err != nil {
		return err
	}
	return nil
}

//GetSession 获取session，如果不存在则新建
func GetSession(c echo.Context, sess Sessioner, opts ...Options) error {
	if storer == nil {
		return ErrStorerNotInit
	}
	var err error
	//先检查session是否存在，如果存在则直接返回
	if err = FindSession(c, sess); err == nil {
		return nil
	}
	return New(c, sess, opts...)
}

//Save save session
func Save(sess Sessioner) error {
	if storer == nil {
		return ErrStorerNotInit
	}
	return storer.Set(sess.GetSessionID(), sess)
}

//Delete delete session
func Delete(sess Sessioner) error {
	if storer == nil {
		return ErrStorerNotInit
	}
	return storer.Delete(sess.GetSessionID())
}

//Delete 删除session中的值
func (s *Session) Delete(key string) {
	delete(s.Data, key)
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

//Get get value
func (s *Session) Get(key string) (interface{}, bool) {
	v, ok := s.Data[key]
	return v, ok
}

//GetData get data
func (s *Session) GetData() map[string]interface{} {
	return s.Data
}

//SetData set data
func (s *Session) SetData(data map[string]interface{}) {
	s.Data = data
}

//SetExpireAt set ExpireAt
func (s *Session) SetExpireAt(expireAt int) {
	s.ExpireAt = expireAt
}

//GetExpireAt get ExpireAt
func (s *Session) GetExpireAt() int {
	return s.ExpireAt
}

//GetSessionID get session id
func (s *Session) GetSessionID() string {
	return s.SessionID
}

//SetSessionID set session id
func (s *Session) SetSessionID(sessionID string) {
	s.SessionID = sessionID
}

//Set 设置值
func Set(sess Sessioner, key string, value interface{}) {
	sess.Set(key, value)
}

//Get 获取session中存储的数据
func Get(sess Sessioner, key string) (interface{}, bool) {
	v, exists := sess.GetData()[key]
	return v, exists
}

//Float64 获取float64型值
func Float64(sess Sessioner, key string) (float64, bool) {
	v, exists := Get(sess, key)
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
func Float32(sess Sessioner, key string) (float32, bool) {
	v, ok := Float64(sess, key)
	return float32(v), ok
}

//Int8 获取int型值
func Int8(sess Sessioner, key string) (int8, bool) {
	v, ok := Float64(sess, key)
	return int8(v), ok
}

//Uint8 获取uint8型值
func Uint8(sess Sessioner, key string) (uint8, bool) {
	v, ok := Float64(sess, key)
	return uint8(v), ok
}

//Int16 获取int16型值
func Int16(sess Sessioner, key string) (int16, bool) {
	v, ok := Float64(sess, key)
	return int16(v), ok
}

//Uint16 获取uint16型值
func Uint16(sess Sessioner, key string) (uint16, bool) {
	v, ok := Float64(sess, key)
	return uint16(v), ok
}

//Int 获取int型值
func Int(sess Sessioner, key string) (int, bool) {
	v, ok := Float64(sess, key)
	return int(v), ok
}

//Uint 获取uint型值
func Uint(sess Sessioner, key string) (uint, bool) {
	v, ok := Float64(sess, key)
	return uint(v), ok
}

//IntSlice 获取[]int型值
func IntSlice(sess Sessioner, key string) ([]int, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]int)
	return value, ok
}

//Int32 获取int32型值
func Int32(sess Sessioner, key string) (int32, bool) {
	v, ok := Float64(sess, key)
	return int32(v), ok
}

//Uint32 获取uint32型值
func Uint32(sess Sessioner, key string) (uint32, bool) {
	v, ok := Float64(sess, key)
	return uint32(v), ok
}

//Int64 获取int64型值
func Int64(sess Sessioner, key string) (int64, bool) {
	v, ok := Float64(sess, key)
	return int64(v), ok
}

//Uint64 获取uint64型值
func Uint64(sess Sessioner, key string) (uint64, bool) {
	v, ok := Float64(sess, key)
	return uint64(v), ok
}

//Int64Slice 获取[]int64型值
func Int64Slice(sess Sessioner, key string) ([]int64, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]int64)
	return value, ok
}

//Byte 获取byte型值
func Byte(sess Sessioner, key string) (byte, bool) {
	v, ok := Float64(sess, key)
	return byte(v), ok
}

//ByteSlice 获取[]byte型值
func ByteSlice(sess Sessioner, key string) ([]byte, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]byte)
	return value, ok
}

//String 获取string型值
func String(sess Sessioner, key string) (string, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return "", false
	}
	value, ok := v.(string)
	return value, ok
}

//StringSlice 获取[]string型值
func StringSlice(sess Sessioner, key string) ([]string, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]string)
	return value, ok
}

//InterfaceSlice 获取[]interface{}型值
func InterfaceSlice(sess Sessioner, key string) ([]interface{}, bool) {
	v, exists := Get(sess, key)
	if !exists {
		return nil, false
	}
	value, ok := v.([]interface{})
	return value, ok
}

//Bool 获取bool型值
func Bool(sess Sessioner, key string) (value bool, ok bool) {
	v, exists := Get(sess, key)
	if !exists {
		return false, false
	}
	value, ok = v.(bool)
	return
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
