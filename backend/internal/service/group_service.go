package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupService struct {
	groupRepo   *repository.GroupRepository
	userRepo    *repository.UserRepository
	inviteRepo  *repository.InviteRepository
	joinReqRepo *repository.JoinRequestRepository
	pollRepo    *repository.PollRepository
	eventRepo   *repository.EventRepository
	messageRepo *repository.MessageRepository
	redisClient *redis.Client
}

func NewGroupService(
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
	inviteRepo *repository.InviteRepository,
	joinReqRepo *repository.JoinRequestRepository,
	pollRepo *repository.PollRepository,
	eventRepo *repository.EventRepository,
	messageRepo *repository.MessageRepository,
	redisClient *redis.Client,
) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		inviteRepo:  inviteRepo,
		joinReqRepo: joinReqRepo,
		pollRepo:    pollRepo,
		eventRepo:   eventRepo,
		messageRepo: messageRepo,
		redisClient: redisClient,
	}
}

// CreateGroup tạo nhóm mới
func (s *GroupService) CreateGroup(ctx context.Context, creatorID string, req models.CreateGroupRequest) (*models.Group, error) {
	creatorOID, err := primitive.ObjectIDFromHex(creatorID)
	if err != nil {
		return nil, errors.New("creator_id không hợp lệ")
	}

	memberOIDs := []primitive.ObjectID{creatorOID}
	for _, idStr := range req.MemberIDs {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err == nil && oid != creatorOID {
			memberOIDs = append(memberOIDs, oid)
		}
	}

	group := &models.Group{
		Name:      req.Name,
		Avatar:    fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=%s", req.Name),
		CreatorID: creatorOID,
		AdminIDs:  []primitive.ObjectID{creatorOID},
		MemberIDs: memberOIDs,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetUserGroups lấy danh sách các nhóm mà user tham gia
func (s *GroupService) GetUserGroups(ctx context.Context, userID string) ([]models.Group, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	return s.groupRepo.GetUserGroups(ctx, userOID)
}

// GetGroupDetails lấy thông tin nhóm kèm danh sách thành viên đầy đủ
func (s *GroupService) GetGroupDetails(ctx context.Context, groupID, userID string) (*models.GroupDetailResponse, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}

	// Kiểm tra user có trong nhóm không
	isMember, err := s.groupRepo.IsMember(ctx, groupOID, userOID)
	if err != nil || !isMember {
		return nil, errors.New("bạn không phải là thành viên của nhóm này")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return nil, err
	}

	// Lấy profile của các thành viên
	users, err := s.userRepo.FindByIDs(ctx, group.MemberIDs)
	if err != nil {
		return nil, err
	}

	adminMap := make(map[string]bool)
	for _, adminOID := range group.AdminIDs {
		adminMap[adminOID.Hex()] = true
	}
	modMap := make(map[string]bool)
	for _, modOID := range group.ModeratorIDs {
		modMap[modOID.Hex()] = true
	}

	now := time.Now()
	memberInfos := make([]models.GroupMemberInfo, 0, len(users))
	for _, u := range users {
		uHex := u.ID.Hex()
		role := "member"
		isAdmin := false
		if u.ID == group.CreatorID {
			role = "owner"
			isAdmin = true
		} else if adminMap[uHex] {
			role = "admin"
			isAdmin = true
		} else if modMap[uHex] {
			role = "moderator"
			isAdmin = false
		}

		isMuted := false
		var mutedUntil *time.Time
		if t, ok := group.MutedMembers[uHex]; ok && t.After(now) {
			isMuted = true
			tCopy := t
			mutedUntil = &tCopy
		}

		memberInfos = append(memberInfos, models.GroupMemberInfo{
			ID:          uHex,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			Role:        role,
			IsAdmin:     isAdmin,
			IsMuted:     isMuted,
			MutedUntil:  mutedUntil,
		})
	}

	res := &models.GroupDetailResponse{
		ID:              group.ID.Hex(),
		Name:            group.Name,
		Avatar:          group.Avatar,
		CreatorID:       group.CreatorID.Hex(),
		Members:         memberInfos,
		SlowModeSeconds: group.SlowModeSeconds,
		RequireApproval: group.RequireApproval,
		IsCommunity:     group.IsCommunity,
		Categories:      group.Categories,
		Channels:        group.Channels,
		CreatedAt:       group.CreatedAt,
	}

	return res, nil
}

// AddMembers thêm thành viên vào nhóm (chỉ Admin)
func (s *GroupService) AddMembers(ctx context.Context, adminID, groupID string, memberIDs []string) error {
	adminOID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return errors.New("admin_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, adminOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền thêm thành viên")
	}

	newOIDs := make([]primitive.ObjectID, 0, len(memberIDs))
	for _, idStr := range memberIDs {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err == nil {
			newOIDs = append(newOIDs, oid)
		}
	}

	if len(newOIDs) == 0 {
		return errors.New("danh sách thành viên thêm vào trống")
	}

	return s.groupRepo.AddMembers(ctx, groupOID, newOIDs)
}

