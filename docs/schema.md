# Database Schema Documentation

## Overview

The application uses SQLite as its database, with the following tables to manage expenses, documents, and tags.

## Tables

### tags

Stores tag information used to categorize both payments and documents.

```sql
CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Column     | Type     | Description                    |
| ---------- | -------- | ------------------------------ |
| id         | TEXT     | Unique identifier (UUID)       |
| name       | TEXT     | Tag name                       |
| color      | TEXT     | Hex color code (e.g., #FF0000) |
| created_at | DATETIME | Record creation timestamp      |

### payments

Stores payment information and invoice attachments.

```sql
CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    info TEXT NOT NULL,
    amount REAL NOT NULL,
    date_paid DATE NOT NULL,
    fully_paid BOOLEAN DEFAULT false,
    invoice_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Column       | Type     | Description                            |
| ------------ | -------- | -------------------------------------- |
| id           | TEXT     | Unique identifier (UUID)               |
| info         | TEXT     | Payment description                    |
| amount       | REAL     | Payment amount                         |
| date_paid    | DATE     | Date when payment was made             |
| fully_paid   | BOOLEAN  | Whether payment is fully completed     |
| invoice_path | TEXT     | Path to stored invoice file (optional) |
| created_at   | DATETIME | Record creation timestamp              |

### payment_tags

Junction table for many-to-many relationship between payments and tags.

```sql
CREATE TABLE payment_tags (
    payment_id TEXT,
    tag_id TEXT,
    PRIMARY KEY (payment_id, tag_id),
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

| Column     | Type | Description                 |
| ---------- | ---- | --------------------------- |
| payment_id | TEXT | Reference to payments table |
| tag_id     | TEXT | Reference to tags table     |

### documents

Stores document information and file metadata.

```sql
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    file_path TEXT NOT NULL,
    original_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Column        | Type     | Description               |
| ------------- | -------- | ------------------------- |
| id            | TEXT     | Unique identifier (UUID)  |
| title         | TEXT     | Document title            |
| description   | TEXT     | Document description      |
| file_path     | TEXT     | Path to stored file       |
| original_name | TEXT     | Original filename         |
| file_size     | INTEGER  | File size in bytes        |
| created_at    | DATETIME | Record creation timestamp |

### document_tags

Junction table for many-to-many relationship between documents and tags.

```sql
CREATE TABLE document_tags (
    document_id TEXT,
    tag_id TEXT,
    PRIMARY KEY (document_id, tag_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

| Column      | Type | Description                  |
| ----------- | ---- | ---------------------------- |
| document_id | TEXT | Reference to documents table |
| tag_id      | TEXT | Reference to tags table      |

## Indexes

```sql
CREATE INDEX idx_payments_date ON payments(date_paid);
CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_documents_title ON documents(title);
```

## File Storage

Document and invoice files are stored in the filesystem:

- Documents: `./storage/documents/{id}{ext}`
- Invoices: `./storage/invoices/{id}{ext}`

The file paths are stored in the database relative to the storage directory.
