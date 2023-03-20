package subscriber

import (
	"context"
	"go-simple-api/component"
	"go-simple-api/modules/restaurant/restaurantstorage"
	"go-simple-api/pubsub"
	"go-simple-api/skio"
)

func RunDecreaseLikeCountAfterUserUnlikeRestaurant(appCtx component.AppContext, rtEngine skio.RealtimeEngine) consumerJob {
	return consumerJob{
		Title: "Decrease like count after user unlikes restaurant",
		Hld: func(ctx context.Context, message *pubsub.Message) error {
			store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
			likeData := message.Data().(HasRestaurantId)
			return store.DecreaseLikeCount(ctx, likeData.GetRestaurantId())
		},
	}
}
