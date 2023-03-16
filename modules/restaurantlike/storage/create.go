package restaurantlikestorage

import (
	"context"
	"go-simple-api/common"
	restaurantlikemodel "go-simple-api/modules/restaurantlike/model"
)

func (s *sqlStore) Create(ctx context.Context, data *restaurantlikemodel.Like) error {
	db := s.db

	if err := db.Create(data).Error; err != nil {
		return common.ErrDB(err)
	}

	return nil
}
