package handlers

import (
	"database/sql"
	"expense_tracker/internal/models"
	"expense_tracker/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TagHandler struct {
	db *sql.DB
}

func NewTagHandler(db *sql.DB) *TagHandler {
	return &TagHandler{db: db}
}

// RegisterRoutes registers all tag-related routes
func (h *TagHandler) RegisterRoutes(router *gin.RouterGroup) {
	tags := router.Group("/tags")
	{
		tags.GET("", h.ListTags)
		tags.POST("", h.CreateTag)
		tags.GET("/:id", h.GetTag)
		tags.PUT("/:id", h.UpdateTag)
		tags.DELETE("/:id", h.DeleteTag)
		tags.GET("/stats", h.GetTagStats)
	}
}

// ListTags returns all tags
func (h *TagHandler) ListTags(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, name, color, created_at FROM tags")
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch tags")
		return
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan tags")
			return
		}
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, tags)
}

// CreateTag creates a new tag
func (h *TagHandler) CreateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid tag data")
		return
	}

	tag.ID = uuid.New().String()
	tag.CreatedAt = time.Now()

	_, err := h.db.Exec(
		"INSERT INTO tags (id, name, color, created_at) VALUES (?, ?, ?, ?)",
		tag.ID, tag.Name, tag.Color, tag.CreatedAt,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to create tag")
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// GetTag returns a specific tag by ID
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")

	var tag models.Tag
	err := h.db.QueryRow(
		"SELECT id, name, color, created_at FROM tags WHERE id = ?",
		id,
	).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)

	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Tag not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch tag")
		return
	}

	c.JSON(http.StatusOK, tag)
}

// UpdateTag updates a specific tag
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")

	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid tag data")
		return
	}

	result, err := h.db.Exec(
		"UPDATE tags SET name = ?, color = ? WHERE id = ?",
		tag.Name, tag.Color, id,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to update tag")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
		return
	}

	if rowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Tag not found")
		return
	}

	tag.ID = id
	c.JSON(http.StatusOK, tag)
}

// DeleteTag deletes a specific tag
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Remove tag from payment_tags
	_, err = tx.Exec("DELETE FROM payment_tags WHERE tag_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to remove tag from payments")
		return
	}

	// Remove tag from document_tags
	_, err = tx.Exec("DELETE FROM document_tags WHERE tag_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to remove tag from documents")
		return
	}

	// Delete the tag
	result, err := tx.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete tag")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
		return
	}

	if rowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Tag not found")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTagStats returns usage statistics for tags
func (h *TagHandler) GetTagStats(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT 
			t.id,
			t.name,
			t.color,
			COUNT(DISTINCT pt.payment_id) as payment_count,
			COUNT(DISTINCT dt.document_id) as document_count,
			COALESCE(SUM(p.amount), 0) as total_amount
		FROM tags t
		LEFT JOIN payment_tags pt ON t.id = pt.tag_id
		LEFT JOIN payments p ON pt.payment_id = p.id
		LEFT JOIN document_tags dt ON t.id = dt.tag_id
		GROUP BY t.id, t.name, t.color
	`)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch tag stats")
		return
	}
	defer rows.Close()

	var stats []gin.H
	for rows.Next() {
		var (
			id           string
			name         string
			color        string
			paymentCount int
			docCount     int
			totalAmount  float64
		)
		if err := rows.Scan(&id, &name, &color, &paymentCount, &docCount, &totalAmount); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan tag stats")
			return
		}
		stats = append(stats, gin.H{
			"id":             id,
			"name":           name,
			"color":          color,
			"payment_count":  paymentCount,
			"document_count": docCount,
			"total_amount":   totalAmount,
		})
	}

	c.JSON(http.StatusOK, stats)
}
