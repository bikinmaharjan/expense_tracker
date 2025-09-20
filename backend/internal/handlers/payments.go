package handlers

import (
	"database/sql"
	"expense_tracker/internal/models"
	"expense_tracker/internal/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	db *sql.DB
}

func NewPaymentHandler(db *sql.DB) *PaymentHandler {
	return &PaymentHandler{db: db}
}

// RegisterRoutes registers all payment-related routes
func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.GET("", h.ListPayments)
		payments.POST("", h.CreatePayment)
		payments.GET("/:id", h.GetPayment)
		payments.PUT("/:id", h.UpdatePayment)
		payments.DELETE("/:id", h.DeletePayment)
		payments.POST("/:id/invoice", h.UploadInvoice)
		// Serve uploaded invoice files
		payments.GET("/:id/invoice", h.DownloadInvoice)
		payments.GET("/analytics", h.GetPaymentAnalytics)
	}
}

// ListPayments returns all payments with optional filtering
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	query := `
		SELECT DISTINCT
			p.id, p.info, p.amount, p.date_paid as datePaid, p.fully_paid as fullyPaid,
			p.invoice_path as invoicePath, p.created_at as createdAt, p.updated_at as updatedAt,
			GROUP_CONCAT(pt.tag_id) as tag_ids
		FROM payments p
		LEFT JOIN payment_tags pt ON p.id = pt.payment_id
	`
	params := []interface{}{}
	whereClause := []string{}

	// Filter by tag if provided
	if tagID := c.Query("tag"); tagID != "" {
		whereClause = append(whereClause, "EXISTS (SELECT 1 FROM payment_tags WHERE payment_id = p.id AND tag_id = ?)")
		params = append(params, tagID)
	}

	// Filter by date range
	if startDate := c.Query("start_date"); startDate != "" {
		whereClause = append(whereClause, "date_paid >= ?")
		params = append(params, startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		whereClause = append(whereClause, "date_paid <= ?")
		params = append(params, endDate)
	}

	// Filter by payment status
	if status := c.Query("fully_paid"); status != "" {
		whereClause = append(whereClause, "fully_paid = ?")
		params = append(params, status == "true")
	}

	// Add WHERE clause if filters exist
	if len(whereClause) > 0 {
		query += " WHERE " + utils.JoinWithAND(whereClause)
	}

	query += " GROUP BY p.id ORDER BY p.date_paid DESC LIMIT ? OFFSET ?"

	// Add pagination params
	page := utils.ParseIntWithDefault(c.Query("page"), 1)
	limit := utils.ParseIntWithDefault(c.Query("limit"), 10)
	offset := (page - 1) * limit
	params = append(params, limit, offset)

	// Execute query
	rows, err := h.db.Query(query, params...)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch payments")
		return
	}
	defer rows.Close()

	payments := make([]models.Payment, 0) // Initialize as empty slice
	for rows.Next() {
		var p models.Payment
		var tagIDs sql.NullString
		err := rows.Scan(
			&p.ID, &p.Info, &p.Amount, &p.DatePaid, &p.FullyPaid,
			&p.InvoicePath, &p.CreatedAt, &p.UpdatedAt, &tagIDs,
		)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan payment")
			return
		}

		if tagIDs.Valid {
			p.Tags = utils.SplitCommaString(tagIDs.String)
		}
		payments = append(payments, p)
	}

	// Get total count for pagination
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM payments").Scan(&total)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get total count")
		return
	}

	// Include payment stats if requested
	var stats gin.H
	if c.Query("stats") == "true" {
		var totalAmount, monthlyAmount float64
		var pendingCount int

		// Get total amount and pending count
		err = h.db.QueryRow(`
	        SELECT
	            COALESCE(SUM(amount), 0),
	            COALESCE(SUM(CASE WHEN NOT fully_paid THEN 1 ELSE 0 END), 0)
	        FROM payments
	    `).Scan(&totalAmount, &pendingCount)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get payment stats")
			return
		}

		// Get monthly amount
		currentMonth := time.Now().Format("2006-01")
		err = h.db.QueryRow(`
	        SELECT COALESCE(SUM(amount), 0)
	        FROM payments
	        WHERE strftime('%Y-%m', date_paid) = ?
	    `, currentMonth).Scan(&monthlyAmount)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get monthly stats")
			return
		}

		stats = gin.H{
			"total":   totalAmount,
			"pending": pendingCount,
			"monthly": monthlyAmount,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"results": payments,
		"stats":   stats,
	})
}

