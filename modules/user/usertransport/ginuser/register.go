package ginuser

import (
	"go-simple-api/common"
	"go-simple-api/component"
	"go-simple-api/component/hasher"
	"go-simple-api/modules/user/userbiz"
	"go-simple-api/modules/user/usermodel"
	"go-simple-api/modules/user/userstorage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(appCtx component.AppContext) func(*gin.Context) {
	return func(c *gin.Context) {
		db := appCtx.GetMainDBConnection()
		var data usermodel.UserCreate

		if err := c.ShouldBind(&data); err != nil {
			panic(err)
		}

		store := userstorage.NewSQLStore(db)
		md5 := hasher.NewMd5Hash()
		biz := userbiz.NewRegisterBusiness(store, md5)

		if err := biz.Register(c.Request.Context(), &data); err != nil {
			panic(err)
		}

		data.Mask(false)

		c.JSON(http.StatusOK, common.SimpleSuccessResponse(data.FakeId.String()))
	}
}
