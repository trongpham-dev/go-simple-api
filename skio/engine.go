package skio

import (
	"context"
	"errors"
	"fmt"
	"go-simple-api/component"
	"go-simple-api/component/tokenprovider/jwt"
	"go-simple-api/modules/user/userstorage"
	"go-simple-api/modules/user/usertransport/skuser"
	"sync"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

type RealtimeEngine interface {
	UserSockets(userId int) []AppSocket
	EmitToRoom(room string, key string, data interface{}) error
	EmitToUser(userId int, key string, data interface{}) error
	Run(ctx component.AppContext, engine *gin.Engine) error
	//Emit(userId int) error
}

type rtEngine struct {
	server  *socketio.Server
	storage map[int][]AppSocket // key: user id ; value: an array of web socket(because 1 user can connect via multiple medias type)
	locker  *sync.RWMutex
}

func NewEngine() *rtEngine {
	return &rtEngine{
		storage: make(map[int][]AppSocket),
		locker:  new(sync.RWMutex),
	}
}

func (engine *rtEngine) saveAppSocket(userId int, appSck AppSocket) {
	engine.locker.Lock()

	//appSck.Join("order-{ordID}")

	if v, ok := engine.storage[userId]; ok {
		engine.storage[userId] = append(v, appSck)
	} else {
		engine.storage[userId] = []AppSocket{appSck}
	}

	engine.locker.Unlock()
}

func (engine *rtEngine) getAppSocket(userId int) []AppSocket {
	engine.locker.RLock()
	defer engine.locker.RUnlock()

	return engine.storage[userId]
}

func (engine *rtEngine) removeAppSocket(userId int, appSck AppSocket) {
	engine.locker.Lock()
	defer engine.locker.Unlock()

	if v, ok := engine.storage[userId]; ok {
		for i := range v {
			if v[i] == appSck {
				engine.storage[userId] = append(v[:i], v[i+1:]...)
				break
			}
		}
	}
}

func (engine *rtEngine) UserSockets(userId int) []AppSocket {
	var sockets []AppSocket

	if scks, ok := engine.storage[userId]; ok {
		return scks
	}

	return sockets
}

func (engine *rtEngine) EmitToRoom(room string, key string, data interface{}) error {
	engine.server.BroadcastToRoom("/", room, key, data)
	return nil
}

func (engine *rtEngine) EmitToUser(userId int, key string, data interface{}) error {
	sockets := engine.getAppSocket(userId)

	for _, s := range sockets {
		s.Emit(key, data)
	}

	return nil
}

// open server
func (engine *rtEngine) Run(appCtx component.AppContext, r *gin.Engine) error {
	server, err := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{websocket.Default},
	})

	if err != nil {
		return err
	}

	engine.server = server

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID(), " IP:", s.RemoteAddr(), s.ID())
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	// Setup

	server.OnEvent("/", "authenticate", func(s socketio.Conn, token string) {
		db := appCtx.GetMainDBConnection()
		store := userstorage.NewSQLStore(db)

		tokenProvider := jwt.NewTokenJWTProvider(appCtx.SecretKey())

		payload, err := tokenProvider.Validate(token)
		if err != nil {
			s.Emit("authentication_failed", err.Error())
			s.Close()
			return
		}

		user, err := store.FindUser(context.Background(), map[string]interface{}{"id": payload.UserId})

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

		appSck := NewAppSocket(s, user)
		engine.saveAppSocket(user.Id, appSck)

		s.Emit("authenticated", user)

		//appSck.Join(user.GetRole()) // the same
		//if user.GetRole() == "admin" {
		//	appSck.Join("admin")
		//}

		server.OnEvent("/", "UserUpdateLocation", skuser.OnUserUpdateLocation(appCtx, user))
	})

	go server.Serve()

	r.GET("/socket.io/*any", gin.WrapH(server))
	r.POST("/socket.io/*any", gin.WrapH(server))

	return nil
}
