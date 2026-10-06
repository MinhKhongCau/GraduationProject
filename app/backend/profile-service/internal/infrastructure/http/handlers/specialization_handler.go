// File: internal/infrastructure/http/handlers/specialization_handler.go
package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateSpecialization tạo mới một chuyên khoa (dành cho Admin).
// @Summary      [Admin] Tạo chuyên khoa mới
// @Description  code và slug là duy nhất; slug bỏ trống sẽ được sinh tự động từ name.
// @Tags         specializations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateSpecializationRequest true "Chuyên khoa"
// @Success      201 {object} response.BaseResponse{result=models.Specialization}
// @Failure      400 {object} response.BaseResponse{result=response.ErrorResult}
// @Failure      409 {object} response.BaseResponse{result=response.ErrorResult}
// @Router       /api/v1/profiles/specializations [post]
func CreateSpecialization(c *gin.Context) {
	var req schemas.CreateSpecializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	slug := utils.Slugify(strings.TrimSpace(req.Slug))
	if strings.TrimSpace(req.Slug) == "" {
		slug = utils.Slugify(req.Name)
	}
	symptoms := req.Symptoms
	if symptoms == nil {
		symptoms = []string{}
	}

	newSpec := models.Specialization{
		Code:        strings.ToUpper(strings.TrimSpace(req.Code)),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Symptoms:    symptoms,
		Location:    req.Location,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	var count int64
	if err := config.DB.Model(&models.Specialization{}).
		Where("code = ? OR slug = ?", newSpec.Code, newSpec.Slug).
		Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn Database", err.Error())
		return
	}
	if count > 0 {
		response.Error(c, http.StatusConflict, "Mã hoặc slug chuyên khoa đã tồn tại", "specialization code or slug already exists")
		return
	}

	if err := config.DB.Create(&newSpec).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu Database", err.Error())
		return
	}

	response.Created(c, "Tạo chuyên khoa thành công", newSpec)
}

// GetAllSpecializations trả về danh sách chuyên khoa đang hoạt động (public).
// @Summary      Danh sách chuyên khoa
// @Tags         specializations
// @Produce      json
// @Success      200 {object} response.BaseResponse{result=[]models.Specialization}
// @Router       /api/v1/profiles/specializations [get]
func GetAllSpecializations(c *gin.Context) {
	specs := []models.Specialization{}
	if err := config.DB.Where("is_active = ?", true).Order("code ASC").Find(&specs).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách chuyên khoa", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách chuyên khoa thành công", specs)
}

// ---------- Quản trị chuyên khoa (Admin) ----------

// specializationAdminView là chuyên khoa kèm số chuyên gia đang đăng ký (cho màn hình quản trị).
type specializationAdminView struct {
	models.Specialization `gorm:"embedded"`
	ExpertCount           int64 `json:"expert_count"`
}

const specializationAdminSelect = "specializations.*, " +
	"(SELECT COUNT(*) FROM expert_specializations es WHERE es.spec_id = specializations.spec_id) AS expert_count"

// ListAllSpecializations trả về mọi chuyên khoa (kể cả ngừng hoạt động) có phân trang.
// @Summary      [Admin] Danh sách tất cả chuyên khoa
// @Tags         specializations
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        search    query string false "Tìm theo tên hoặc mã"
// @Param        is_active query bool   false "Lọc theo trạng thái hoạt động"
// @Success      200 {object} response.BaseResponse{result=response.PageResult[specializationAdminView]}
// @Failure      400 {object} response.BaseResponse
// @Router       /api/v1/profiles/specializations/all [get]
func ListAllSpecializations(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Specialization{})
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if rawActive := strings.TrimSpace(c.Query("is_active")); rawActive != "" {
		isActive, err := strconv.ParseBool(rawActive)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Giá trị is_active không hợp lệ", err.Error())
			return
		}
		query = query.Where("is_active = ?", isActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách chuyên khoa", err.Error())
		return
	}

	var specs []specializationAdminView
	if err := query.Select(specializationAdminSelect).
		Order("code ASC").
		Limit(pagination.PageSize).
		Offset((pagination.Page - 1) * pagination.PageSize).
		Find(&specs).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách chuyên khoa", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách chuyên khoa thành công",
		response.NewPageResult(specs, total, pagination.Page, pagination.PageSize))
}

// GetSpecialization trả về chi tiết 1 chuyên khoa (kể cả ngừng hoạt động).
// @Summary      [Admin] Xem chi tiết chuyên khoa
// @Tags         specializations
// @Security     BearerAuth
// @Produce      json
// @Param        specId path string true "Specialization ID"
// @Success      200 {object} response.BaseResponse{result=specializationAdminView}
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/specializations/{specId} [get]
func GetSpecialization(c *gin.Context) {
	spec, ok := findSpecializationView(c)
	if !ok {
		return
	}
	response.Success(c, "Lấy thông tin chuyên khoa thành công", spec)
}

