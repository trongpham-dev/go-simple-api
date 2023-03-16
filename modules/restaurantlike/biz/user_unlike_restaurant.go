package rstlikebiz

import (
	"context"
	"go-simple-api/common"
	restaurantlikemodel "go-simple-api/modules/restaurantlike/model"
	"go-simple-api/pubsub"
)

type UserUnlikeRestaurantStore interface {
	Delete(ctx context.Context, userId, restaurantId int) error
}

// type DecreaseLikeCountStore interface {
// 	DecreaseLikeCount(ctx context.Context, id int) error
// }

type userUnlikeRestaurantBiz struct {
	store UserUnlikeRestaurantStore
	// decStore DecreaseLikeCountStore
	pubsub pubsub.Pubsub
}

func NewUserUnlikeRestaurantBiz(store UserUnlikeRestaurantStore, pubsub pubsub.Pubsub) *userUnlikeRestaurantBiz {
	return &userUnlikeRestaurantBiz{store: store, pubsub: pubsub}
}

func (biz *userUnlikeRestaurantBiz) UnlikeRestaurant(
	ctx context.Context,
	userId,
	restaurantId int,
) error {
	err := biz.store.Delete(ctx, userId, restaurantId)

	if err != nil {
		return restaurantlikemodel.ErrCannotUnlikeRestaurant(err)
	}

	// side effect
	// go func() {
	// 	defer common.AppRecover()
	// 	job := asyncjob.NewJob(func(ctx context.Context) error {
	// 		return biz.decStore.DecreaseLikeCount(ctx, restaurantId)
	// 	})

	// 	//job.SetRetryDurations([]time.Duration{time.Second * 3})

	// 	_ = asyncjob.NewGroup(true, job).Run(ctx)
	// }()

	// _ = biz.decStore.DecreaseLikeCount(ctx, restaurantId)

	biz.pubsub.Publish(ctx, common.TopicUserDislikeRestaurant, pubsub.NewMessage(&restaurantlikemodel.Like{
		RestaurantId: restaurantId,
		UserId:       userId,
	}))

	return nil
}
