package handlers

import (
	"database/sql"
	"encoding/json"
	"expense_tracker/internal/models"
	"expense_tracker/internal/utils"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	db *sql.DB
}

func NewDocumentHandler(db *sql.DB) *DocumentHandler {
	return &DocumentHandler{db: db}
}

func (h *DocumentHandler) RegisterRoutes(router *gin.RouterGroup) {
	documents := router.Group("/documents")
	{
		documents.POST("", h.CreateDocument)
		documents.GET("", h.ListDocuments)
		documents.GET("/:id", h.GetDocument)
		documents.PUT("/:id", h.UpdateDocument)
		documents.DELETE("/:id", h.DeleteDocument)
		documents.GET("/:id/download", h.DownloadDocument)
	}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "File is required")
		return
	}

	var doc models.Document
	if err := c.ShouldBind(&doc); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid document data")
		return
	}

	if tagsArr := c.PostFormArray("tags"); len(tagsArr) > 0 {
		doc.Tags = tagsArr
	} else if tagsStr := c.PostForm("tags"); tagsStr != "" {
		var tagIds []string
		if err := json.Unmarshal([]byte(tagsStr), &tagIds); err == nil {
			doc.Tags = tagIds
		}
	}

	doc.ID = uuid.New().String()
	filename := filepath.Join("storage", "documents", doc.ID+filepath.Ext(file.Filename))

	if err := c.SaveUploadedFile(file, filename); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to save file")
		return
	}

	doc.FilePath = filename
	doc.OriginalName = file.Filename
	doc.FileSize = file.Size
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO documents (id, title, description, file_path, original_name, file_size, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.Title, doc.Description, doc.FilePath, doc.OriginalName, doc.FileSize, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to create document")
		return
	}

	for _, tagID := range doc.Tags {
		_, err = tx.Exec("INSERT INTO document_tags (document_id, tag_id) VALUES (?, ?)", doc.ID, tagID)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to associate tags")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           doc.ID,
		"title":        doc.Title,
		"description":  doc.Description,
		"filePath":     doc.FilePath,
		"originalName": doc.OriginalName,
		"fileSize":     doc.FileSize,
		"tags":         doc.Tags,
		"createdAt":    doc.CreatedAt,
		"updatedAt":    doc.UpdatedAt,
	})
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	query := `SELECT DISTINCT d.id, d.title, d.description, d.file_path as filePath, d.original_name as originalName, d.file_size as fileSize, d.created_at as createdAt, d.updated_at as updatedAt, GROUP_CONCAT(dt.tag_id) as tag_ids FROM documents d LEFT JOIN document_tags dt ON d.id = dt.document_id`
	var params []interface{}
	var whereClause []string

	if tag := c.Query("tag"); tag != "" {
		whereClause = append(whereClause, "EXISTS (SELECT 1 FROM document_tags WHERE document_id = d.id AND tag_id = ?)")
		params = append(params, tag)
	}

	if len(whereClause) > 0 {
		query += " WHERE " + utils.JoinWithAND(whereClause)
	}

	query += " GROUP BY d.id ORDER BY d.created_at DESC"

	limit := utils.ParseIntWithDefault(c.Query("limit"), 10)
	offset := utils.ParseIntWithDefault(c.Query("offset"), 0)
	query += " LIMIT ? OFFSET ?"
	params = append(params, limit, offset)

	rows, err := h.db.Query(query, params...)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch documents")
		return
	}
	defer rows.Close()

	documents := make([]models.Document, 0)
	for rows.Next() {
		var doc models.Document
		var tagIDs sql.NullString
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Description, &doc.FilePath, &doc.OriginalName, &doc.FileSize, &doc.CreatedAt, &doc.UpdatedAt, &tagIDs)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to scan document")
			return
		}

		if tagIDs.Valid {
			doc.Tags = utils.SplitCommaString(tagIDs.String)
		}
		documents = append(documents, doc)
	}

	resp := make([]gin.H, 0, len(documents))
	for _, d := range documents {
		resp = append(resp, gin.H{"id": d.ID, "title": d.Title, "description": d.Description, "filePath": d.FilePath, "originalName": d.OriginalName, "fileSize": d.FileSize, "tags": d.Tags, "createdAt": d.CreatedAt, "updatedAt": d.UpdatedAt})
	}

	c.JSON(http.StatusOK, gin.H{"results": resp, "total": len(documents)})
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")

	var doc models.Document
	var tagIDs sql.NullString

	err := h.db.QueryRow(`SELECT d.id, d.title, d.description, d.file_path, d.original_name, d.file_size, d.created_at, d.updated_at, GROUP_CONCAT(dt.tag_id) as tag_ids FROM documents d LEFT JOIN document_tags dt ON d.id = dt.document_id WHERE d.id = ? GROUP BY d.id`, id).Scan(&doc.ID, &doc.Title, &doc.Description, &doc.FilePath, &doc.OriginalName, &doc.FileSize, &doc.CreatedAt, &doc.UpdatedAt, &tagIDs)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Document not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch document")
		return
	}

	if tagIDs.Valid {
		doc.Tags = utils.SplitCommaString(tagIDs.String)
	}

	c.JSON(http.StatusOK, gin.H{"id": doc.ID, "title": doc.Title, "description": doc.Description, "filePath": doc.FilePath, "originalName": doc.OriginalName, "fileSize": doc.FileSize, "tags": doc.Tags, "createdAt": doc.CreatedAt, "updatedAt": doc.UpdatedAt})
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id := c.Param("id")

	var doc models.Document
	contentType := c.GetHeader("Content-Type")
	var fileHeader *multipart.FileHeader
	var fileErr error

	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		if err := c.ShouldBind(&doc); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid document data")
			return
		}

		if tagsArr := c.PostFormArray("tags"); len(tagsArr) > 0 {
			doc.Tags = tagsArr
		} else if tagsStr := c.PostForm("tags"); tagsStr != "" {
			var tagIds []string
			if err := json.Unmarshal([]byte(tagsStr), &tagIds); err == nil {
				doc.Tags = tagIds
			}
		}

		fileHeader, fileErr = c.FormFile("file")
	} else {
		if err := c.ShouldBindJSON(&doc); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, err, "Invalid document data")
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	if fileErr == nil && fileHeader != nil {
		var oldPath string
		_ = tx.QueryRow("SELECT file_path FROM documents WHERE id = ?", id).Scan(&oldPath)

		newFilename := filepath.Join("storage", "documents", id+filepath.Ext(fileHeader.Filename))
		if err := c.SaveUploadedFile(fileHeader, newFilename); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to save uploaded file")
			return
		}

		_, err = tx.Exec("UPDATE documents SET title = ?, description = ?, file_path = ?, original_name = ?, file_size = ?, updated_at = ? WHERE id = ?", doc.Title, doc.Description, newFilename, fileHeader.Filename, fileHeader.Size, time.Now(), id)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to update document with new file")
			return
		}

		if oldPath != "" && oldPath != newFilename {
			_ = utils.DeleteFile(oldPath)
		}
	} else {
		result, err := tx.Exec("UPDATE documents SET title = ?, description = ?, updated_at = ? WHERE id = ?", doc.Title, doc.Description, time.Now(), id)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to update document")
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
			return
		}

		if rowsAffected == 0 {
			utils.RespondWithError(c, http.StatusNotFound, nil, "Document not found")
			return
		}
	}

	_, err = tx.Exec("DELETE FROM document_tags WHERE document_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to remove old tags")
		return
	}

	for _, tagID := range doc.Tags {
		_, err = tx.Exec("INSERT INTO document_tags (document_id, tag_id) VALUES (?, ?)", id, tagID)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to associate tags")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	doc.ID = id
	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var filePath string
	err = tx.QueryRow("SELECT file_path FROM documents WHERE id = ?", id).Scan(&filePath)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Document not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch document")
		return
	}

	_, err = tx.Exec("DELETE FROM document_tags WHERE document_id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete document tags")
		return
	}

	result, err := tx.Exec("DELETE FROM documents WHERE id = ?", id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to delete document")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to get rows affected")
		return
	}

	if rowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, nil, "Document not found")
		return
	}

	if err := utils.DeleteFile(filePath); err != nil {
		c.Error(err)
	}

	if err := tx.Commit(); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	id := c.Param("id")

	var filePath string
	err := h.db.QueryRow("SELECT file_path FROM documents WHERE id = ?", id).Scan(&filePath)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusNotFound, err, "Document not found")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err, "Failed to fetch document")
		return
	}

	c.File(filePath)
}
