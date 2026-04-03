package model

import "time"

type Company struct {
	ID                int        `db:"id"`
	Kode              string     `db:"kode"`
	NamaPerusahaan    string     `db:"nama_perusahaan"`
	TanggalPencatatan *time.Time `db:"tanggal_pencatatan"`
	JumlahSaham       *int64     `db:"jumlah_saham"`
	PapanPencatatan   *string    `db:"papan_pencatatan"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}
