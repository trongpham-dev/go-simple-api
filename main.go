package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-simple-api/common"
	"go-simple-api/component"
	"go-simple-api/component/tokenprovider/jwt"
	"go-simple-api/component/uploadprovider"
	"go-simple-api/middleware"
	"go-simple-api/modules/restaurant/restauranttransport/ginrestaurant"
	"go-simple-api/modules/restaurantlike/transport/ginrestaurantlike"
	"go-simple-api/modules/upload/uploadtransport/ginupload"
	"go-simple-api/modules/user/userstorage"
	"go-simple-api/modules/user/usertransport/ginuser"
	"go-simple-api/pubsub/pblocal"
	"go-simple-api/skio"
	"go-simple-api/subscriber"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("")
	log.Print(dsn)
	db, err := gorm.Open(mysql.Open(""), &gorm.Config{})

	s3BucketName := ""
	s3Region := ""
	s3APIKey := ""
	s3SecretKey := ""
	s3Domain := ""
	secretKey := ""

	s3Provider := uploadprovider.NewS3Provider(s3BucketName, s3Region, s3APIKey, s3SecretKey, s3Domain)

	if err != nil {
		log.Fatalln(err)
	}

	if err := runService(db, s3Provider, secretKey); err != nil {
		log.Fatalln(err)
	}
}

func runService(db *gorm.DB, upProvider uploadprovider.UploadProvider, secretKey string) error {
	appCtx := component.NewAppContext(db, upProvider, secretKey, pblocal.NewPubSub())

	r := gin.Default()

	rtEngine := skio.NewEngine()

	//subscriber.Setup(appCtx)
	if err := subscriber.NewEngine(appCtx, rtEngine).Start(); err != nil {
		log.Fatalln(err)
	}

	if err := rtEngine.Run(appCtx, r); err != nil {
		log.Fatal(err)
	}

	r.Use(middleware.Recover(appCtx))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// CRUD
	r.StaticFile("/demo/", "./demo.html")
	v1 := r.Group("/v1")

	v1.POST("/upload", ginupload.Upload(appCtx))

	v1.POST("/register", ginuser.Register(appCtx))
	v1.POST("/login", ginuser.Login(appCtx))
	v1.GET("/profile", middleware.RequiredAuth(appCtx), ginuser.GetProfile(appCtx))

	restaurants := v1.Group("/restaurants", middleware.RequiredAuth(appCtx))
	{
		restaurants.POST("", ginrestaurant.CreateRestaurant(appCtx))
		restaurants.GET("/:id", ginrestaurant.GetRestaurant(appCtx))
		restaurants.GET("", ginrestaurant.ListRestaurant(appCtx))
		restaurants.PATCH("/:id", ginrestaurant.UpdateRestaurant(appCtx))
		restaurants.DELETE("/:id", ginrestaurant.DeleteRestaurant(appCtx))

		restaurants.GET("/:id/liked-users", ginrestaurantlike.ListUser(appCtx))
		restaurants.POST("/:id/like", ginrestaurantlike.UserLikeRestaurant(appCtx))
		restaurants.DELETE("/:id/unlike", ginrestaurantlike.UserUnlikeRestaurant(appCtx))
	}

	v1.GET("/encode-uid", func(c *gin.Context) {
		type reqData struct {
			DbType int `form:"type"`
			RealId int `form:"id"`
		}

		var d reqData
		c.ShouldBind(&d)

		c.JSON(http.StatusOK, gin.H{
			"id": common.NewUID(uint32(d.RealId), d.DbType, 1),
		})
	})

	// startSocketIOServer(r, appCtx)

	return r.Run()
}

func startSocketIOServer(engine *gin.Engine, appCtx component.AppContext) {
	server, _ := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{websocket.Default},
	})

	server.OnConnect("/", func(s socketio.Conn) error {
		//s.SetContext("")
		fmt.Println("connected:", s.ID(), " IP:", s.RemoteAddr())

		//s.Join("Shipper")
		//server.BroadcastToRoom("/", "Shipper", "test", "Hello 200lab")

		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
		// Remove socket from socket engine (from app context)
	})

	server.OnEvent("/", "authenticate", func(s socketio.Conn, token string) {

		// Validate token
		// If false: s.Close(), and return

		// If true
		// => UserId
		// Fetch db find user by Id
		// Here: s belongs to who? (user_id)
		// We need a map[user_id][]socketio.Conn

		db := appCtx.GetMainDBConnection()
		store := userstorage.NewSQLStore(db)
		//
		tokenProvider := jwt.NewTokenJWTProvider(appCtx.SecretKey())
		//
		payload, err := tokenProvider.Validate(token)

		if err != nil {
			s.Emit("authentication_failed", err.Error())
			s.Close()
			return
		}
		//
		user, err := store.FindUser(context.Background(), map[string]interface{}{"id": payload.UserId})
		//
		if err != nil {
			s.Emit("authentication_failed", err.Error())
			s.Close()
			return
		}

		if user.Status == 0 {
			s.Emit("authentication_failed", errors.New("you has been banned/deleted"))
			s.Close()
			return
		}

		user.Mask(false)

		s.Emit("your_profile", user)
	})

	server.OnEvent("/", "test", func(s socketio.Conn, msg string) {
		log.Println(msg)
	})

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	server.OnEvent("/", "notice", func(s socketio.Conn, p Person) {
		fmt.Println("server receive notice:", p.Name, p.Age)

		p.Age = 33
		s.Emit("notice", p)

	})

	server.OnEvent("/", "test", func(s socketio.Conn, msg string) {
		fmt.Println("server receive test:", msg)
	})
	//
	//server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
	//	s.SetContext(msg)
	//	return "recv " + msg
	//})
	//
	//server.OnEvent("/", "bye", func(s socketio.Conn) string {
	//	last := s.Context().(string)
	//	s.Emit("bye", last)
	//	s.Close()
	//	return last
	//})
	//
	//server.OnEvent("/", "noteSumit", func(s socketio.Conn) string {
	//	last := s.Context().(string)
	//	s.Emit("bye", last)
	//	s.Close()
	//	return last
	//})

	go server.Serve()

	// bat buoc
	engine.GET("/socket.io/*any", gin.WrapH(server))
	engine.POST("/socket.io/*any", gin.WrapH(server))
}