// CreatePayment creates a new payment
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var payment models.Payment
	var payload struct {
		Info      string   `json:"info"`
		Amount    float64  `json:"amount"`
		DatePaid  string   `json:"datePaid"`
		FullyPaid bool     `json:"fullyPaid"`
		Tags      []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid payment data")
		return
	}

	// Parse date string to time.Time
	datePaid, err := time.Parse("2006-01-02", payload.DatePaid)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid date format")
		return
	}

	payment.Info = payload.Info
	payment.Amount = payload.Amount
	payment.DatePaid = datePaid
	payment.FullyPaid = payload.FullyPaid
	payment.Tags = payload.Tags

	payment.ID = uuid.New().String()
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Insert payment
	_, err = tx.Exec(`
		INSERT INTO payments (id, info, amount, date_paid, fully_paid, invoice_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		payment.ID, payment.Info, payment.Amount, payment.DatePaid,
		payment.FullyPaid, payment.InvoicePath, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to create payment")
		return
	}

	// Insert tags
	for _, tagID := range payment.Tags {
		_, err = tx.Exec("INSERT INTO payment_tags (payment_id, tag_id) VALUES (?, ?)", payment.ID, tagID)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to associate tags")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPayment returns a specific payment by ID
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")

	var payment models.Payment
	var tagIDs sql.NullString

	err := h.db.QueryRow(`
		SELECT 
			p.id, p.info, p.amount, p.date_paid, p.fully_paid,
			p.invoice_path, p.created_at, p.updated_at,
			GROUP_CONCAT(pt.tag_id) as tag_ids
		FROM payments p
		LEFT JOIN payment_tags pt ON p.id = pt.payment_id
		WHERE p.id = ?
		GROUP BY p.id
	`, id).Scan(
		&payment.ID, &payment.Info, &payment.Amount, &payment.DatePaid,
		&payment.FullyPaid, &payment.InvoicePath, &payment.CreatedAt,
		&payment.UpdatedAt, &tagIDs,
	)

	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Payment not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch payment")
		return
	}

	if tagIDs.Valid {
		payment.Tags = utils.SplitCommaString(tagIDs.String)
	}

	c.JSON(http.StatusOK, payment)
}

// UpdatePayment updates a specific payment
func (h *PaymentHandler) UpdatePayment(c *gin.Context) {
	id := c.Param("id")

	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid payment data")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update payment
	result, err := tx.Exec(`
		UPDATE payments 
		SET info = ?, amount = ?, date_paid = ?, fully_paid = ?, updated_at = ?
		WHERE id = ?
	`,
		payment.Info, payment.Amount, payment.DatePaid,
		payment.FullyPaid, time.Now(), id,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to update payment")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
		return
	}

	if rowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Payment not found")
		return
	}

	// Update tags
	_, err = tx.Exec("DELETE FROM payment_tags WHERE payment_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to remove old tags")
		return
	}

	for _, tagID := range payment.Tags {
		_, err = tx.Exec("INSERT INTO payment_tags (payment_id, tag_id) VALUES (?, ?)", id, tagID)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to associate tags")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	payment.ID = id
	c.JSON(http.StatusOK, payment)
}

// DeletePayment deletes a specific payment
func (h *PaymentHandler) DeletePayment(c *gin.Context) {
	id := c.Param("id")

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Get the payment to check if it has an invoice
	var invoicePath sql.NullString
	err = tx.QueryRow("SELECT invoice_path FROM payments WHERE id = ?", id).Scan(&invoicePath)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Payment not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch payment")
		return
	}

	// Delete invoice file if it exists
	if invoicePath.Valid && invoicePath.String != "" {
		if err := utils.DeleteFile(invoicePath.String); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete invoice file")
			return
		}
	}

	// Delete payment tags
	_, err = tx.Exec("DELETE FROM payment_tags WHERE payment_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete payment tags")
		return
	}

	// Delete payment
	result, err := tx.Exec("DELETE FROM payments WHERE id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete payment")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
		return
	}

	if rowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Payment not found")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	c.Status(http.StatusNoContent)
}

// UploadInvoice handles invoice file upload for a payment
func (h *PaymentHandler) UploadInvoice(c *gin.Context) {
	id := c.Param("id")

	// Check if payment exists
	var payment models.Payment
	err := h.db.QueryRow("SELECT id, invoice_path FROM payments WHERE id = ?", id).Scan(&payment.ID, &payment.InvoicePath)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Payment not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch payment")
		return
	}

	// Handle file upload
	file, err := c.FormFile("invoice")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "No file uploaded")
		return
	}

	// Generate unique filename
	filename := filepath.Join("storage", "invoices", uuid.New().String()+filepath.Ext(file.Filename))

	// Save file
	if err := c.SaveUploadedFile(file, filename); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to save file")
		return
	}

	// Delete old invoice if it exists
	if payment.InvoicePath != "" {
		if err := utils.DeleteFile(payment.InvoicePath); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete old invoice")
			return
		}
	}

	// Update payment with new invoice path
	_, err = h.db.Exec("UPDATE payments SET invoice_path = ?, updated_at = ? WHERE id = ?",
		filename, time.Now(), id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to update payment")
		return
	}

	fileInfo := models.FileInfo{
		FileName:     filepath.Base(filename),
		FilePath:     filename,
		FileSize:     file.Size,
		ContentType:  file.Header.Get("Content-Type"),
		OriginalName: file.Filename,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invoice uploaded successfully",
		"file_info":  fileInfo,
		"payment_id": id,
	})
}

// DownloadInvoice serves the uploaded invoice file for a payment
func (h *PaymentHandler) DownloadInvoice(c *gin.Context) {
	id := c.Param("id")

	// Fetch payment to get invoice path
	var invoicePath sql.NullString
	err := h.db.QueryRow("SELECT invoice_path FROM payments WHERE id = ?", id).Scan(&invoicePath)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Payment not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch payment")
		return
	}

	if !invoicePath.Valid || invoicePath.String == "" {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Invoice not found")
		return
	}

	// Check file exists
	if _, err := os.Stat(invoicePath.String); os.IsNotExist(err) {
		utils.RespondWithError(c, http.StatusNotFound, err, "Invoice file not found on disk")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to access invoice file")
		return
	}

	// Serve the file as an attachment or inline depending on client preference
	c.File(invoicePath.String)
}

// GetPaymentAnalytics returns analytics data for payments
func (h *PaymentHandler) GetPaymentAnalytics(c *gin.Context) {
	type monthlyStats struct {
		Year   string  `json:"year"`
		Month  string  `json:"month"`
		Amount float64 `json:"amount"`
		Count  int     `json:"count"`
	}

	type tagStats struct {
		TagName  string  `json:"tag_name"`
		TagColor string  `json:"tag_color"`
		Amount   float64 `json:"amount"`
		Count    int     `json:"count"`
	}

	// Initialize response structure with empty arrays
	stats := struct {
		TotalStats struct {
			TotalAmount  float64 `json:"total_amount"`
			TotalCount   int     `json:"total_count"`
			PaidAmount   float64 `json:"paid_amount"`
			UnpaidAmount float64 `json:"unpaid_amount"`
		} `json:"total_stats"`
		MonthlyStats []monthlyStats `json:"monthly_stats"`
		TagStats     []tagStats     `json:"tag_stats"`
	}{
		MonthlyStats: make([]monthlyStats, 0),
		TagStats:     make([]tagStats, 0),
	}

	// Check if there are any payments
	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM payments").Scan(&count); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get payment count")
		return
	}

	if count == 0 {
		c.JSON(http.StatusOK, stats)
		return
	}

	// Get total stats for non-empty database
	err := h.db.QueryRow(`
		SELECT
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as total_count,
			COALESCE(SUM(CASE WHEN fully_paid THEN amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN NOT fully_paid THEN amount ELSE 0 END), 0) as unpaid_amount
		FROM payments
	`).Scan(
		&stats.TotalStats.TotalAmount,
		&stats.TotalStats.TotalCount,
		&stats.TotalStats.PaidAmount,
		&stats.TotalStats.UnpaidAmount,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get total stats")
		return
	}

	// Get monthly stats
	rows, err := h.db.Query(`
		WITH payment_months AS (
			SELECT
				substr(date_paid, 1, 4) as year,
				substr(date_paid, 6, 2) as month,
				amount
			FROM payments
			WHERE date_paid IS NOT NULL
		)
		SELECT
			year,
			month,
			COALESCE(SUM(amount), 0) as amount,
			COUNT(*) as count
		FROM payment_months
		GROUP BY year, month
		ORDER BY year DESC, month DESC
	`)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get monthly stats")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat monthlyStats
		if err := rows.Scan(&stat.Year, &stat.Month, &stat.Amount, &stat.Count); err != nil {
			log.Printf("Error scanning monthly stats: %v", err)
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan monthly stats")
			return
		}
		stats.MonthlyStats = append(stats.MonthlyStats, stat)
	}

	// Get tag stats
	rows, err = h.db.Query(`
		SELECT 
			t.name as tag_name,
			t.color as tag_color,
			SUM(p.amount) as amount,
			COUNT(DISTINCT p.id) as count
		FROM tags t
		JOIN payment_tags pt ON t.id = pt.tag_id
		JOIN payments p ON pt.payment_id = p.id
		GROUP BY t.id, t.name, t.color
	`)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get tag stats")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat tagStats
		if err := rows.Scan(&stat.TagName, &stat.TagColor, &stat.Amount, &stat.Count); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan tag stats")
			return
		}
		stats.TagStats = append(stats.TagStats, stat)
	}

	c.JSON(http.StatusOK, stats)
}
