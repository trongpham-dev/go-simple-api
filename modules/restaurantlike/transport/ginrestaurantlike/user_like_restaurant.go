package ginrestaurantlike

import (
	"go-simple-api/common"
	"go-simple-api/component"
	rstlikebiz "go-simple-api/modules/restaurantlike/biz"
	restaurantlikemodel "go-simple-api/modules/restaurantlike/model"
	restaurantlikestorage "go-simple-api/modules/restaurantlike/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /v1/restaurants/:id/like

func UserLikeRestaurant(appCtx component.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := common.FromBase58(c.Param("id"))

		if err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		requester := c.MustGet(common.CurrentUser).(common.Requester)

		data := restaurantlikemodel.Like{
			RestaurantId: int(uid.GetLocalID()),
			UserId:       requester.GetUserId(),
		}

		store := restaurantlikestorage.NewSQLStore(appCtx.GetMainDBConnection())
		//incStore := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
		biz := rstlikebiz.NewUserLikeRestaurantBiz(store, appCtx.GetPubsub())

		if err := biz.LikeRestaurant(c.Request.Context(), &data); err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, common.SimpleSuccessResponse(true))
	}
}
