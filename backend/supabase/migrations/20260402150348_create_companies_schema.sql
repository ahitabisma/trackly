CREATE TABLE companies (
    id              SERIAL PRIMARY KEY,
    kode            VARCHAR(10)  NOT NULL UNIQUE,
    nama_perusahaan VARCHAR(255) NOT NULL,
    tanggal_pencatatan DATE,
    jumlah_saham    BIGINT,
    papan_pencatatan VARCHAR(50), -- Utama / Pemantauan Khusus / Pengembangan
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);