package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type QpayState string // Enum or custom type for state

type Invoice struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IpAddress     string         `json:"ipAddress" gorm:"type:varchar(45);not null"`
	CalledAt      time.Time      `json:"calledAt" gorm:"not null"`
	ExpireAt      *time.Time     `json:"expireAt,omitempty"`
	InvoiceNumber string         `json:"invoiceNumber" gorm:"unique;not null"`
	Request       datatypes.JSON `json:"request" gorm:"type:jsonb;not null"`
	Response      datatypes.JSON `json:"response" gorm:"type:jsonb;not null"`
	InvoiceID     string         `json:"invoiceID" gorm:"unique;not null"`
	State         QpayState      `json:"state" gorm:"type:varchar(50);not null"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	CallbackURL   string         `json:"callbackUrl,omitempty" gorm:"type:text"`
}
