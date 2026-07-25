CREATE TABLE companies (
    id BIGSERIAL PRIMARY KEY,
    kode VARCHAR(10) NOT NULL UNIQUE,
    nama_perusahaan VARCHAR(255) NOT NULL,
    tanggal_pencatatan DATE NULL,
    jumlah_saham BIGINT NULL,
    papan_pencatatan VARCHAR(50) NULL, -- Utama / Pemantauan Khusus / Pengembangan
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_companies_kode ON companies(kode);