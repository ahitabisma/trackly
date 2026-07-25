## Membuat Migration

```cmd
migrate create -ext sql -dir database/migrations -seq create_users_table
```

## Contoh Migration

### UP

```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL
);
```

### DOWN

```sql
DROP TABLE users;
```

## ▶️ Menjalankan Migration

Windows (CMD / PowerShell)

```cmd
migrate -path database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_starter" up
```

## ⏪ Rollback Migration

```cmd
migrate -path database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_starter" down 1
```

## 📊 Cek Versi Migration

```cmd
migrate -path database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_starter" version
```
