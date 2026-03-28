package models

import "gorm.io/gorm"

// PlatformSetting stores singleton platform configuration (single row, id=1).
type PlatformSetting struct {
	gorm.Model
	MaintenanceMode         bool    `json:"maintenanceMode" gorm:"default:false"`
	SiteName                string  `json:"siteName" gorm:"size:200;default:'E-Commerce'"`
	CommissionRate          float64 `json:"commissionRate" gorm:"default:0"` // GP %
	Currency                string  `json:"currency" gorm:"size:10;default:'THB'"`
	ManualProductApproval   bool    `json:"manualProductApproval" gorm:"default:false"`
}

func (PlatformSetting) TableName() string {
	return "platform_settings"
}
