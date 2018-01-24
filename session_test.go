package session

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/jiazhoulvke/gocache"
	"github.com/jiazhoulvke/gocache/drivers/redis"
	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	gocache.Open(redis.Options{
		Host:        "127.0.0.1",
		Port:        6379,
		IdleTimeout: 60,
	})
	Init(gocache.Store("SESSION"))
}

var (
	mySessionID string

	testData = map[string]interface{}{
		"uint8":          8,
		"int8":           8,
		"uint16":         16,
		"int16":          16,
		"uint32":         32,
		"int32":          32,
		"uint":           64,
		"int":            64,
		"uint64":         64,
		"int64":          64,
		"float32":        32,
		"float64":        64,
		"bool":           true,
		"byte":           8,
		"string":         "hello",
		"stringslice":    []string{"hello", "world"},
		"byteslice":      []byte{8},
		"intslice":       []int{32},
		"int64slice":     []int64{64},
		"interfaceslice": []interface{}{"foo", "bar", 123},
	}
)

type handler struct {
}

func (h *handler) CreateSession(c echo.Context) error {
	var sess *Session
	var err error
	_, err = FindSession(c)
	if err == nil {
		return c.String(200, "ERROR: err must not nil")
	}
	sess, err = GetSession(c)
	if err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}

	sess.Data["answer"] = int(42)
	defer sess.Save()
	return c.String(200, sess.SessionID)
}

func (h *handler) ChangeSession(c echo.Context) error {
	if !IsLogOn(c) {
		return c.String(200, "ERROR:state error,not login")
	}
	sess, err := FindSession(c)
	if err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	answer, exists := sess.Data["answer"]
	if !exists {
		return c.String(200, "ERROR:answer not found")
	}
	var num float64
	var ok bool
	//int类型序列化成json后，再反序列化时默认会转为float64,所以这里取出的时候要用float64
	if num, ok = answer.(float64); !ok {
		fmt.Println("[answer]:", answer)
		return c.String(200, "ERROR:answer is not number")
	}
	if num != 42 {
		return c.String(200, fmt.Sprintf("ERROR:answer not 42, is %v", num))
	}
	sess.Data["answer"] = 21
	sess.Save()
	sess, err = FindSession(c)
	if err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if sess.Data["answer"].(float64) != 21 {
		return c.String(200, "ERROR:answer not 21")
	}
	return c.String(200, "OK")
}

func (h *handler) DeleteSession(c echo.Context) error {
	sess, err := FindSession(c)
	if err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if err := sess.Delete(); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	_, err = FindSession(c)
	if err == nil {
		return c.String(200, "ERROR:err should not be nil")
	}
	return c.String(200, "OK")
}

func TestCreateSession(t *testing.T) {
	Convey("测试创建Session", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &handler{}
		err := h.CreateSession(c)
		So(err, ShouldBeNil)
		body := rec.Body.String()
		t.Log("[TestCreateSession] body:", body)
		So(body[0:5], ShouldNotEqual, "ERROR")
		mySessionID = body //存储SessionID，供后续使用
	})
}

func TestChangeSession(t *testing.T) {
	Convey("测试修改Session", t, func() {
		e := echo.New()
		values := make(url.Values)
		values.Set(HTTPKey, mySessionID)
		req := httptest.NewRequest(echo.POST, "/", strings.NewReader(values.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &handler{}
		err := h.ChangeSession(c)
		So(err, ShouldBeNil)
		body := rec.Body.String()
		t.Log("[TestChangeSession] body:", body)
		So(body, ShouldEqual, "OK")
	})
}

func TestDeleteSession(t *testing.T) {
	Convey("测试删除Session", t, func() {
		e := echo.New()
		values := make(url.Values)
		values.Set(HTTPKey, mySessionID)
		req := httptest.NewRequest(echo.POST, "/", strings.NewReader(values.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &handler{}
		err := h.DeleteSession(c)
		So(err, ShouldBeNil)
		body := rec.Body.String()
		t.Log("[TestDeleteSession] body:", body)
		So(body, ShouldEqual, "OK")
	})
}

//TestValue 测试对值的读取
func TestValue(t *testing.T) {
	Convey("测试对值的读取", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		sess, err := New(c)
		So(err, ShouldBeNil)
		So(sess.SessionID, ShouldNotEqual, "")
		Convey("测试读取不存在的值", func() {
			for k := range testData {
				_, ok := sess.Get(k)
				So(ok, ShouldBeFalse)
			}
		})
		//将值放入session
		for k, v := range testData {
			sess.Set(k, v)
		}
		So(sess.Save(), ShouldBeNil)
		Convey("测试读取存在的值", func() {
			for k, v := range testData {
				switch k {
				case "uint8":
					value, ok := sess.Uint8(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int8":
					value, ok := sess.Int8(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint16":
					value, ok := sess.Uint16(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int16":
					value, ok := sess.Int16(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint32":
					value, ok := sess.Uint32(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int32":
					value, ok := sess.Int32(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint":
					value, ok := sess.Uint(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int":
					value, ok := sess.Int(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint64":
					value, ok := sess.Uint64(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int64":
					value, ok := sess.Int64(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "float32":
					value, ok := sess.Float32(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "float64":
					value, ok := sess.Float64(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "byte":
					value, ok := sess.Byte(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "string":
					value, ok := sess.String(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "bool":
					value, ok := sess.Bool(k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "byteslice":
					value, ok := sess.ByteSlice(k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "intslice":
					value, ok := sess.IntSlice(k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "int64slice":
					value, ok := sess.Int64Slice(k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "stringslice":
					value, ok := sess.StringSlice(k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "interfaceslice":
					value, ok := sess.InterfaceSlice(k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				}
			}
		})
	})
}
