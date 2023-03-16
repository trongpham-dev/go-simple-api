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

// GET /v1/restaurants/:id/liked-users

func ListUser(appCtx component.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := common.FromBase58(c.Param("id"))
		//var filter restaurantlikemodel.Filter
		//
		if err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		filter := restaurantlikemodel.Filter{
			RestaurantId: int(uid.GetLocalID()),
		}

		var paging common.Paging

		if err := c.ShouldBind(&paging); err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		paging.Fulfill()

		store := restaurantlikestorage.NewSQLStore(appCtx.GetMainDBConnection())
		biz := rstlikebiz.NewListUserLikeRestaurantBiz(store)

		result, err := biz.ListUsers(c.Request.Context(), &filter, &paging)

		if err != nil {
			panic(err)
		}

		for i := range result {
			result[i].Mask(false)
		}

		c.JSON(http.StatusOK, common.NewSuccessResponse(result, paging, filter))
	}
}
