# API Documentation

## Base URL

`http://localhost:8080/api/v1`

## Authentication

Currently, the API does not require authentication.

## Endpoints

### Tags

#### List Tags

```http
GET /tags
```

Returns a list of all tags.

**Response** `200 OK`

```json
[
  {
    "id": "string",
    "name": "string",
    "color": "string"
  }
]
```

#### Create Tag

```http
POST /tags
```

Create a new tag.

**Request Body**

```json
{
  "name": "string",
  "color": "string"
}
```

**Response** `201 Created`

```json
{
  "id": "string",
  "name": "string",
  "color": "string"
}
```

### Payments

#### List Payments

```http
GET /payments
```

Returns a list of all payments.

**Response** `200 OK`

```json
[
  {
    "id": "string",
    "info": "string",
    "amount": "number",
    "tags": ["string"],
    "datePaid": "string",
    "fullyPaid": "boolean",
    "invoicePath": "string",
    "createdAt": "string"
  }
]
```

#### Create Payment

```http
POST /payments
```

Create a new payment record.

**Request Body** (multipart/form-data)

- `info`: Payment information (string)
- `amount`: Payment amount (number)
- `tags`: Array of tag IDs (JSON string)
- `datePaid`: Payment date (string, YYYY-MM-DD)
- `fullyPaid`: Payment status (boolean)
- `invoice`: Invoice file (file, optional)

**Response** `201 Created`

```json
{
  "id": "string",
  "info": "string",
  "amount": "number",
  "tags": ["string"],
  "datePaid": "string",
  "fullyPaid": "boolean",
  "invoicePath": "string",
  "createdAt": "string"
}
```

#### Download Invoice

```http
GET /payments/{id}/invoice
```

Download the invoice file for a payment.

**Response** `200 OK`
Binary file stream

### Documents

#### List Documents

```http
GET /documents
```

Returns a list of all documents.

**Response** `200 OK`

```json
[
  {
    "id": "string",
    "title": "string",
    "description": "string",
    "tags": ["string"],
    "filePath": "string",
    "originalName": "string",
    "fileSize": "number",
    "createdAt": "string"
  }
]
```

#### Upload Document

```http
POST /documents
```

Upload a new document.

**Request Body** (multipart/form-data)

- `title`: Document title (string)
- `description`: Document description (string)
- `tags`: Array of tag IDs (JSON string)
- `file`: Document file (file)

**Response** `201 Created`

```json
{
  "id": "string",
  "title": "string",
  "description": "string",
  "tags": ["string"],
  "filePath": "string",
  "originalName": "string",
  "fileSize": "number",
  "createdAt": "string"
}
```

#### Download Document

```http
GET /documents/{id}/download
```

Download a document file.

**Response** `200 OK`
Binary file stream

## Error Responses

All endpoints may return the following errors:

### 400 Bad Request

```json
{
  "error": "string",
  "message": "string"
}
```

### 404 Not Found

```json
{
  "error": "string",
  "message": "string"
}
```

### 500 Internal Server Error

```json
{
  "error": "string",
  "message": "string"
}
```
