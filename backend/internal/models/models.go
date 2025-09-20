package models

import "time"

const (
	// Collection names
	TagsCollection      = "tags"
	PaymentsCollection  = "payments"
	DocumentsCollection = "documents"
)

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Color     string    `json:"color" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID          string    `json:"id"`
	Info        string    `json:"info" binding:"required"`
	Amount      float64   `json:"amount" binding:"required"`
	DatePaid    time.Time `json:"datePaid" binding:"required"`
	FullyPaid   bool      `json:"fullyPaid"`
	InvoicePath string    `json:"invoicePath,omitempty"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Document struct {
	ID           string    `json:"id"`
	Title        string    `form:"title" json:"title" binding:"required"`
	Description  string    `form:"description" json:"description"`
	FilePath     string    `json:"file_path"`
	OriginalName string    `json:"original_name"`
	FileSize     int64     `json:"file_size"`
	Tags         []string  `form:"tags" json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FileInfo struct {
	FileName     string `json:"file_name"`
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	OriginalName string `json:"original_name"`
}
