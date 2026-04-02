package repository

import (
	"context"
	"strings"
	"time"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/pkg/cursor"
	"gorm.io/gorm"
)

// sellerPagingMode mirrors productListPagingMode for the seller product list.
type sellerPagingMode int

const (
	sellerPagingFirstPage sellerPagingMode = iota
	sellerPagingForward
	sellerPagingBackward
)

func resolveSellerListPaging(beforeCursor, afterCursor, fp string) (sellerPagingMode, cursor.SellerProductCursor) {
	var anchor cursor.SellerProductCursor
	if strings.TrimSpace(beforeCursor) != "" {
		if err := cursor.Decode(beforeCursor, &anchor); err == nil && anchor.FP == fp {
			return sellerPagingBackward, anchor
		}
		return sellerPagingFirstPage, anchor
	}
	if strings.TrimSpace(afterCursor) != "" {
		if err := cursor.Decode(afterCursor, &anchor); err == nil && anchor.FP == fp {
			return sellerPagingForward, anchor
		}
		return sellerPagingFirstPage, anchor
	}
	return sellerPagingFirstPage, anchor
}

// sellerProductsBaseQuery returns the shared filter for seller product list (count + list).
func (r *productRepository) sellerProductsBaseQuery(ctx context.Context, sellerID uint, search string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Product{}).
		Joins("LEFT JOIN categories ON categories.id = products.category_id").
		Where("products.seller_id = ?", sellerID)
	if search != "" {
		st := "%" + strings.ToLower(search) + "%"
		q = q.Where("(LOWER(products.name) LIKE ? OR LOWER(categories.name) LIKE ?)", st, st)
	}
	return q
}

func sellerSeekAfterNewerThanAnchor(q *gorm.DB, anchor cursor.SellerProductCursor) *gorm.DB {
	t := time.UnixMilli(anchor.CA).UTC()
	return q.Where("(products.created_at > ?) OR (products.created_at = ? AND products.id > ?)", t, t, anchor.ID)
}

func sellerSeekBeforeOlderThanAnchor(q *gorm.DB, anchor cursor.SellerProductCursor) *gorm.DB {
	t := time.UnixMilli(anchor.CA).UTC()
	return q.Where("(products.created_at < ?) OR (products.created_at = ? AND products.id < ?)", t, t, anchor.ID)
}

func sellerOrderNewestFirst(q *gorm.DB) *gorm.DB {
	return q.Order("products.created_at DESC, products.id DESC")
}

func sellerOrderOldestFirst(q *gorm.DB) *gorm.DB {
	return q.Order("products.created_at ASC, products.id ASC")
}