// LeaveOrRemove xóa thành viên hoặc tự rời nhóm
func (s *GroupService) LeaveOrRemove(ctx context.Context, requesterID, groupID, targetUserID string) error {
	reqOID, err := primitive.ObjectIDFromHex(requesterID)
	if err != nil {
		return errors.New("requester_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	targetOID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return errors.New("target_user_id không hợp lệ")
	}

	// Nếu tự rời nhóm
	if reqOID == targetOID {
		return s.groupRepo.RemoveMember(ctx, groupOID, targetOID)
	}

	// Nếu xóa người khác -> Phải là Admin
	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, reqOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền xóa thành viên khác")
	}

	return s.groupRepo.RemoveMember(ctx, groupOID, targetOID)
}

// UpdateSlowMode cập nhật chế độ chậm cho nhóm (chỉ Admin)
func (s *GroupService) UpdateSlowMode(ctx context.Context, adminID, groupID string, seconds int) error {
	adminOID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return errors.New("admin_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, adminOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền đổi chế độ chậm")
	}

	return s.groupRepo.UpdateSlowMode(ctx, groupOID, seconds)
}

// CreateCategory tạo danh mục kênh mới
func (s *GroupService) CreateCategory(ctx context.Context, userID, groupID, name string) (*models.ChannelCategory, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, userOID)
	if err != nil || !isAdmin {
		return nil, errors.New("chỉ Quản trị viên mới có quyền tạo danh mục kênh")
	}

	cat := &models.ChannelCategory{
		ID:        primitive.NewObjectID().Hex(),
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.groupRepo.AddCategory(ctx, groupOID, *cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// DeleteCategory xóa danh mục kênh
func (s *GroupService) DeleteCategory(ctx context.Context, userID, groupID, catID string) error {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, userOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền xóa danh mục kênh")
	}

	return s.groupRepo.DeleteCategory(ctx, groupOID, catID)
}

// CreateChannel tạo kênh chat con mới trong nhóm
func (s *GroupService) CreateChannel(ctx context.Context, userID, groupID string, req models.CreateChannelRequest) (*models.Channel, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, userOID)
	if err != nil || !isAdmin {
		return nil, errors.New("chỉ Quản trị viên mới có quyền tạo kênh")
	}

	chType := req.Type
	if chType != models.ChannelTypeAnnouncement {
		chType = models.ChannelTypeText
	}

	chanObj := &models.Channel{
		ID:          primitive.NewObjectID().Hex(),
		Name:        req.Name,
		Description: req.Description,
		Type:        chType,
		CategoryID:  req.CategoryID,
		CreatedAt:   time.Now(),
	}

	if err := s.groupRepo.AddChannel(ctx, groupOID, *chanObj); err != nil {
		return nil, err
	}
	return chanObj, nil
}

// DeleteChannel xóa kênh khỏi nhóm
func (s *GroupService) DeleteChannel(ctx context.Context, userID, groupID, chanID string) error {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, userOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền xóa kênh")
	}

	return s.groupRepo.DeleteChannel(ctx, groupOID, chanID)
}

