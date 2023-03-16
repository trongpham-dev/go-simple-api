package ginrestaurantlike

import (
	"go-simple-api/common"
	"go-simple-api/component"
	rstlikebiz "go-simple-api/modules/restaurantlike/biz"
	restaurantlikestorage "go-simple-api/modules/restaurantlike/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DELETE /v1/restaurants/:id/like

func UserUnlikeRestaurant(appCtx component.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := common.FromBase58(c.Param("id"))

		if err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		requester := c.MustGet(common.CurrentUser).(common.Requester)

		//data := restaurantlikemodel.Like{
		//	RestaurantId: int(uid.GetLocalID()),
		//	UserId:       requester.GetUserId(),
		//}

		store := restaurantlikestorage.NewSQLStore(appCtx.GetMainDBConnection())
		// incStore := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
		biz := rstlikebiz.NewUserUnlikeRestaurantBiz(store, appCtx.GetPubsub())

		if err := biz.UnlikeRestaurant(c.Request.Context(), requester.GetUserId(), int(uid.GetLocalID())); err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, common.SimpleSuccessResponse(true))
	}
}