// UpdateSpecialization thay toàn bộ thông tin chuyên khoa.
// @Summary      [Admin] Cập nhật chuyên khoa
// @Description  code và slug là duy nhất; slug bỏ trống sẽ được sinh lại từ name. Không đổi trạng thái hoạt động.
// @Tags         specializations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        specId  path string                               true "Specialization ID"
// @Param        request body schemas.UpdateSpecializationRequest true "Chuyên khoa"
// @Success      200 {object} response.BaseResponse{result=specializationAdminView}
// @Failure      400 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Failure      409 {object} response.BaseResponse
// @Router       /api/v1/profiles/specializations/{specId} [put]
func UpdateSpecialization(c *gin.Context) {
	spec, ok := findSpecializationView(c)
	if !ok {
		return
	}

	var req schemas.UpdateSpecializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))
	slug := utils.Slugify(strings.TrimSpace(req.Slug))
	if strings.TrimSpace(req.Slug) == "" {
		slug = utils.Slugify(req.Name)
	}
	symptoms := req.Symptoms
	if symptoms == nil {
		symptoms = []string{}
	}

	var count int64
	if err := config.DB.Model(&models.Specialization{}).
		Where("(code = ? OR slug = ?) AND spec_id <> ?", code, slug, spec.SpecID).
		Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn Database", err.Error())
		return
	}
	if count > 0 {
		response.Error(c, http.StatusConflict, "Mã hoặc slug chuyên khoa đã tồn tại", "specialization code or slug already exists")
		return
	}

	// Updates qua struct (không phải map) để serializer:json của symptoms được áp dụng;
	// Select liệt kê cột để ghi cả giá trị rỗng.
	updates := models.Specialization{
		Code:        code,
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Symptoms:    symptoms,
		Location:    req.Location,
		ImageURL:    req.ImageURL,
	}
	if err := config.DB.Model(&spec.Specialization).
		Select("code", "name", "slug", "description", "symptoms", "location", "image_url").
		Updates(&updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu Database", err.Error())
		return
	}

	spec.Code, spec.Name, spec.Slug = code, req.Name, slug
	spec.Description, spec.Symptoms, spec.Location, spec.ImageURL = req.Description, symptoms, req.Location, req.ImageURL
	response.Success(c, "Cập nhật chuyên khoa thành công", spec)
}

// UpdateSpecializationStatus bật/tắt hoạt động của chuyên khoa.
// @Summary      [Admin] Bật/tắt chuyên khoa
// @Description  Chuyên khoa ngừng hoạt động bị ẩn khỏi danh sách public và không đăng ký mới được; chuyên gia đã đăng ký vẫn giữ.
// @Tags         specializations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        specId  path string                                     true "Specialization ID"
// @Param        request body schemas.UpdateSpecializationStatusRequest true "Trạng thái"
// @Success      200 {object} response.BaseResponse{result=specializationAdminView}
// @Failure      400 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/specializations/{specId}/status [patch]
func UpdateSpecializationStatus(c *gin.Context) {
	spec, ok := findSpecializationView(c)
	if !ok {
		return
	}

	var req schemas.UpdateSpecializationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	if err := config.DB.Model(&spec.Specialization).Update("is_active", *req.IsActive).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu Database", err.Error())
		return
	}

	spec.IsActive = *req.IsActive
	response.Success(c, "Cập nhật trạng thái chuyên khoa thành công", spec)
}

// DeleteSpecialization xoá hẳn chuyên khoa chưa có chuyên gia nào đăng ký.
// @Summary      [Admin] Xoá chuyên khoa
// @Description  Chuyên khoa đang có chuyên gia đăng ký không xoá được (409) - hãy ngừng hoạt động thay vì xoá.
// @Tags         specializations
// @Security     BearerAuth
// @Produce      json
// @Param        specId path string true "Specialization ID"
// @Success      200 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Failure      409 {object} response.BaseResponse
// @Router       /api/v1/profiles/specializations/{specId} [delete]
func DeleteSpecialization(c *gin.Context) {
	spec, ok := findSpecializationView(c)
	if !ok {
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Khoá dòng chuyên khoa để không có chuyên gia nào đăng ký chen vào giữa lúc kiểm tra và xoá.
		var locked models.Specialization
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "spec_id = ?", spec.SpecID).Error; err != nil {
			return err
		}
		var inUse int64
		if err := tx.Table("expert_specializations").Where("spec_id = ?", spec.SpecID).Count(&inUse).Error; err != nil {
			return err
		}
		if inUse > 0 {
			return errSpecializationInUse
		}
		return tx.Delete(&models.Specialization{}, "spec_id = ?", spec.SpecID).Error
	})
	if errors.Is(err, errSpecializationInUse) {
		response.Error(c, http.StatusConflict,
			"Chuyên khoa đang có chuyên gia đăng ký, hãy ngừng hoạt động thay vì xoá", err.Error())
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi xoá chuyên khoa", err.Error())
		return
	}

	response.Success(c, "Xoá chuyên khoa thành công", nil)
}

var errSpecializationInUse = errors.New("specialization is assigned to experts")

// findSpecializationView tìm chuyên khoa theo path param specId; tự trả 404 nếu không có.
func findSpecializationView(c *gin.Context) (*specializationAdminView, bool) {
	specID, err := uuid.Parse(c.Param("specId"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy chuyên khoa", err.Error())
		return nil, false
	}

	var spec specializationAdminView
	err = config.DB.Model(&models.Specialization{}).
		Select(specializationAdminSelect).
		Where("specializations.spec_id = ?", specID).
		Take(&spec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "Không tìm thấy chuyên khoa", err.Error())
		return nil, false
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn chuyên khoa", err.Error())
		return nil, false
	}
	return &spec, true
}
