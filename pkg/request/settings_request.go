package request

type UpdatePlatformSettingsRequest struct {
	MaintenanceMode       *bool    `json:"maintenanceMode"`
	SiteName                *string  `json:"siteName"`
	CommissionRate        *float64 `json:"commissionRate"`
	Currency                *string  `json:"currency"`
	ManualProductApproval   *bool    `json:"manualProductApproval"`
}

type UpdateSellerShopRequest struct {
	ShopName      *string `json:"shopName"`
	Description   *string `json:"description"`
	LogoURL       *string `json:"logoUrl"`
	PickupAddress *string `json:"pickupAddress"`
	BankName      *string `json:"bankName"`
	AccountNumber *string `json:"accountNumber"`
	AccountHolder *string `json:"accountHolder"`
}
