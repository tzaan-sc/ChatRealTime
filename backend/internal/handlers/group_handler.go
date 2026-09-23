package handlers

import (
	"net/http"
	"strconv"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *service.GroupService
	msgRepo      *repository.MessageRepository
}

func NewGroupHandler(groupService *service.GroupService, msgRepo *repository.MessageRepository) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		msgRepo:      msgRepo,
	}
}

// Create tạo nhóm mới
func (h *GroupHandler) Create(c *gin.Context) {
	currentUserID := c.GetString("user_id")

	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu tạo nhóm không hợp lệ: " + err.Error()})
		return
	}

	group, err := h.groupService.CreateGroup(c.Request.Context(), currentUserID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo nhóm thành công",
		"data":    group,
	})
}

// GetMyGroups lấy danh sách nhóm của user
func (h *GroupHandler) GetMyGroups(c *gin.Context) {
	currentUserID := c.GetString("user_id")

	groups, err := h.groupService.GetUserGroups(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": groups,
	})
}

// GetDetails lấy chi tiết nhóm và danh sách thành viên
func (h *GroupHandler) GetDetails(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	details, err := h.groupService.GetGroupDetails(c.Request.Context(), groupID, currentUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": details,
	})
}

// AddMembers thêm thành viên vào nhóm (chỉ Admin)
func (h *GroupHandler) AddMembers(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh sách thành viên không hợp lệ"})
		return
	}

	if err := h.groupService.AddMembers(c.Request.Context(), currentUserID, groupID, req.MemberIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thêm thành viên vào nhóm thành công",
	})
}

// RemoveMember xóa thành viên khỏi nhóm (Admin) hoặc tự rời nhóm
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.groupService.LeaveOrRemove(c.Request.Context(), currentUserID, groupID, targetUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thao tác thành công",
	})
}

// GetGroupMessages lấy lịch sử tin nhắn của nhóm hoặc theo kênh con
func (h *GroupHandler) GetGroupMessages(c *gin.Context) {
	groupID := c.Param("id")
	channelID := c.Query("channel_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)

	messages, err := h.msgRepo.GetByGroup(c.Request.Context(), groupID, channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tải tin nhắn nhóm"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

// UpdateSlowMode API cập nhật thời gian chế độ chậm (Slow Mode)
func (h *GroupHandler) UpdateSlowMode(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req struct {
		Seconds int `json:"seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if err := h.groupService.UpdateSlowMode(c.Request.Context(), currentUserID, groupID, req.Seconds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã cập nhật chế độ chậm thành công",
		"seconds": req.Seconds,
	})
}

// CreateCategory API tạo danh mục kênh mới
func (h *GroupHandler) CreateCategory(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên danh mục không hợp lệ"})
		return
	}

	cat, err := h.groupService.CreateCategory(c.Request.Context(), currentUserID, groupID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo danh mục thành công",
		"data":    cat,
	})
}

// DeleteCategory API xóa danh mục kênh
func (h *GroupHandler) DeleteCategory(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	catID := c.Param("catId")

	if err := h.groupService.DeleteCategory(c.Request.Context(), currentUserID, groupID, catID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa danh mục thành công",
	})
}

// CreateChannel API tạo kênh chat con mới
func (h *GroupHandler) CreateChannel(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req models.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thông tin kênh không hợp lệ"})
		return
	}

	ch, err := h.groupService.CreateChannel(c.Request.Context(), currentUserID, groupID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo kênh thành công",
		"data":    ch,
	})
}

// DeleteChannel API xóa kênh khỏi nhóm
func (h *GroupHandler) DeleteChannel(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	chanID := c.Param("chanId")

	if err := h.groupService.DeleteChannel(c.Request.Context(), currentUserID, groupID, chanID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa kênh thành công",
	})
}

