# echo-session #

基于[echo](https://github.com/labstack/echo)和[gocache](https://github.com/jiazhoulvke/gocache)的session库。

```go
package main

import (
	"log"

	session "github.com/jiazhoulvke/echo-session"
	"github.com/jiazhoulvke/gocache"
	"github.com/jiazhoulvke/gocache/drivers/redis"
	"github.com/labstack/echo"
)

func main() {
	//初始化缓存库
	if err := gocache.Open(redis.Options{
		Host:        "127.0.0.1",
		Port:        6379,
		IdleTimeout: 60,
	}); err != nil {
		panic(err)
	}
	//初始化session库
	session.Init(gocache.Store("SESSION"))

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		sess, err := session.GetSession(c)
		if err != nil {
			return c.String(200, err.Error())
		}
		sess.Set("foo", "bar")
		if err := sess.Save(); err != nil {
			return c.String(200, err.Error())
		}
		return c.String(200, sess.SessionID)
	})
	e.GET("/check", func(c echo.Context) error {
		sess, err := session.GetSession(c)
		if err != nil {
			return c.String(200, err.Error())
		}
		hello, ok := sess.Data["hello"]
		if ok {
			return c.String(200, hello.(string))
		}
		return c.String(200, "not found")
	})
	e.DELETE("/", func(c echo.Context) error {
		sess, err := session.GetSession(c)
		if err != nil {
			return c.String(200, err.Error())
		}
		if err := sess.Delete(); err != nil {
			return c.String(200, err.Error())
		}
		return c.String(200, "DELETED")
	})
	log.Fatal(e.Start(":8877"))
}
```
