# Company Import API

## Overview
The Company Import endpoint allows you to bulk import companies from a CSV file.

## Endpoint
```
POST /companies/import
```

## Request
- **Content-Type**: `multipart/form-data`
- **File Field**: `file` (CSV file)

## CSV Format

The CSV file should include the following columns (tab-separated, first row is header):

| Column | Required | Format | Example |
|--------|----------|--------|---------|
| No | No | Number | 1 |
| Kode | Yes | String | AALI |
| Nama Perusahaan | Yes | String | Astra Agro Lestari Tbk. |
| Tanggal Pencatatan | No | Date (DD MMM YYYY) | 09 Des 1997 |
| Saham | No | Number (with . separator) | 1.924.688.333 |
| Papan Pencatatan | No | String | Utama |

### Example CSV File
```
No	Kode	Nama Perusahaan	Tanggal Pencatatan	Saham	Papan Pencatatan
1	AALI	Astra Agro Lestari Tbk.	09 Des 1997	1924688333	Utama
2	ABBA	Mahaka Media Tbk.	03 Apr 2002	3935892857	Pemantauan Khusus
3	ABDA	Asuransi Bina Dana Arta Tbk.	06 Jul 1989	620806680	Pemantauan Khusus
```

## Response

### Success Response (200 OK)
```json
{
  "success": true,
  "message": "Import completed",
  "content": {
    "total_rows": 3,
    "success_count": 3,
    "failure_count": 0,
    "failed_records": null
  }
}
```

### Partial Failure Response (200 OK)
```json
{
  "success": true,
  "message": "Import completed",
  "content": {
    "total_rows": 3,
    "success_count": 2,
    "failure_count": 1,
    "failed_records": [
      {
        "row_number": 3,
        "kode": "ABDA",
        "reason": "ERROR: duplicate key value violates unique constraint \"companies_kode_key\" (SQLSTATE 23505)"
      }
    ]
  }
}
```

### Error Response (400 Bad Request)
```json
{
  "success": false,
  "message": "File is required",
  "content": {
    "code": "BAD_REQUEST"
  }
}
```

## cURL Example
```bash
curl -X POST http://localhost:8080/companies/import \
  -F "file=@companies_import.csv"
```

## Notes
- The endpoint accepts a maximum file size of 10MB
- Duplicate entries (same Kode) will fail but won't stop the import process
- Failed records are returned in the response for review
- Date format must be "DD MMM YYYY" (e.g., "09 Des 1997")
- Numbers with thousands separators (dots) are automatically converted
- Supported month abbreviations in Indonesian and English (Jan, Januari, Des, Dec, etc.)
