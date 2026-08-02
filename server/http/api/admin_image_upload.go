package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminUploadImage uploads an image to Aliyun OSS and returns the public CDN URL.
// @Summary Upload image (admin)
// @Tags admin-images
// @Accept mpfd
// @Produce json
// @Param file formData file true "Image file (jpg/png/webp/gif, max 5MB)"
// @Success 200 {object} data.StandardResponse{data=data.ImageUploadData} "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "oss not configured"
// @Failure 500 {object} data.StandardResponse "internal error"
// @Router /admin/images/upload [post]
func AdminUploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "file is required", nil)
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	defer f.Close()

	out, err := service.GetImageUploadService().Upload(c.Request.Context(), fileHeader.Filename, f, fileHeader.Size)
	if err != nil {
		if errors.Is(err, proxy.ErrOSSNotConfigured) {
			data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
			return
		}
		msg := err.Error()
		if strings.Contains(msg, "too large") ||
			strings.Contains(msg, "unsupported") ||
			strings.Contains(msg, "empty file") {
			data.JSON(c, http.StatusBadRequest, -1, msg, nil)
			return
		}
		data.JSON(c, http.StatusInternalServerError, -1, msg, nil)
		return
	}
	data.OK(c, out)
}
