package subscriber

import (
	"context"
	"go-simple-api/common"
	"go-simple-api/component"
	"go-simple-api/modules/restaurant/restaurantstorage"
	"go-simple-api/pubsub"
)

type HasRestaurantId interface {
	GetRestaurantId() int
}

func IncreaseLikeCountAfterUserLikeRestaurant(appCtx component.AppContext, ctx context.Context) {
	c, _ := appCtx.GetPubsub().Subscribe(ctx, common.TopicUserLikeRestaurant)

	store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())

	go func() {
		defer common.AppRecover()
		for {
			msg := <-c
			likeData := msg.Data().(HasRestaurantId)
			_ = store.IncreaseLikeCount(ctx, likeData.GetRestaurantId())
		}
	}()
}

// I wish I could do something like that
//func RunIncreaseLikeCountAfterUserLikeRestaurant(appCtx component.AppContext) func(ctx context.Context, message *pubsub.Message) error {
//	store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
//
//	return func(ctx context.Context, message *pubsub.Message) error {
//		likeData := message.Data().(HasRestaurantId)
//		return store.IncreaseLikeCount(ctx, likeData.GetRestaurantId())
//	}
//}

func RunIncreaseLikeCountAfterUserLikeRestaurant(appCtx component.AppContext) consumerJob {
	return consumerJob{
		Title: "Increase like count after user likes restaurant",
		Hld: func(ctx context.Context, message *pubsub.Message) error {
			store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
			likeData := message.Data().(HasRestaurantId)
			return store.IncreaseLikeCount(ctx, likeData.GetRestaurantId())
		},
	}
}

func RunDecreaseLikeCountAfterUserLikeRestaurant(appCtx component.AppContext) consumerJob {
	return consumerJob{
		Title: "Decrease like count after user likes restaurant",
		Hld: func(ctx context.Context, message *pubsub.Message) error {
			store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
			likeData := message.Data().(HasRestaurantId)
			return store.DecreaseLikeCount(ctx, likeData.GetRestaurantId())
		},
	}
}
