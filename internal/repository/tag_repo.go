package repository

import (
	"fmt"

	"github.com/itxryx/tacktack/internal/db"
	"gorm.io/gorm"
)

// GormTagRepository は TagRepository の GORM 実装。
type GormTagRepository struct {
	db *gorm.DB
}

func NewTagRepository(database *gorm.DB) TagRepository {
	return &GormTagRepository{db: database}
}

func (r *GormTagRepository) FindAll() ([]db.Tag, error) {
	var tags []db.Tag
	if err := r.db.Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("FindAll tags: %w", err)
	}
	return tags, nil
}

func (r *GormTagRepository) FindByType(tagType string) ([]db.Tag, error) {
	var tags []db.Tag
	if err := r.db.Where("type = ?", tagType).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("FindByType: %w", err)
	}
	return tags, nil
}

func (r *GormTagRepository) FindOrCreate(name string, tagType string) (*db.Tag, error) {
	// 1. 通常クエリ（deleted_at IS NULL）で検索
	var tag db.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("FindOrCreate tag: %w", err)
	}

	// 2. ソフトデリート済みの同名タグを検索して復活
	var deleted db.Tag
	err = r.db.Unscoped().Where("name = ? AND deleted_at IS NOT NULL", name).First(&deleted).Error
	if err == nil {
		deleted.Type = tagType
		deleted.DeletedAt = gorm.DeletedAt{}
		if err := r.db.Unscoped().Save(&deleted).Error; err != nil {
			return nil, fmt.Errorf("restore soft-deleted tag: %w", err)
		}
		return &deleted, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("FindOrCreate tag (unscoped): %w", err)
	}

	// 3. 新規作成
	newTag := db.Tag{Name: name, Type: tagType}
	if err := r.db.Create(&newTag).Error; err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return &newTag, nil
}

func (r *GormTagRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// task_tags 中間テーブルから紐付けを削除
		if err := tx.Exec("DELETE FROM task_tags WHERE tag_id = ?", id).Error; err != nil {
			return fmt.Errorf("clear task_tags: %w", err)
		}
		// タグ本体を削除
		if err := tx.Delete(&db.Tag{}, id).Error; err != nil {
			return fmt.Errorf("delete tag: %w", err)
		}
		return nil
	})
}
