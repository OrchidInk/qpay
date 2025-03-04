// models/invoice.go

package models

import (
	"context"
	"net/http"
	"os"
	"qpay/config"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var httpClient = &http.Client{}
var validate = validator.New()


type QpayState string

const (
	Unpaid QpayState = "unpaid"
	Paid   QpayState = "paid"
	Failed QpayState = "failed"
	Pending QpayState = "pending"
)


type Invoice struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IpAddress     string         `json:"ipAddress" gorm:"type:varchar(45);not null"`
	CalledAt      time.Time      `json:"calledAt" gorm:"not null"`
	ExpireAt      *time.Time     `json:"expireAt,omitempty"`
	InvoiceNumber string         `json:"invoiceNumber" gorm:"unique;not null"`
	Request  	  []byte 		 `json:"-" gorm:"type:jsonb"`
    Response 	  []byte 		 `json:"-" gorm:"type:jsonb"`
	InvoiceID     string         `json:"invoiceID" gorm:"unique;not null"`
	State         QpayState      `json:"state" gorm:"type:varchar(50);not null;default:'unpaid'"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	CallbackURL   string         `json:"callbackUrl,omitempty" gorm:"type:text"`
}

func MigrateInvoiceModel() {
	err := config.DB.AutoMigrate(&Invoice{})
	if err != nil {
		panic("❌ Failed to migrate Invoice model: " + err.Error())
	}
}

func (i *Invoice) Create(ctx context.Context) (err error) {
    if err = validate.Struct(i); err != nil {
        return
    }
    err = config.DB.WithContext(ctx).Model(&Invoice{}).Create(i).Error
    return
}

func (i *Invoice) Read(ctx context.Context) (err error) {
	err = config.DB.WithContext(ctx).First(i, "id = ?", i.InvoiceID).Error
	return
}

func (i *Invoice) ReadForInvoiceID(ctx context.Context) (err error) {
	err = config.DB.WithContext(ctx).First(i, "invoice_id = ?", i.InvoiceID).Error
	return
}

func (i *Invoice) Update(ctx context.Context, vals Invoice) (err error) {
	if err = validate.Struct(vals); err != nil {
		return
	}

	err = config.DB.WithContext(ctx).Model(i).Updates(vals).Error
	if err != nil {
		return
	}

	err = i.Read(ctx)
	return
}

func (i *Invoice) Delete(ctx context.Context) (err error) {
	err = config.DB.WithContext(ctx).Delete(i).Error
	return
}

func (i *Invoice) GenerateCallbackURL() string {
	return os.Getenv("URL") + "/api/v1/invoices/callback/" + i.ID.String()
}

func (i *Invoice) CallCallbackURL() {
	if i.CallbackURL != "" {
		request, err := http.NewRequest("GET", i.CallbackURL, nil)
		if err == nil {
			httpClient.Do(request)
		}
	}
}

// // SetRequestJSON sets the Request field as JSON
// func (i *Invoice) SetRequestJSON(req map[string]interface{}) error {
// 	jsonData, err := json.Marshal(req)
// 	if err != nil {
// 		return err
// 	}
// 	i.Request = jsonData
// 	return nil
// }

// // SetResponseJSON sets the Response field as JSON
// func (i *Invoice) SetResponseJSON(res map[string]interface{}) error {
// 	jsonData, err := json.Marshal(res)
// 	if err != nil {
// 		return err
// 	}
// 	i.Response = jsonData
// 	return nil
// }
