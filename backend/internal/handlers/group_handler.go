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

// UpdateMemberRole API cập nhật vai trò thành viên
func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	targetUserID := c.Param("userId")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vai trò không hợp lệ"})
		return
	}

	if err := h.groupService.UpdateMemberRole(c.Request.Context(), currentUserID, groupID, targetUserID, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật vai trò thành công"})
}

// MuteMember API cấm chat hoặc bỏ cấm chat thành viên
func (h *GroupHandler) MuteMember(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	targetUserID := c.Param("userId")

	var req models.MuteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if err := h.groupService.MuteMember(c.Request.Context(), currentUserID, groupID, targetUserID, req.DurationMinutes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thao tác thành công",
		"duration_minutes": req.DurationMinutes,
	})
}

// UpdateSettings API cập nhật cài đặt duyệt thành viên và chế độ chậm
func (h *GroupHandler) UpdateSettings(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req models.UpdateGroupSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu cài đặt không hợp lệ"})
		return
	}

	if err := h.groupService.UpdateSettings(c.Request.Context(), currentUserID, groupID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã lưu cài đặt nhóm thành công"})
}

// CreateInvite API tạo liên kết mời vào nhóm
func (h *GroupHandler) CreateInvite(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	var req models.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu tạo link mời không hợp lệ"})
		return
	}

	inv, err := h.groupService.CreateInvite(c.Request.Context(), currentUserID, groupID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo liên kết mời thành công",
		"data":    inv,
	})
}

// GetGroupInvites API lấy danh sách liên kết mời của nhóm
func (h *GroupHandler) GetGroupInvites(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	list, err := h.groupService.GetGroupInvites(c.Request.Context(), currentUserID, groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

// RevokeInvite API thu hồi mã mời
func (h *GroupHandler) RevokeInvite(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	code := c.Param("code")

	if err := h.groupService.RevokeInvite(c.Request.Context(), currentUserID, groupID, code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã thu hồi liên kết mời thành công"})
}

// PreviewInvite API xem trước thông tin nhóm qua link mời
func (h *GroupHandler) PreviewInvite(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	code := c.Param("code")

	preview, err := h.groupService.PreviewInvite(c.Request.Context(), currentUserID, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": preview})
}

// JoinViaInvite API tham gia nhóm qua link mời
func (h *GroupHandler) JoinViaInvite(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	code := c.Param("code")

	isJoined, requiresApproval, err := h.groupService.JoinViaInvite(c.Request.Context(), currentUserID, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_joined":         isJoined,
		"requires_approval": requiresApproval,
		"message": func() string {
			if isJoined {
				return "Tham gia nhóm thành công"
			}
			return "Đã gửi yêu cầu tham gia. Vui lòng chờ Quản trị viên duyệt!"
		}(),
	})
}

// GetPendingJoinRequests API xem danh sách đơn chờ duyệt
func (h *GroupHandler) GetPendingJoinRequests(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")

	list, err := h.groupService.GetPendingJoinRequests(c.Request.Context(), currentUserID, groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

// ApproveJoinRequest API duyệt thành viên vào nhóm
func (h *GroupHandler) ApproveJoinRequest(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	requestID := c.Param("requestId")

	if err := h.groupService.ApproveJoinRequest(c.Request.Context(), currentUserID, groupID, requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã phê duyệt thành viên vào nhóm"})
}

// RejectJoinRequest API từ chối đơn vào nhóm
func (h *GroupHandler) RejectJoinRequest(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	groupID := c.Param("id")
	requestID := c.Param("requestId")

	if err := h.groupService.RejectJoinRequest(c.Request.Context(), currentUserID, groupID, requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã từ chối đơn tham gia"})
}


