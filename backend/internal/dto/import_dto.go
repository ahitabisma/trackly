package dto

// CompanyImportRequest represents a single company record from CSV import
type CompanyImportRequest struct {
	Kode              string  `csv:"kode"`
	NamaPerusahaan    string  `csv:"nama_perusahaan"`
	TanggalPencatatan *string `csv:"tanggal_pencatatan"`
	JumlahSaham       *int64  `csv:"jumlah_saham"`
	PapanPencatatan   *string `csv:"papan_pencatatan"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalRows     int            `json:"total_rows"`
	SuccessCount  int            `json:"success_count"`
	FailureCount  int            `json:"failure_count"`
	FailedRecords []FailedRecord `json:"failed_records,omitempty"`
}

// FailedRecord represents a failed import record
type FailedRecord struct {
	RowNumber int    `json:"row_number"`
	Kode      string `json:"kode"`
	Reason    string `json:"reason"`
}