// UpdateMemberRole cập nhật vai trò ("admin", "moderator", "member")
func (s *GroupService) UpdateMemberRole(ctx context.Context, operatorID, groupID, targetUserID, newRole string) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	targetOID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return errors.New("target_user_id không hợp lệ")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return err
	}

	if targetOID == group.CreatorID {
		return errors.New("không thể thay đổi vai trò của Chủ phòng (Owner)")
	}

	isCreator := (opOID == group.CreatorID)
	isAdmin, _ := s.groupRepo.IsAdmin(ctx, groupOID, opOID)

	if !isCreator && !isAdmin {
		return errors.New("bạn không có quyền phân bổ vai trò trong nhóm")
	}

	// Chỉ Chủ phòng mới có quyền thăng cấp thành Admin
	if newRole == "admin" && !isCreator {
		return errors.New("chỉ Chủ phòng (Owner) mới có quyền chỉ định Quản trị viên (Admin)")
	}

	return s.groupRepo.UpdateMemberRole(ctx, groupOID, targetOID, newRole)
}

// MuteMember cấm chat thành viên trong số phút chỉ định
func (s *GroupService) MuteMember(ctx context.Context, operatorID, groupID, targetUserID string, durationMinutes int) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	targetOID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return errors.New("target_user_id không hợp lệ")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return err
	}

	if targetOID == group.CreatorID {
		return errors.New("không thể cấm chat Chủ phòng (Owner)")
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, opOID)
	if !canManage {
		return errors.New("bạn không có quyền cấm chat thành viên")
	}

	// Moderator không được cấm Admin/Creator
	isOpAdmin, _ := s.groupRepo.IsAdmin(ctx, groupOID, opOID)
	isTargetAdmin, _ := s.groupRepo.IsAdmin(ctx, groupOID, targetOID)
	if !isOpAdmin && isTargetAdmin {
		return errors.New("Điều hành viên không thể cấm chat Quản trị viên")
	}

	return s.groupRepo.MuteMember(ctx, groupOID, targetOID, durationMinutes)
}

// UpdateSettings cập nhật cài đặt duyệt thành viên và chế độ chậm
func (s *GroupService) UpdateSettings(ctx context.Context, operatorID, groupID string, req models.UpdateGroupSettingsRequest) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	isAdmin, err := s.groupRepo.IsAdmin(ctx, groupOID, opOID)
	if err != nil || !isAdmin {
		return errors.New("chỉ Quản trị viên mới có quyền đổi cài đặt nhóm")
	}

	return s.groupRepo.UpdateSettings(ctx, groupOID, req.RequireApproval, req.SlowModeSeconds)
}

// CreateInvite sinh liên kết mời vào nhóm mới
func (s *GroupService) CreateInvite(ctx context.Context, operatorID, groupID string, req models.CreateInviteRequest) (*models.GroupInvite, error) {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return nil, errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}

	isMember, err := s.groupRepo.IsMember(ctx, groupOID, opOID)
	if err != nil || !isMember {
		return nil, errors.New("chỉ thành viên mới có thể tạo liên kết mời")
	}

	code := fmt.Sprintf("inv_%s", primitive.NewObjectID().Hex()[:10])
	var expiresAt *time.Time
	if req.ExpireHours > 0 {
		exp := time.Now().Add(time.Duration(req.ExpireHours) * time.Hour)
		expiresAt = &exp
	}

	invite := &models.GroupInvite{
		Code:      code,
		GroupID:   groupOID,
		CreatedBy: opOID,
		MaxUses:   req.MaxUses,
		UsesCount: 0,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}
	return invite, nil
}

// GetGroupInvites lấy danh sách liên kết mời của nhóm
func (s *GroupService) GetGroupInvites(ctx context.Context, operatorID, groupID string) ([]models.GroupInvite, error) {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return nil, errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}

	isMember, err := s.groupRepo.IsMember(ctx, groupOID, opOID)
	if err != nil || !isMember {
		return nil, errors.New("không có quyền xem liên kết mời")
	}

	return s.inviteRepo.GetActiveByGroup(ctx, groupOID)
}

// RevokeInvite thu hồi / xóa liên kết mời
func (s *GroupService) RevokeInvite(ctx context.Context, operatorID, groupID, code string) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, opOID)
	inv, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		return errors.New("liên kết mời không tồn tại")
	}

	if !canManage && inv.CreatedBy != opOID {
		return errors.New("bạn không có quyền thu hồi liên kết này")
	}

	return s.inviteRepo.DeleteByCode(ctx, code)
}

