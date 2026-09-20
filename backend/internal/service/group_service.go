package service

import (
	"context"
	"errors"
	"fmt"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupService struct {
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
}

func NewGroupService(groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
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

	memberInfos := make([]models.GroupMemberInfo, 0, len(users))
	for _, u := range users {
		memberInfos = append(memberInfos, models.GroupMemberInfo{
			ID:          u.ID.Hex(),
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			IsAdmin:     adminMap[u.ID.Hex()],
		})
	}

	res := &models.GroupDetailResponse{
		ID:        group.ID.Hex(),
		Name:      group.Name,
		Avatar:    group.Avatar,
		CreatorID: group.CreatorID.Hex(),
		Members:   memberInfos,
		CreatedAt: group.CreatedAt,
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
