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

// GetGroupMessages lấy lịch sử tin nhắn của nhóm
func (h *GroupHandler) GetGroupMessages(c *gin.Context) {
	groupID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)

	messages, err := h.msgRepo.GetByGroup(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tải tin nhắn nhóm"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}