// PreviewInvite xem trước thông tin nhóm khi mở link mời
func (s *GroupService) PreviewInvite(ctx context.Context, currentUserID, code string) (*models.InvitePreviewResponse, error) {
	inv, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil || inv == nil {
		return nil, errors.New("liên kết mời không hợp lệ hoặc đã bị thu hồi")
	}

	group, err := s.groupRepo.GetByID(ctx, inv.GroupID)
	if err != nil || group == nil {
		return nil, errors.New("nhóm không tồn tại hoặc đã bị giải tán")
	}

	now := time.Now()
	isExpired := inv.ExpiresAt != nil && inv.ExpiresAt.Before(now)
	isMaxedOut := inv.MaxUses > 0 && inv.UsesCount >= inv.MaxUses

	inviterName := "Một thành viên"
	if inviter, err := s.userRepo.FindByID(ctx, inv.CreatedBy.Hex()); err == nil && inviter != nil {
		if inviter.DisplayName != "" {
			inviterName = inviter.DisplayName
		} else {
			inviterName = inviter.Username
		}
	}

	userOID, _ := primitive.ObjectIDFromHex(currentUserID)
	isAlreadyMember, _ := s.groupRepo.IsMember(ctx, inv.GroupID, userOID)
	hasPendingReq := false
	if s.joinReqRepo != nil {
		hasPendingReq, _ = s.joinReqRepo.HasPending(ctx, inv.GroupID, userOID)
	}

	return &models.InvitePreviewResponse{
		Code:            inv.Code,
		GroupID:         group.ID.Hex(),
		GroupName:       group.Name,
		GroupAvatar:     group.Avatar,
		MemberCount:     len(group.MemberIDs),
		RequireApproval: group.RequireApproval,
		InviterName:     inviterName,
		ExpiresAt:       inv.ExpiresAt,
		IsExpired:       isExpired,
		IsMaxedOut:      isMaxedOut,
		IsAlreadyMember: isAlreadyMember,
		HasPendingReq:   hasPendingReq,
	}, nil
}

// JoinViaInvite tham gia nhóm qua mã mời (hoặc gửi đơn chờ duyệt nếu nhóm yêu cầu duyệt)
func (s *GroupService) JoinViaInvite(ctx context.Context, currentUserID, code string) (bool, bool, error) {
	inv, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil || inv == nil {
		return false, false, errors.New("liên kết mời không hợp lệ hoặc đã bị thu hồi")
	}

	now := time.Now()
	if inv.ExpiresAt != nil && inv.ExpiresAt.Before(now) {
		return false, false, errors.New("liên kết mời này đã hết hạn sử dụng")
	}
	if inv.MaxUses > 0 && inv.UsesCount >= inv.MaxUses {
		return false, false, errors.New("liên kết mời này đã đạt giới hạn số lượt tham gia")
	}

	group, err := s.groupRepo.GetByID(ctx, inv.GroupID)
	if err != nil || group == nil {
		return false, false, errors.New("nhóm không tồn tại")
	}

	userOID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return false, false, errors.New("user_id không hợp lệ")
	}

	isMember, _ := s.groupRepo.IsMember(ctx, group.ID, userOID)
	if isMember {
		return true, false, nil // Đã là thành viên
	}

	// Nếu nhóm yêu cầu phê duyệt
	if group.RequireApproval {
		hasPending, _ := s.joinReqRepo.HasPending(ctx, group.ID, userOID)
		if hasPending {
			return false, true, nil
		}

		req := &models.GroupJoinRequest{
			GroupID:   group.ID,
			UserID:    userOID,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		if err := s.joinReqRepo.Create(ctx, req); err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	// Gia nhập trực tiếp
	if err := s.groupRepo.AddMembers(ctx, group.ID, []primitive.ObjectID{userOID}); err != nil {
		return false, false, err
	}
	_ = s.inviteRepo.IncrementUses(ctx, code)
	return true, false, nil
}

// GetPendingJoinRequests lấy danh sách đơn chờ duyệt
func (s *GroupService) GetPendingJoinRequests(ctx context.Context, operatorID, groupID string) ([]models.JoinRequestDetail, error) {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return nil, errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, opOID)
	if !canManage {
		return nil, errors.New("bạn không có quyền xem đơn xin gia nhập nhóm")
	}

	requests, err := s.joinReqRepo.GetPendingByGroup(ctx, groupOID)
	if err != nil {
		return nil, err
	}

	userOIDs := make([]primitive.ObjectID, 0, len(requests))
	for _, r := range requests {
		userOIDs = append(userOIDs, r.UserID)
	}

	userList, _ := s.userRepo.FindByIDs(ctx, userOIDs)
	userMap := make(map[string]models.GroupMemberInfo)
	for _, u := range userList {
		userMap[u.ID.Hex()] = models.GroupMemberInfo{
			ID:          u.ID.Hex(),
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
		}
	}

	details := make([]models.JoinRequestDetail, 0, len(requests))
	for _, r := range requests {
		details = append(details, models.JoinRequestDetail{
			ID:        r.ID.Hex(),
			GroupID:   groupID,
			User:      userMap[r.UserID.Hex()],
			Status:    r.Status,
			Note:      r.Note,
			CreatedAt: r.CreatedAt,
		})
	}
	return details, nil
}

