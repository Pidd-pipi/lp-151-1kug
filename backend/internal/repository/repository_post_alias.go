package repository

import (
	"errors"
	"fmt"

	"github.com/gbtreehole/backend/internal/model"
	"gorm.io/gorm"
)

type PostAliasRepository interface {
	Create(alias *model.PostAlias) error
	FindByPostAndIdentity(postID, identityID uint) (*model.PostAlias, error)
	ListByPostIDs(postIDs []uint) ([]model.PostAlias, error)
}

type postAliasRepository struct {
	db *gorm.DB
}

func NewPostAliasRepository(db *gorm.DB) PostAliasRepository {
	return &postAliasRepository{db: db}
}

func (r *postAliasRepository) Create(alias *model.PostAlias) error {
	if err := r.db.Create(alias).Error; err != nil {
		return fmt.Errorf("create post alias: %w", err)
	}
	return nil
}

func (r *postAliasRepository) FindByPostAndIdentity(postID, identityID uint) (*model.PostAlias, error) {
	var alias model.PostAlias
	if err := r.db.Where("post_id = ? AND identity_id = ?", postID, identityID).First(&alias).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find post alias: %w", err)
	}
	return &alias, nil
}

func (r *postAliasRepository) ListByPostIDs(postIDs []uint) ([]model.PostAlias, error) {
	aliases := make([]model.PostAlias, 0)
	if len(postIDs) == 0 {
		return aliases, nil
	}
	if err := r.db.Where("post_id IN ?", postIDs).Find(&aliases).Error; err != nil {
		return nil, fmt.Errorf("list post aliases: %w", err)
	}
	return aliases, nil
}
