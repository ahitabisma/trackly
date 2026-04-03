package dto

import (
	"time"
	"trackly-backend/internal/model"
)

type CompanyResponse struct {
	ID                int        `json:"id"`
	Kode              string     `json:"kode"`
	NamaPerusahaan    string     `json:"nama_perusahaan"`
	TanggalPencatatan *time.Time `json:"tanggal_pencatatan,omitempty"`
	JumlahSaham       *int64     `json:"jumlah_saham,omitempty"`
	PapanPencatatan   *string    `json:"papan_pencatatan,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateCompanyRequest struct {
	Kode              string  `json:"kode" binding:"required"`
	NamaPerusahaan    string  `json:"nama_perusahaan" binding:"required"`
	TanggalPencatatan *string `json:"tanggal_pencatatan"`
	JumlahSaham       *int64  `json:"jumlah_saham"`
	PapanPencatatan   *string `json:"papan_pencatatan"`
}

type UpdateCompanyRequest struct {
	Kode              *string `json:"kode"`
	NamaPerusahaan    *string `json:"nama_perusahaan"`
	TanggalPencatatan *string `json:"tanggal_pencatatan"`
	JumlahSaham       *int64  `json:"jumlah_saham"`
	PapanPencatatan   *string `json:"papan_pencatatan"`
}

func ToCompanyResponse(company *model.Company) *CompanyResponse {
	if company == nil {
		return nil
	}

	return &CompanyResponse{
		ID:                company.ID,
		Kode:              company.Kode,
		NamaPerusahaan:    company.NamaPerusahaan,
		TanggalPencatatan: company.TanggalPencatatan,
		JumlahSaham:       company.JumlahSaham,
		PapanPencatatan:   company.PapanPencatatan,
		CreatedAt:         company.CreatedAt,
		UpdatedAt:         company.UpdatedAt,
	}
}

func ToCompanyResponseList(companies []model.Company) []CompanyResponse {
	if companies == nil {
		return []CompanyResponse{}
	}

	res := make([]CompanyResponse, len(companies))
	for i, company := range companies {
		resp := ToCompanyResponse(&company)
		if resp != nil {
			res[i] = *resp
		}
	}
	return res
}
