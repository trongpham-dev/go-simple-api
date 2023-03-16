package main

import (
	"log"
	"net/http"
	"os"

	"go-simple-api/common"
	"go-simple-api/component"
	"go-simple-api/component/uploadprovider"
	"go-simple-api/middleware"
	"go-simple-api/modules/restaurant/restauranttransport/ginrestaurant"
	"go-simple-api/modules/restaurantlike/transport/ginrestaurantlike"
	"go-simple-api/modules/upload/uploadtransport/ginupload"
	"go-simple-api/modules/user/usertransport/ginuser"
	"go-simple-api/pubsub/pblocal"
	"go-simple-api/subscriber"

	"github.com/gin-gonic/gin"
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

	//subscriber.Setup(appCtx)
	if err := subscriber.NewEngine(appCtx).Start(); err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()

	r.Use(middleware.Recover(appCtx))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// CRUD

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

	return r.Run()
}