// ApproveJoinRequest duyệt cho thành viên vào nhóm
func (s *GroupService) ApproveJoinRequest(ctx context.Context, operatorID, groupID, requestID string) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	reqOID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("request_id không hợp lệ")
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, opOID)
	if !canManage {
		return errors.New("bạn không có quyền phê duyệt đơn gia nhập")
	}

	req, err := s.joinReqRepo.GetByID(ctx, reqOID)
	if err != nil || req == nil {
		return errors.New("yêu cầu không tồn tại")
	}

	// Thêm thành viên vào nhóm
	if err := s.groupRepo.AddMembers(ctx, groupOID, []primitive.ObjectID{req.UserID}); err != nil {
		return err
	}

	// Cập nhật trạng thái duyệt
	return s.joinReqRepo.UpdateStatus(ctx, reqOID, "approved", opOID)
}

// RejectJoinRequest từ chối đơn vào nhóm
func (s *GroupService) RejectJoinRequest(ctx context.Context, operatorID, groupID, requestID string) error {
	opOID, err := primitive.ObjectIDFromHex(operatorID)
	if err != nil {
		return errors.New("operator_id không hợp lệ")
	}
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	reqOID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("request_id không hợp lệ")
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, opOID)
	if !canManage {
		return errors.New("bạn không có quyền từ chối đơn gia nhập")
	}

	return s.joinReqRepo.UpdateStatus(ctx, reqOID, "rejected", opOID)
}

// broadcastGroup phát sóng sự kiện qua Redis Pub/Sub đến tất cả thành viên trong nhóm
func (s *GroupService) broadcastGroup(groupID string, eventName string, payload interface{}) {
	if s.redisClient == nil {
		return
	}
	event := models.WSEvent{
		Event:   eventName,
		Payload: payload,
	}
	bytes, err := json.Marshal(event)
	if err == nil {
		s.redisClient.Publish(context.Background(), fmt.Sprintf("group:chat:%s", groupID), string(bytes))
	}
}

// ============================================================================
// PHASE 3: POLLS & QUIZZES
// ============================================================================

