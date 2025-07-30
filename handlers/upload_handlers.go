package handlers

import (
	"database/sql"
	"fmt"
	"generalusermanagement/config"
	"generalusermanagement/database"
	"generalusermanagement/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadProfilePicture godoc
// @Summary Upload profile picture
// @Description Upload a profile picture for the authenticated user
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Profile picture file"
// @Success 200 {object} models.APIResponse{data=map[string]string}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 413 {object} models.APIResponse
// @Router /upload/profile-picture [post]
func UploadProfilePicture(c *gin.Context) {
	userID, _ := c.Get("user_id")

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	// Get the file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No file uploaded",
			Error:   err.Error(),
		})
		return
	}

	// Check file size
	if file.Size > config.AppConfig.MaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("File too large. Maximum size is %d bytes", config.AppConfig.MaxFileSize),
		})
		return
	}

	// Check file type
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	isValidType := false
	for _, allowedType := range allowedTypes {
		if fileExt == allowedType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid file type. Allowed types: " + strings.Join(allowedTypes, ", "),
		})
		return
	}

	// Create upload directory if it doesn't exist
	uploadDir := config.AppConfig.UploadPath
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create upload directory",
			Error:   err.Error(),
		})
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("profile_%s_%d%s", userUUID.String(), timestamp, fileExt)
	filePath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to save file",
			Error:   err.Error(),
		})
		return
	}

	// Update user's profile picture in database
	relativePath := filepath.Join("uploads", filename)
	updateQuery := `UPDATE users SET profile_picture = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = database.DB.Exec(updateQuery, relativePath, userUUID)
	if err != nil {
		// If database update fails, remove the uploaded file
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update profile picture in database",
			Error:   err.Error(),
		})
		return
	}

	// Generate URL for the uploaded file
	fileURL := fmt.Sprintf("/files/%s", filename)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile picture uploaded successfully",
		Data: map[string]string{
			"filename": filename,
			"url":      fileURL,
			"path":     relativePath,
		},
	})
}

// ServeFile godoc
// @Summary Serve uploaded file
// @Description Serve static files (like profile pictures)
// @Tags Upload
// @Param filename path string true "Filename"
// @Success 200 {file} file
// @Failure 404 {object} models.APIResponse
// @Router /files/{filename} [get]
func ServeFile(c *gin.Context) {
	filename := c.Param("filename")

	// Validate filename to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid filename",
		})
		return
	}

	filePath := filepath.Join(config.AppConfig.UploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "File not found",
		})
		return
	}

	// Serve the file
	c.File(filePath)
}

// DeleteProfilePicture godoc
// @Summary Delete profile picture
// @Description Delete the profile picture of the authenticated user
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /upload/profile-picture [delete]
func DeleteProfilePicture(c *gin.Context) {
	userID, _ := c.Get("user_id")

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	// Get current profile picture path
	var profilePicture sql.NullString
	query := `SELECT profile_picture FROM users WHERE id = $1 AND deleted_at IS NULL`
	err = database.DB.QueryRow(query, userUUID).Scan(&profilePicture)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if !profilePicture.Valid || profilePicture.String == "" {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "No profile picture to delete",
		})
		return
	}

	// Remove profile picture from database
	updateQuery := `UPDATE users SET profile_picture = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = database.DB.Exec(updateQuery, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to remove profile picture from database",
			Error:   err.Error(),
		})
		return
	}

	// Try to delete the actual file (don't fail if file doesn't exist)
	if profilePicture.Valid {
		filePath := filepath.Join(config.AppConfig.UploadPath, filepath.Base(profilePicture.String))
		os.Remove(filePath) // Ignore error - file might already be deleted
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile picture deleted successfully",
	})
}
