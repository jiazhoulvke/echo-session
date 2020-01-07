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

type mySession struct {
	Session
	UserID int
}

func (h *handler) CreateSession(c echo.Context) error {
	var sess mySession
	if err := FindSession(c, &sess); err == nil {
		return c.String(200, "ERROR: error must not nil")
	}
	if err := GetSession(c, &sess); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	sess.Data["answer"] = int(42)
	sess.UserID = 9527
	if err := Save(&sess); err != nil {
		return err
	}
	return c.String(200, sess.SessionID)
}

func (h *handler) ChangeSession(c echo.Context) error {
	var sess mySession
	if err := FindSession(c, &sess); err != nil {
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
	if sess.UserID != 9527 {
		return c.String(200, fmt.Sprintf("ERROR:UserID not 9527, is %v", sess.UserID))
	}
	sess.Data["answer"] = 21
	if err := Save(&sess); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if err := FindSession(c, &sess); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if sess.Data["answer"].(float64) != 21 {
		return c.String(200, "ERROR:answer not 21")
	}
	return c.String(200, "OK")
}

func (h *handler) DeleteSession(c echo.Context) error {
	var sess mySession
	if err := FindSession(c, &sess); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if err := Delete(&sess); err != nil {
		return c.String(200, "ERROR:"+err.Error())
	}
	if err := FindSession(c, &sess); err == nil {
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
		values.Set(FormKey, mySessionID)
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
		values.Set(FormKey, mySessionID)
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
		var sess mySession
		err := New(c, &sess)
		So(err, ShouldBeNil)
		So(sess.SessionID, ShouldNotEqual, "")
		Convey("测试读取不存在的值", func() {
			for k := range testData {
				_, ok := Get(&sess, k)
				So(ok, ShouldBeFalse)
			}
		})
		//将值放入session
		for k, v := range testData {
			sess.Set(k, v)
		}
		So(Save(&sess), ShouldBeNil)
		Convey("测试读取存在的值", func() {
			for k, v := range testData {
				switch k {
				case "uint8":
					value, ok := Uint8(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int8":
					value, ok := Int8(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint16":
					value, ok := Uint16(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int16":
					value, ok := Int16(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint32":
					value, ok := Uint32(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int32":
					value, ok := Int32(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint":
					value, ok := Uint(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int":
					value, ok := Int(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "uint64":
					value, ok := Uint64(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "int64":
					value, ok := Int64(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "float32":
					value, ok := Float32(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "float64":
					value, ok := Float64(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "byte":
					value, ok := Byte(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "string":
					value, ok := String(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "bool":
					value, ok := Bool(&sess, k)
					So(ok, ShouldBeTrue)
					So(value, ShouldEqual, v)
				case "byteslice":
					value, ok := ByteSlice(&sess, k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "intslice":
					value, ok := IntSlice(&sess, k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "int64slice":
					value, ok := Int64Slice(&sess, k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "stringslice":
					value, ok := StringSlice(&sess, k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				case "interfaceslice":
					value, ok := InterfaceSlice(&sess, k)
					So(ok, ShouldBeTrue)
					So(reflect.DeepEqual(v, value), ShouldBeTrue)
				}
			}
		})
	})
}