// CreatePoll tạo một cuộc bình chọn mới trong nhóm/kênh
func (s *GroupService) CreatePoll(ctx context.Context, userID, groupID string, req models.CreatePollRequest) (*models.Poll, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return nil, errors.New("nhóm không tồn tại")
	}

	// Kiểm tra xem người dùng có trong nhóm không
	isMember := false
	for _, m := range group.MemberIDs {
		if m == userOID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("bạn không phải thành viên của nhóm")
	}

	// Kiểm tra cấm chat
	if group.MutedMembers != nil {
		if mutedUntil, ok := group.MutedMembers[userID]; ok && mutedUntil.After(time.Now()) {
			return nil, errors.New("bạn đang bị cấm gửi tin trong nhóm này")
		}
	}

	channelID := req.ChannelID
	if channelID == "" && len(group.Channels) > 0 {
		channelID = group.Channels[0].ID
	}
	for _, ch := range group.Channels {
		if ch.ID == channelID && ch.Type == "announcement" {
			canPost, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, userOID)
			if !canPost {
				return nil, errors.New("kênh thông báo chỉ dành cho quản trị viên đăng bài")
			}
		}
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	creatorName := "Thành viên"
	creatorAvatar := ""
	if user != nil {
		creatorName = user.DisplayName
		if creatorName == "" {
			creatorName = user.Username
		}
		creatorAvatar = user.AvatarURL
	}

	options := make([]models.PollOption, 0, len(req.Options))
	for i, optText := range req.Options {
		options = append(options, models.PollOption{
			ID:        fmt.Sprintf("opt_%d", i+1),
			Text:      optText,
			VoterIDs:  []string{},
			VoteCount: 0,
		})
	}

	poll := &models.Poll{
		GroupID:        groupOID,
		ChannelID:      channelID,
		Question:       req.Question,
		Options:        options,
		MultipleChoice: req.MultipleChoice,
		IsAnonymous:    req.IsAnonymous,
		IsClosed:       false,
		TotalVotes:     0,
		CreatedBy:      userOID,
		CreatorName:    creatorName,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	// Tạo Message đại diện trong kênh chat
	msg := &models.Message{
		GroupID:      groupID,
		ChannelID:    channelID,
		SenderID:     userID,
		SenderName:   creatorName,
		SenderAvatar: creatorAvatar,
		Content:      req.Question,
		Type:         "poll",
		PollID:       poll.ID.Hex(),
		Poll:         poll,
		CreatedAt:    time.Now(),
	}
	if err := s.messageRepo.Create(ctx, msg); err == nil {
		s.pollRepo.UpdateMessageID(ctx, poll.ID, msg.ID.Hex())
		s.broadcastGroup(groupID, "group:receive", msg)
	}

	return poll, nil
}

// VotePoll thực hiện bỏ phiếu trong bình chọn
func (s *GroupService) VotePoll(ctx context.Context, userID, groupID, pollID, optionID string) (*models.Poll, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	pollOID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		return nil, errors.New("poll_id không hợp lệ")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return nil, errors.New("nhóm không tồn tại")
	}

	isMember := false
	for _, m := range group.MemberIDs {
		if m == userOID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("bạn không phải thành viên của nhóm")
	}

	// Kiểm tra cấm chat
	if group.MutedMembers != nil {
		if mutedUntil, ok := group.MutedMembers[userID]; ok && mutedUntil.After(time.Now()) {
			return nil, errors.New("bạn đang bị cấm tương tác trong nhóm này")
		}
	}

	updatedPoll, err := s.pollRepo.Vote(ctx, pollOID, userID, optionID)
	if err != nil {
		return nil, err
	}

	s.broadcastGroup(groupID, "group:poll_updated", updatedPoll)
	return updatedPoll, nil
}

// ClosePoll đóng cuộc bình chọn
func (s *GroupService) ClosePoll(ctx context.Context, userID, groupID, pollID string) (*models.Poll, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	pollOID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		return nil, errors.New("poll_id không hợp lệ")
	}

	poll, err := s.pollRepo.GetByID(ctx, pollOID)
	if err != nil {
		return nil, err
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, userOID)
	if poll.CreatedBy != userOID && !canManage {
		return nil, errors.New("bạn không có quyền đóng cuộc bình chọn này")
	}

	closedPoll, err := s.pollRepo.Close(ctx, pollOID)
	if err != nil {
		return nil, err
	}

	s.broadcastGroup(groupID, "group:poll_updated", closedPoll)
	return closedPoll, nil
}

// GetPoll lấy thông tin chi tiết cuộc bình chọn
func (s *GroupService) GetPoll(ctx context.Context, userID, groupID, pollID string) (*models.Poll, error) {
	pollOID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		return nil, errors.New("poll_id không hợp lệ")
	}
	return s.pollRepo.GetByID(ctx, pollOID)
}

// ============================================================================
// PHASE 3: GROUP EVENTS
// ============================================================================

