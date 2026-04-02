package cursor

import (
	"fmt"
	"strings"
)

// ProductListCursor is used for GET /v1/products (keyset pagination).
type ProductListCursor struct {
	ID    uint    `json:"id"`
	CA    int64   `json:"ca"` // CreatedAt Unix milliseconds
	Price float64 `json:"p"`
	Sort  string  `json:"s"` // normalized sort key, see NormalizeProductSort
	FP    string  `json:"fp"`
}

func ProductFilterFP(category, price string) string {
	return category + "\x00" + price
}

// NormalizeProductSort maps sort query (e.g. "price_asc") to a stable key.
func NormalizeProductSort(sort string) string {
	if sort == "" {
		return "created_at_desc"
	}
	parts := strings.Split(sort, "_")
	if len(parts) != 2 {
		return "created_at_desc"
	}
	col, dir := parts[0], parts[1]
	if (col != "price" && col != "created_at") || (dir != "asc" && dir != "desc") {
		return "created_at_desc"
	}
	return col + "_" + dir
}

// SellerProductCursor is used for seller product list (fixed sort: created_at DESC, id DESC).
type SellerProductCursor struct {
	ID uint   `json:"id"`
	CA int64  `json:"ca"`
	FP string `json:"fp"`
}

func SellerListFP(sellerID uint, search string) string {
	return fmt.Sprintf("%d\x00%s", sellerID, search)
}
