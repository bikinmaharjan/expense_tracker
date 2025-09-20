# Expense Tracker

An application to track expenses for house building process.

## Features

- Track payments with detailed information (amount, date, status)
- Categorize expenses with color-coded tags
- Store and manage documents and invoices
- Analytics dashboard with monthly statistics

## API Endpoints

### Tags

- `GET /api/tags` - List all tags
- `POST /api/tags` - Create a new tag
- `GET /api/tags/:id` - Get tag details
- `PUT /api/tags/:id` - Update a tag
- `DELETE /api/tags/:id` - Delete a tag
- `GET /api/tags/stats` - Get tag statistics

### Payments

- `GET /api/payments` - List payments (supports pagination with `limit` and `sort`)
- `POST /api/payments` - Create a new payment
- `GET /api/payments/:id` - Get payment details
- `PUT /api/payments/:id` - Update a payment
- `DELETE /api/payments/:id` - Delete a payment
- `POST /api/payments/:id/invoice` - Upload invoice for a payment
- `GET /api/payments/analytics` - Get payment analytics and statistics

### Documents

- `GET /api/documents` - List all documents
- `POST /api/documents` - Upload a new document
- `GET /api/documents/:id` - Get document details
- `PUT /api/documents/:id` - Update document details
- `DELETE /api/documents/:id` - Delete a document
- `GET /api/documents/download/:id` - Download a document

## Query Parameters

### Payments List

- `limit` - Number of items per page (default: 10)
- `page` - Page number (default: 1)
- `sort` - Sort field (e.g., `-datePaid` for descending order by date)
- `tag` - Filter by tag ID
- `start_date` - Filter by start date
- `end_date` - Filter by end date
- `fully_paid` - Filter by payment status (true/false)

## Development

### Backend (Go)

1. Navigate to backend directory:
   ```bash
   cd backend
   ```
2. Run the server:
   ```bash
   go run cmd/api/main.go
   ```
   Server will start on http://localhost:8080

### Database

The application uses SQLite database stored in `data/expense_tracker.db`
