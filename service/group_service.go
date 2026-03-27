package service

import (
	"errors"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validGroupReactionTypes = map[string]bool{
	"like": true, "love": true, "haha": true,
	"wow": true, "sad": true, "angry": true,
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type GroupRequest struct {
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Rules       *string `json:"rules"`
	BannerURL   *string `json:"banner_url"`
}

type GroupArticleRequest struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	MediaURL *string `json:"media_url"`
}

type GroupCommentRequest struct {
	Content         string `json:"content"`
	ParentCommentID *uint  `json:"parent_comment_id"`
}

type GroupReactionRequest struct {
	Type string `json:"type"`
}

// ─── Group Comment Node (recursive) ──────────────────────────────────────────

type GroupCommentNode struct {
	ID              uint                   `json:"id"`
	ArticleID       uint                   `json:"article_id"`
	UserID          uint                   `json:"user_id"`
	User            models.User            `json:"user"`
	Content         string                 `json:"content"`
	ParentCommentID *uint                  `json:"parent_comment_id"`
	Reactions       []models.GroupReaction `json:"reactions"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Replies         []*GroupCommentNode    `json:"replies"`
}

type GroupArticleDetailResponse struct {
	ID        uint                   `json:"id"`
	GroupID   uint                   `json:"group_id"`
	UserID    uint                   `json:"user_id"`
	User      models.User            `json:"user"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	MediaURL  *string                `json:"media_url"`
	Reactions []models.GroupReaction `json:"reactions"`
	Comments  []*GroupCommentNode    `json:"comments"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func buildGroupCommentTree(comments []models.GroupComment) []*GroupCommentNode {
	nodeMap := make(map[uint]*GroupCommentNode, len(comments))
	for i := range comments {
		c := &comments[i]
		nodeMap[c.ID] = &GroupCommentNode{
			ID:              c.ID,
			ArticleID:       c.ArticleID,
			UserID:          c.UserID,
			User:            c.User,
			Content:         c.Content,
			ParentCommentID: c.ParentCommentID,
			Reactions:       c.Reactions,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
			Replies:         []*GroupCommentNode{},
		}
	}
	var roots []*GroupCommentNode
	for i := range comments {
		c := &comments[i]
		node := nodeMap[c.ID]
		if c.ParentCommentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*c.ParentCommentID]; ok {
				parent.Replies = append(parent.Replies, node)
			} else {
				roots = append(roots, node)
			}
		}
	}
	if roots == nil {
		roots = []*GroupCommentNode{}
	}
	return roots
}

// ─── Interface ────────────────────────────────────────────────────────────────

type GroupService interface {
	CreateGroup(userID uint, req GroupRequest) (*models.Group, error)
	GetGroups() ([]models.Group, error)
	GetGroupByID(id uint) (*models.Group, error)
	UpdateGroup(userID uint, groupID uint, req GroupRequest) (*models.Group, error)
	DeleteGroup(userID uint, groupID uint) error

	JoinGroup(userID uint, groupID uint) error
	LeaveGroup(userID uint, groupID uint) error
	GetGroupMembers(groupID uint) ([]models.GroupMember, error)
	GetJoinedGroups(userID uint) ([]models.Group, error)
	KickMember(ownerID uint, groupID uint, targetUserID uint) error

	CreateGroupArticle(userID uint, groupID uint, req GroupArticleRequest) (*models.GroupArticle, error)
	GetGroupArticleDetail(userID uint, articleID uint) (*GroupArticleDetailResponse, error)
	UpdateGroupArticle(userID uint, articleID uint, req GroupArticleRequest) (*models.GroupArticle, error)
	DeleteGroupArticle(userID uint, articleID uint) error

	AddGroupComment(userID uint, articleID uint, req GroupCommentRequest) (*models.GroupComment, error)
	UpdateGroupComment(userID uint, commentID uint, req GroupCommentRequest) (*models.GroupComment, error)
	DeleteGroupComment(userID uint, commentID uint) error

	ReactToGroupArticle(userID uint, articleID uint, req GroupReactionRequest) (string, error)
	ReactToGroupComment(userID uint, commentID uint, req GroupReactionRequest) (string, error)
}

// ─── Struct ───────────────────────────────────────────────────────────────────

type groupService struct {
	groupRepo         repository.GroupRepository
	memberRepo        repository.GroupMemberRepository
	articleRepo       repository.GroupArticleRepository
	commentRepo       repository.GroupCommentRepository
	groupReactionRepo repository.GroupReactionRepository
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	memberRepo repository.GroupMemberRepository,
	articleRepo repository.GroupArticleRepository,
	commentRepo repository.GroupCommentRepository,
	groupReactionRepo repository.GroupReactionRepository,
) GroupService {
	return &groupService{
		groupRepo:         groupRepo,
		memberRepo:        memberRepo,
		articleRepo:       articleRepo,
		commentRepo:       commentRepo,
		groupReactionRepo: groupReactionRepo,
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *groupService) isMember(groupID, userID uint) bool {
	_, err := s.memberRepo.FindGroupMember(groupID, userID)
	return err == nil
}

func (s *groupService) isOwner(groupID, userID uint) bool {
	m, err := s.memberRepo.FindGroupMember(groupID, userID)
	return err == nil && m.Role == "owner"
}

// ─── Group CRUD ───────────────────────────────────────────────────────────────

func (s *groupService) CreateGroup(userID uint, req GroupRequest) (*models.Group, error) {
	if req.Category == "" || req.Title == "" {
		return nil, errors.New("kategori dan judul wajib diisi")
	}
	group := &models.Group{
		OwnerID:     userID,
		Category:    req.Category,
		Title:       req.Title,
		Description: req.Description,
		Rules:       req.Rules,
		BannerURL:   req.BannerURL,
	}
	if err := s.groupRepo.CreateGroup(group); err != nil {
		return nil, errors.New("gagal membuat grup")
	}
	member := &models.GroupMember{GroupID: group.ID, UserID: userID, Role: "owner"}
	if err := s.memberRepo.AddGroupMember(member); err != nil {
		return nil, errors.New("gagal menambahkan pemilik sebagai anggota grup")
	}
	return group, nil
}

func (s *groupService) GetGroups() ([]models.Group, error) {
	return s.groupRepo.FindGroups()
}

func (s *groupService) GetGroupByID(id uint) (*models.Group, error) {
	group, err := s.groupRepo.FindGroupByID(id)
	if err != nil {
		return nil, errors.New("grup tidak ditemukan")
	}
	return group, nil
}

func (s *groupService) UpdateGroup(userID uint, groupID uint, req GroupRequest) (*models.Group, error) {
	if !s.isOwner(groupID, userID) {
		return nil, errors.New("akses ditolak: hanya pemilik grup")
	}
	group, err := s.groupRepo.FindGroupByID(groupID)
	if err != nil {
		return nil, errors.New("grup tidak ditemukan")
	}
	if req.Category != "" {
		group.Category = req.Category
	}
	if req.Title != "" {
		group.Title = req.Title
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	if req.Rules != nil {
		group.Rules = req.Rules
	}
	if req.BannerURL != nil {
		group.BannerURL = req.BannerURL
	}
	if err := s.groupRepo.UpdateGroup(group); err != nil {
		return nil, errors.New("gagal memperbarui grup")
	}
	return group, nil
}

func (s *groupService) DeleteGroup(userID uint, groupID uint) error {
	if !s.isOwner(groupID, userID) {
		return errors.New("akses ditolak: hanya pemilik grup")
	}
	return s.groupRepo.DeleteGroup(groupID)
}

// ─── Membership ───────────────────────────────────────────────────────────────

func (s *groupService) JoinGroup(userID uint, groupID uint) error {
	if _, err := s.groupRepo.FindGroupByID(groupID); err != nil {
		return errors.New("grup tidak ditemukan")
	}
	if s.isMember(groupID, userID) {
		return errors.New("sudah menjadi anggota grup ini")
	}
	return s.memberRepo.AddGroupMember(&models.GroupMember{GroupID: groupID, UserID: userID, Role: "member"})
}

func (s *groupService) LeaveGroup(userID uint, groupID uint) error {
	if s.isOwner(groupID, userID) {
		return errors.New("pemilik grup tidak dapat meninggalkan grup")
	}
	if !s.isMember(groupID, userID) {
		return errors.New("Anda bukan anggota grup ini")
	}
	return s.memberRepo.RemoveGroupMember(groupID, userID)
}

func (s *groupService) GetGroupMembers(groupID uint) ([]models.GroupMember, error) {
	return s.memberRepo.FindGroupMembers(groupID)
}

func (s *groupService) GetJoinedGroups(userID uint) ([]models.Group, error) {
	return s.memberRepo.FindJoinedGroups(userID)
}

func (s *groupService) KickMember(ownerID uint, groupID uint, targetUserID uint) error {
	if !s.isOwner(groupID, ownerID) {
		return errors.New("akses ditolak: hanya pemilik grup")
	}
	if ownerID == targetUserID {
		return errors.New("pemilik tidak dapat mengeluarkan dirinya sendiri")
	}
	if !s.isMember(groupID, targetUserID) {
		return errors.New("pengguna yang dituju bukan anggota grup")
	}
	return s.memberRepo.RemoveGroupMember(groupID, targetUserID)
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func (s *groupService) CreateGroupArticle(userID uint, groupID uint, req GroupArticleRequest) (*models.GroupArticle, error) {
	if !s.isMember(groupID, userID) {
		return nil, errors.New("akses ditolak: hanya anggota grup")
	}
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("judul dan konten wajib diisi")
	}
	article := &models.GroupArticle{
		GroupID:  groupID,
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		MediaURL: req.MediaURL,
	}
	if err := s.articleRepo.CreateGroupArticle(article); err != nil {
		return nil, errors.New("gagal membuat artikel")
	}
	return article, nil
}

func (s *groupService) GetGroupArticleDetail(userID uint, articleID uint) (*GroupArticleDetailResponse, error) {
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("artikel tidak ditemukan")
	}
	comments, _ := s.articleRepo.FindAllCommentsByArticleID(articleID)
	if !s.isMember(article.GroupID, userID) {
		comments = []models.GroupComment{}
	}
	return &GroupArticleDetailResponse{
		ID:        article.ID,
		GroupID:   article.GroupID,
		UserID:    article.UserID,
		User:      article.User,
		Title:     article.Title,
		Content:   article.Content,
		MediaURL:  article.MediaURL,
		Reactions: article.Reactions,
		Comments:  buildGroupCommentTree(comments),
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}, nil
}

func (s *groupService) UpdateGroupArticle(userID uint, articleID uint, req GroupArticleRequest) (*models.GroupArticle, error) {
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("artikel tidak ditemukan")
	}
	if article.UserID != userID && !s.isOwner(article.GroupID, userID) {
		return nil, errors.New("akses ditolak")
	}
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.MediaURL != nil {
		article.MediaURL = req.MediaURL
	}
	if err := s.articleRepo.UpdateGroupArticle(article); err != nil {
		return nil, errors.New("gagal memperbarui artikel")
	}
	return article, nil
}

func (s *groupService) DeleteGroupArticle(userID uint, articleID uint) error {
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return errors.New("artikel tidak ditemukan")
	}
	if article.UserID != userID && !s.isOwner(article.GroupID, userID) {
		return errors.New("akses ditolak")
	}
	return s.articleRepo.DeleteGroupArticle(articleID)
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func (s *groupService) AddGroupComment(userID uint, articleID uint, req GroupCommentRequest) (*models.GroupComment, error) {
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("artikel tidak ditemukan")
	}
	if !s.isMember(article.GroupID, userID) {
		return nil, errors.New("akses ditolak: hanya anggota grup")
	}
	if req.Content == "" {
		return nil, errors.New("konten wajib diisi")
	}
	comment := &models.GroupComment{
		ArticleID:       articleID,
		UserID:          userID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
	}
	if err := s.commentRepo.CreateGroupComment(comment); err != nil {
		return nil, errors.New("gagal menambahkan komentar")
	}
	return comment, nil
}

func (s *groupService) UpdateGroupComment(userID uint, commentID uint, req GroupCommentRequest) (*models.GroupComment, error) {
	if req.Content == "" {
		return nil, errors.New("konten wajib diisi")
	}
	comment, err := s.commentRepo.FindGroupCommentByID(commentID)
	if err != nil {
		return nil, errors.New("komentar tidak ditemukan")
	}
	if comment.UserID != userID {
		return nil, errors.New("akses ditolak")
	}
	comment.Content = req.Content
	if err := s.commentRepo.UpdateGroupComment(comment); err != nil {
		return nil, errors.New("gagal memperbarui komentar")
	}
	return comment, nil
}

func (s *groupService) DeleteGroupComment(userID uint, commentID uint) error {
	comment, err := s.commentRepo.FindGroupCommentByID(commentID)
	if err != nil {
		return errors.New("komentar tidak ditemukan")
	}
	article, _ := s.articleRepo.FindGroupArticleByID(comment.ArticleID)
	if comment.UserID != userID && (article == nil || !s.isOwner(article.GroupID, userID)) {
		return errors.New("akses ditolak")
	}
	return s.commentRepo.DeleteGroupComment(commentID)
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func (s *groupService) ReactToGroupArticle(userID uint, articleID uint, req GroupReactionRequest) (string, error) {
	if !validGroupReactionTypes[req.Type] {
		return "", errors.New("jenis reaksi tidak valid")
	}
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return "", errors.New("artikel tidak ditemukan")
	}
	if !s.isMember(article.GroupID, userID) {
		return "", errors.New("akses ditolak: hanya anggota grup")
	}
	existing, err := s.groupReactionRepo.FindGroupReaction(userID, &articleID, nil)
	if err == nil {
		if existing.Type == req.Type {
			if err := s.groupReactionRepo.DeleteGroupReaction(existing.ID); err != nil {
				return "", errors.New("gagal menghapus reaksi")
			}
			return "removed", nil
		}
		existing.Type = req.Type
		if err := s.groupReactionRepo.UpdateGroupReaction(existing); err != nil {
			return "", errors.New("gagal memperbarui reaksi")
		}
		return "updated", nil
	}
	if err := s.groupReactionRepo.CreateGroupReaction(&models.GroupReaction{UserID: userID, ArticleID: &articleID, Type: req.Type}); err != nil {
		return "", errors.New("gagal menambahkan reaksi")
	}
	return "added", nil
}

func (s *groupService) ReactToGroupComment(userID uint, commentID uint, req GroupReactionRequest) (string, error) {
	if !validGroupReactionTypes[req.Type] {
		return "", errors.New("jenis reaksi tidak valid")
	}
	comment, err := s.commentRepo.FindGroupCommentByID(commentID)
	if err != nil {
		return "", errors.New("komentar tidak ditemukan")
	}
	article, _ := s.articleRepo.FindGroupArticleByID(comment.ArticleID)
	if article != nil && !s.isMember(article.GroupID, userID) {
		return "", errors.New("akses ditolak: hanya anggota grup")
	}
	existing, err := s.groupReactionRepo.FindGroupReaction(userID, nil, &commentID)
	if err == nil {
		if existing.Type == req.Type {
			if err := s.groupReactionRepo.DeleteGroupReaction(existing.ID); err != nil {
				return "", errors.New("gagal menghapus reaksi")
			}
			return "removed", nil
		}
		existing.Type = req.Type
		if err := s.groupReactionRepo.UpdateGroupReaction(existing); err != nil {
			return "", errors.New("gagal memperbarui reaksi")
		}
		return "updated", nil
	}
	if err := s.groupReactionRepo.CreateGroupReaction(&models.GroupReaction{UserID: userID, CommentID: &commentID, Type: req.Type}); err != nil {
		return "", errors.New("gagal menambahkan reaksi")
	}
	return "added", nil
}