// CreateEvent tạo sự kiện mới trong nhóm
func (s *GroupService) CreateEvent(ctx context.Context, userID, groupID string, req models.CreateEventRequest) (*models.GroupEvent, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}

	group, err := s.groupRepo.GetByID(ctx, groupOID)
	if err != nil {
		return nil, errors.New("nhóm không tồn tại")
	}

	isMember := false
	for _, m := range group.MemberIDs {
		if m == userOID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("bạn không phải thành viên của nhóm")
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	creatorName := "Thành viên"
	creatorAvatar := ""
	if user != nil {
		creatorName = user.DisplayName
		if creatorName == "" {
			creatorName = user.Username
		}
		creatorAvatar = user.AvatarURL
	}

	channelID := req.ChannelID
	if channelID == "" && len(group.Channels) > 0 {
		channelID = group.Channels[0].ID
	}

	event := &models.GroupEvent{
		GroupID:       groupOID,
		ChannelID:     channelID,
		Title:         req.Title,
		Description:   req.Description,
		Location:      req.Location,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		CreatedBy:     userOID,
		CreatorName:   creatorName,
		CreatorAvatar: creatorAvatar,
		Attendees: []models.EventAttendee{
			{
				UserID:     userID,
				UserName:   creatorName,
				UserAvatar: creatorAvatar,
				Status:     "going",
				UpdatedAt:  time.Now(),
			},
		},
		Status: "upcoming",
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	// Đăng một thẻ tin nhắn sự kiện vào kênh chat
	msg := &models.Message{
		GroupID:      groupID,
		ChannelID:    channelID,
		SenderID:     userID,
		SenderName:   creatorName,
		SenderAvatar: creatorAvatar,
		Content:      fmt.Sprintf("📅 Sự kiện mới: %s", req.Title),
		Type:         "event",
		EventID:      event.ID.Hex(),
		Event:        event,
		CreatedAt:    time.Now(),
	}
	if err := s.messageRepo.Create(ctx, msg); err == nil {
		s.eventRepo.UpdateMessageID(ctx, event.ID, msg.ID.Hex())
		s.broadcastGroup(groupID, "group:receive", msg)
	}

	s.broadcastGroup(groupID, "group:event_created", event)
	return event, nil
}

// GetGroupEvents lấy danh sách sự kiện sắp tới của nhóm
func (s *GroupService) GetGroupEvents(ctx context.Context, userID, groupID string) ([]models.GroupEvent, error) {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("group_id không hợp lệ")
	}
	return s.eventRepo.GetUpcomingByGroup(ctx, groupOID)
}

// RSVPEvent phản hồi tham gia sự kiện
func (s *GroupService) RSVPEvent(ctx context.Context, userID, groupID, eventID, status string) (*models.GroupEvent, error) {
	if _, err := primitive.ObjectIDFromHex(userID); err != nil {
		return nil, errors.New("user_id không hợp lệ")
	}
	eventOID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, errors.New("event_id không hợp lệ")
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	userName := userID
	userAvatar := ""
	if user != nil {
		userName = user.DisplayName
		if userName == "" {
			userName = user.Username
		}
		userAvatar = user.AvatarURL
	}

	attendee := models.EventAttendee{
		UserID:     userID,
		UserName:   userName,
		UserAvatar: userAvatar,
		Status:     status,
		UpdatedAt:  time.Now(),
	}

	updatedEvent, err := s.eventRepo.RSVP(ctx, eventOID, attendee)
	if err != nil {
		return nil, err
	}

	s.broadcastGroup(groupID, "group:event_updated", updatedEvent)
	return updatedEvent, nil
}

// DeleteEvent hủy sự kiện
func (s *GroupService) DeleteEvent(ctx context.Context, userID, groupID, eventID string) error {
	groupOID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return errors.New("group_id không hợp lệ")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user_id không hợp lệ")
	}
	eventOID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return errors.New("event_id không hợp lệ")
	}

	event, err := s.eventRepo.GetByID(ctx, eventOID)
	if err != nil {
		return err
	}

	canManage, _ := s.groupRepo.IsModeratorOrAdmin(ctx, groupOID, userOID)
	if event.CreatedBy != userOID && !canManage {
		return errors.New("bạn không có quyền xóa sự kiện này")
	}

	if err := s.eventRepo.Delete(ctx, eventOID); err != nil {
		return err
	}

	s.broadcastGroup(groupID, "group:event_deleted", map[string]string{"event_id": eventID})
	return nil
}
