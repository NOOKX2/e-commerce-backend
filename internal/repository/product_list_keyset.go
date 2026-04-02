package repository

import (
	"strings"
	"time"

	"github.com/NOOKX2/e-commerce-backend/pkg/cursor"
	"gorm.io/gorm"
)

// productListPagingMode is how the public product list should seek rows.
type productListPagingMode int

const (
	productPagingFirstPage productListPagingMode = iota
	productPagingForward
	productPagingBackward
)

// resolveProductListPaging decodes cursors; invalid or mismatched cursors fall back to the first page.
func resolveProductListPaging(beforeCursor, afterCursor, sortKey, fp string) (productListPagingMode, cursor.ProductListCursor) {
	var anchor cursor.ProductListCursor
	if strings.TrimSpace(beforeCursor) != "" {
		if err := cursor.Decode(beforeCursor, &anchor); err == nil && anchor.Sort == sortKey && anchor.FP == fp {
			return productPagingBackward, anchor
		}
		return productPagingFirstPage, anchor
	}
	if strings.TrimSpace(afterCursor) != "" {
		if err := cursor.Decode(afterCursor, &anchor); err == nil && anchor.Sort == sortKey && anchor.FP == fp {
			return productPagingForward, anchor
		}
		return productPagingFirstPage, anchor
	}
	return productPagingFirstPage, anchor
}

// productSortAxis is the primary column used for keyset pagination on the public product list.
type productSortAxis int

const (
	productSortByCreatedAt productSortAxis = iota
	productSortByPrice
)

// parseProductSortKey maps cursor.NormalizeProductSort output to axis + descending flag.
func parseProductSortKey(sortKey string) (axis productSortAxis, desc bool) {
	switch sortKey {
	case "created_at_desc":
		return productSortByCreatedAt, true
	case "created_at_asc":
		return productSortByCreatedAt, false
	case "price_desc":
		return productSortByPrice, true
	case "price_asc":
		return productSortByPrice, false
	default:
		return productSortByCreatedAt, true
	}
}

func productListOrderForward(q *gorm.DB, axis productSortAxis, desc bool) *gorm.DB {
	col := "created_at"
	if axis == productSortByPrice {
		col = "price"
	}
	if desc {
		return q.Order(col + " DESC, id DESC")
	}
	return q.Order(col + " ASC, id ASC")
}

// productListOrderBackward is the inverse SQL order used before reversing rows (prev page).
func productListOrderBackward(q *gorm.DB, axis productSortAxis, desc bool) *gorm.DB {
	col := "created_at"
	if axis == productSortByPrice {
		col = "price"
	}
	if desc {
		return q.Order(col + " ASC, id ASC")
	}
	return q.Order(col + " DESC, id DESC")
}

func productListSeekAfter(q *gorm.DB, axis productSortAxis, desc bool, c cursor.ProductListCursor) *gorm.DB {
	t := time.UnixMilli(c.CA).UTC()
	if axis == productSortByCreatedAt {
		if desc {
			return q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", t, t, c.ID)
		}
		return q.Where("(created_at > ?) OR (created_at = ? AND id > ?)", t, t, c.ID)
	}
	if desc {
		return q.Where("(price < ?) OR (price = ? AND id < ?)", c.Price, c.Price, c.ID)
	}
	return q.Where("(price > ?) OR (price = ? AND id > ?)", c.Price, c.Price, c.ID)
}

func productListSeekBefore(q *gorm.DB, axis productSortAxis, desc bool, c cursor.ProductListCursor) *gorm.DB {
	t := time.UnixMilli(c.CA).UTC()
	if axis == productSortByCreatedAt {
		if desc {
			return q.Where("(created_at > ?) OR (created_at = ? AND id > ?)", t, t, c.ID)
		}
		return q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", t, t, c.ID)
	}
	if desc {
		return q.Where("(price > ?) OR (price = ? AND id > ?)", c.Price, c.Price, c.ID)
	}
	return q.Where("(price < ?) OR (price = ? AND id < ?)", c.Price, c.Price, c.ID)
}
