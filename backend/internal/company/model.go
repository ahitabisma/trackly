package company

import "time"

type Company struct {
	ID                uint       `gorm:"primaryKey"`
	Kode              string     `gorm:"type:varchar(10);uniqueIndex;not null"`
	NamaPerusahaan    string     `gorm:"type:varchar(255);not null"`
	TanggalPencatatan *time.Time `gorm:"type:date;null"`
	JumlahSaham       *int64     `gorm:"null"`
	PapanPencatatan   *string    `gorm:"type:varchar(50);null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
