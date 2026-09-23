package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupService struct {
	groupRepo   *repository.GroupRepository
	userRepo    *repository.UserRepository
	inviteRepo  *repository.InviteRepository
	joinReqRepo *repository.JoinRequestRepository
}

func NewGroupService(
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
	inviteRepo *repository.InviteRepository,
	joinReqRepo *repository.JoinRequestRepository,
) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		inviteRepo:  inviteRepo,
		joinReqRepo: joinReqRepo,
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



