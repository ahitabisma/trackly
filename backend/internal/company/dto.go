package company

import "time"

type CompanyResponse struct {
	ID                int        `json:"id"`
	Kode              string     `json:"kode"`
	NamaPerusahaan    string     `json:"nama_perusahaan"`
	TanggalPencatatan *time.Time `json:"tanggal_pencatatan"`
	JumlahSaham       *int64     `json:"jumlah_saham"`
	PapanPencatatan   *string    `json:"papan_pencatatan"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateCompanyRequest struct {
	Kode              string  `json:"kode" validate:"required"`
	NamaPerusahaan    string  `json:"nama_perusahaan" validate:"required"`
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

type CompanyImportRequest struct {
	Kode              string  `csv:"kode"`
	NamaPerusahaan    string  `csv:"nama_perusahaan"`
	TanggalPencatatan *string `csv:"tanggal_pencatatan"`
	JumlahSaham       *int64  `csv:"jumlah_saham"`
	PapanPencatatan   *string `csv:"papan_pencatatan"`
}

type CompanyImportResponse struct {
	TotalRows     int                   `json:"total_rows"`
	SuccessCount  int                   `json:"success_count"`
	FailureCount  int                   `json:"failure_count"`
	FailedRecords []CompanyFailedRecord `json:"failed_records,omitempty"`
}

type CompanyFailedRecord struct {
	RowNumber int    `json:"row_number"`
	Kode      string `json:"kode"`
	Reason    string `json:"reason"`
}

func ToCompanyResponse(company *Company) *CompanyResponse {
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

func ToCompanyResponseList(companies []*Company) []*CompanyResponse {
	if companies == nil {
		return nil
	}

	res := make([]*CompanyResponse, len(companies))
	for i, company := range companies {
		resp := ToCompanyResponse(company)
		if resp != nil {
			res[i] = resp
		}
	}
	return res
}
