package subscriber

import (
	"context"
	"go-simple-api/component"
)

func Setup(ctx component.AppContext) {
	IncreaseLikeCountAfterUserLikeRestaurant(ctx, context.Background())
}
