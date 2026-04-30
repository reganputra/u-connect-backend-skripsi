package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
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
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	MediaURL  *string  `json:"media_url"`  // Deprecated: keeping for backward compatibility
	MediaURLs []string `json:"media_urls"` // New: multiple image URLs
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
	ID           uint                   `json:"id"`
	GroupID      uint                   `json:"group_id"`
	UserID       uint                   `json:"user_id"`
	User         models.User            `json:"user"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	MediaURL     *string                `json:"media_url"`  // Deprecated: first image only
	MediaURLs    []string               `json:"media_urls"` // New: all images
	CommentCount int                    `json:"comment_count"`
	Reactions    []models.GroupReaction `json:"reactions"`
	Comments     []*GroupCommentNode    `json:"comments"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type GroupListItemResponse struct {
	*models.Group
	MemberCount  int `json:"member_count"`
	ArticleCount int `json:"article_count"`
}

// PaginationMeta is included in responses with paginated sub-collections.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type GroupArticleWithCount struct {
	ID           uint                   `json:"ID"`
	CreatedAt    time.Time              `json:"CreatedAt"`
	UpdatedAt    time.Time              `json:"UpdatedAt"`
	DeletedAt    gorm.DeletedAt         `json:"DeletedAt"`
	GroupID      uint                   `json:"GroupID"`
	UserID       uint                   `json:"UserID"`
	Title        string                 `json:"Title"`
	Content      string                 `json:"Content"`
	MediaURL     *string                `json:"MediaURL"`   // Deprecated: first image only
	MediaURLs    []string               `json:"media_urls"` // New: all images
	User         models.User            `json:"User"`
	Comments     []models.GroupComment  `json:"Comments"`
	Reactions    []models.GroupReaction `json:"Reactions"`
	CommentCount int                    `json:"comment_count"`
}

type GroupDetailResponse struct {
	ID                uint                    `json:"ID"`
	CreatedAt         time.Time               `json:"CreatedAt"`
	UpdatedAt         time.Time               `json:"UpdatedAt"`
	DeletedAt         gorm.DeletedAt          `json:"DeletedAt"`
	OwnerID           uint                    `json:"OwnerID"`
	Category          string                  `json:"Category"`
	Title             string                  `json:"Title"`
	Description       *string                 `json:"Description"`
	Rules             *string                 `json:"Rules"`
	BannerURL         *string                 `json:"BannerURL"`
	Owner             models.User             `json:"Owner"`
	Members           []models.GroupMember    `json:"Members"`
	Articles          []GroupArticleWithCount `json:"Articles"`
	MemberCount       int                     `json:"member_count"`
	ArticleCount      int                     `json:"article_count"`
	ArticlePagination PaginationMeta          `json:"article_pagination"`
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
	for _, root := range roots {
		hydrateGroupCommentNode(root)
	}
	return roots
}

// ─── Interface ────────────────────────────────────────────────────────────────

type GroupService interface {
	CreateGroup(userID uint, req GroupRequest) (*models.Group, error)
	GetGroups(page, limit int) ([]*GroupListItemResponse, int64, error)
	// GetGroupByID returns the group detail with paginated articles.
	// articlePage/articleLimit control which page of articles to return (default 1/10).
	GetGroupByID(id uint, articlePage, articleLimit int) (*GroupDetailResponse, error)
	GetOwnedGroups(userID uint, page, limit int) ([]models.Group, int64, error)
	UpdateGroup(userID uint, groupID uint, req GroupRequest) (*models.Group, error)
	DeleteGroup(userID uint, groupID uint) error

	JoinGroup(userID uint, groupID uint) error
	LeaveGroup(userID uint, groupID uint) error
	GetGroupMembers(groupID uint, page, limit int) ([]models.GroupMember, int64, error)
	GetJoinedGroups(userID uint, page, limit int) ([]models.Group, int64, error)
	KickMember(ownerID uint, groupID uint, targetUserID uint, reason string) error

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
	notifSvc          NotificationService
}

func applyGroupUserPicture(user *models.User) {
	if user == nil {
		return
	}
	if user.Profile != nil {
		if user.Profile.ProfilePicture != "" {
			picture := user.Profile.ProfilePicture
			user.PictureURL = &picture
		} else {
			user.PictureURL = nil
		}
		user.Profile = nil
	}
}

func hydrateGroupCommentNode(node *GroupCommentNode) {
	if node == nil {
		return
	}
	applyGroupUserPicture(&node.User)
	for _, reply := range node.Replies {
		hydrateGroupCommentNode(reply)
	}
}

func hydrateGroupArticleComments(comments []models.GroupComment) {
	for i := range comments {
		applyGroupUserPicture(&comments[i].User)
	}
}

func hydrateGroupArticle(article *GroupArticleWithCount) {
	if article == nil {
		return
	}
	applyGroupUserPicture(&article.User)
	hydrateGroupArticleComments(article.Comments)
}

func hydrateGroupDetail(response *GroupDetailResponse) {
	if response == nil {
		return
	}
	applyGroupUserPicture(&response.Owner)
	for i := range response.Members {
		applyGroupUserPicture(&response.Members[i].User)
	}
	for i := range response.Articles {
		hydrateGroupArticle(&response.Articles[i])
	}
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	memberRepo repository.GroupMemberRepository,
	articleRepo repository.GroupArticleRepository,
	commentRepo repository.GroupCommentRepository,
	groupReactionRepo repository.GroupReactionRepository,
	notifSvc NotificationService,
) GroupService {
	return &groupService{
		groupRepo:         groupRepo,
		memberRepo:        memberRepo,
		articleRepo:       articleRepo,
		commentRepo:       commentRepo,
		groupReactionRepo: groupReactionRepo,
		notifSvc:          notifSvc,
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

func (s *groupService) GetGroups(page, limit int) ([]*GroupListItemResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	groups, total, err := s.groupRepo.FindGroups(page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Batch-fetch member + article counts in one query (avoids N+1).
	groupIDs := make([]uint, 0, len(groups))
	for i := range groups {
		groupIDs = append(groupIDs, groups[i].ID)
	}
	stats, err := s.groupRepo.CountGroupStats(groupIDs)
	if err != nil {
		// Non-fatal: degrade gracefully with zero counts.
		stats = map[uint][2]int{}
	}

	result := make([]*GroupListItemResponse, 0, len(groups))
	for i := range groups {
		group := groups[i]
		applyGroupUserPicture(&group.Owner)
		counts := stats[group.ID]
		result = append(result, &GroupListItemResponse{
			Group:        &group,
			MemberCount:  counts[0],
			ArticleCount: counts[1],
		})
	}

	if result == nil {
		result = []*GroupListItemResponse{}
	}
	return result, total, nil
}

func (s *groupService) GetGroupByID(id uint, articlePage, articleLimit int) (*GroupDetailResponse, error) {
	if articlePage < 1 {
		articlePage = 1
	}
	if articleLimit < 1 || articleLimit > 50 {
		articleLimit = 10
	}

	group, err := s.groupRepo.FindGroupByID(id)
	if err != nil {
		return nil, errors.New("grup tidak ditemukan")
	}

	// Fetch articles with pagination (bounded); avoids loading all articles at once.
	rawArticles, articleTotal, err := s.articleRepo.FindGroupArticlesPaginated(id, articlePage, articleLimit)
	if err != nil {
		rawArticles = []models.GroupArticle{}
		articleTotal = 0
	}

	articles := make([]GroupArticleWithCount, 0, len(rawArticles))
	for i := range rawArticles {
		article := rawArticles[i]
		applyGroupUserPicture(&article.User)

		// Fetch article images
		images, _ := s.articleRepo.FindArticleImages(article.ID)
		mediaURLs := make([]string, 0)
		for _, img := range images {
			mediaURLs = append(mediaURLs, img.ImageURL)
		}
		if len(mediaURLs) == 0 && article.MediaURL != nil {
			mediaURLs = append(mediaURLs, *article.MediaURL)
		}

		articles = append(articles, GroupArticleWithCount{
			ID:           article.ID,
			CreatedAt:    article.CreatedAt,
			UpdatedAt:    article.UpdatedAt,
			DeletedAt:    article.DeletedAt,
			GroupID:      article.GroupID,
			UserID:       article.UserID,
			Title:        article.Title,
			Content:      article.Content,
			MediaURL:     article.MediaURL,
			MediaURLs:    mediaURLs,
			User:         article.User,
			Comments:     []models.GroupComment{}, // comments loaded per-article via detail endpoint
			Reactions:    article.Reactions,
			CommentCount: 0, // not loaded here; use article detail endpoint for accurate count
		})
	}

	// Build article pagination metadata.
	totalPages := 0
	if articleLimit > 0 && articleTotal > 0 {
		totalPages = int((articleTotal + int64(articleLimit) - 1) / int64(articleLimit))
	}
	articlePagination := PaginationMeta{
		Total:      articleTotal,
		Page:       articlePage,
		Limit:      articleLimit,
		TotalPages: totalPages,
		HasNext:    articlePage < totalPages,
		HasPrev:    articlePage > 1,
	}

	response := &GroupDetailResponse{
		ID:                group.ID,
		CreatedAt:         group.CreatedAt,
		UpdatedAt:         group.UpdatedAt,
		DeletedAt:         group.DeletedAt,
		OwnerID:           group.OwnerID,
		Category:          group.Category,
		Title:             group.Title,
		Description:       group.Description,
		Rules:             group.Rules,
		BannerURL:         group.BannerURL,
		Owner:             group.Owner,
		Members:           group.Members,
		Articles:          articles,
		MemberCount:       len(group.Members),
		ArticleCount:      int(articleTotal),
		ArticlePagination: articlePagination,
	}
	hydrateGroupDetail(response)
	return response, nil
}

func (s *groupService) GetOwnedGroups(userID uint, page, limit int) ([]models.Group, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	groups, total, err := s.groupRepo.FindGroupsByOwner(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range groups {
		applyGroupUserPicture(&groups[i].Owner)
	}
	return groups, total, nil
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
	applyGroupUserPicture(&group.Owner)
	for i := range group.Members {
		applyGroupUserPicture(&group.Members[i].User)
	}
	for i := range group.Articles {
		applyGroupUserPicture(&group.Articles[i].User)
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

func (s *groupService) GetGroupMembers(groupID uint, page, limit int) ([]models.GroupMember, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	members, total, err := s.memberRepo.FindGroupMembers(groupID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range members {
		applyGroupUserPicture(&members[i].User)
	}
	return members, total, nil
}

func (s *groupService) GetJoinedGroups(userID uint, page, limit int) ([]models.Group, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	groups, total, err := s.memberRepo.FindJoinedGroups(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range groups {
		applyGroupUserPicture(&groups[i].Owner)
	}
	return groups, total, nil
}

func (s *groupService) KickMember(ownerID uint, groupID uint, targetUserID uint, reason string) error {
	if !s.isOwner(groupID, ownerID) {
		return errors.New("akses ditolak: hanya pemilik grup")
	}
	if ownerID == targetUserID {
		return errors.New("pemilik tidak dapat mengeluarkan dirinya sendiri")
	}
	if !s.isMember(groupID, targetUserID) {
		return errors.New("pengguna yang dituju bukan anggota grup")
	}
	if err := s.memberRepo.RemoveGroupMember(groupID, targetUserID); err != nil {
		return err
	}
	// Notify kicked user
	group, err := s.groupRepo.FindGroupByID(groupID)
	if err == nil {
		body := fmt.Sprintf("Kamu dikeluarkan dari grup %s", group.Title)
		if reason != "" {
			body = fmt.Sprintf("%s: %s", body, reason)
		}
		_ = s.notifSvc.Notify(
			targetUserID,
			"group_kicked",
			"Kamu dikeluarkan dari grup",
			body,
			"group",
			groupID,
		)
	}
	return nil
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func (s *groupService) CreateGroupArticle(userID uint, groupID uint, req GroupArticleRequest) (*models.GroupArticle, error) {
	if !s.isMember(groupID, userID) {
		return nil, errors.New("akses ditolak: hanya anggota grup")
	}
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("judul dan konten wajib diisi")
	}
	legacyMediaURL := req.MediaURL
	if len(req.MediaURLs) > 0 {
		first := req.MediaURLs[0]
		legacyMediaURL = &first
	}

	article := &models.GroupArticle{
		GroupID:  groupID,
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		MediaURL: legacyMediaURL,
	}
	if err := s.articleRepo.CreateGroupArticle(article); err != nil {
		return nil, errors.New("gagal membuat artikel")
	}

	// Save multiple images
	for _, mediaURL := range req.MediaURLs {
		if mediaURL != "" {
			img := &models.GroupArticleImage{
				ArticleID: article.ID,
				ImageURL:  mediaURL,
			}
			if err := s.articleRepo.CreateArticleImage(img); err != nil {
				_ = s.articleRepo.DeleteGroupArticle(article.ID)
				return nil, errors.New("gagal menyimpan media artikel")
			}
		}
	}

	// Reload to include preloaded relations expected by some clients (e.g. User).
	created, err := s.articleRepo.FindGroupArticleByID(article.ID)
	if err != nil {
		return article, nil
	}
	applyGroupUserPicture(&created.User)
	return created, nil
}

func (s *groupService) GetGroupArticleDetail(userID uint, articleID uint) (*GroupArticleDetailResponse, error) {
	article, err := s.articleRepo.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("artikel tidak ditemukan")
	}
	applyGroupUserPicture(&article.User)
	comments, _ := s.articleRepo.FindAllCommentsByArticleID(articleID)
	if !s.isMember(article.GroupID, userID) {
		comments = []models.GroupComment{}
	}
	hydrateGroupArticleComments(comments)
	commentTree := buildGroupCommentTree(comments)

	// Fetch article images
	images, _ := s.articleRepo.FindArticleImages(articleID)
	mediaURLs := make([]string, 0)
	for _, img := range images {
		mediaURLs = append(mediaURLs, img.ImageURL)
	}
	// If no images found but MediaURL exists (backward compat), use it
	if len(mediaURLs) == 0 && article.MediaURL != nil {
		mediaURLs = append(mediaURLs, *article.MediaURL)
	}

	return &GroupArticleDetailResponse{
		ID:           article.ID,
		GroupID:      article.GroupID,
		UserID:       article.UserID,
		User:         article.User,
		Title:        article.Title,
		Content:      article.Content,
		MediaURL:     article.MediaURL,
		MediaURLs:    mediaURLs,
		CommentCount: len(comments),
		Reactions:    article.Reactions,
		Comments:     commentTree,
		CreatedAt:    article.CreatedAt,
		UpdatedAt:    article.UpdatedAt,
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
	if len(req.MediaURLs) > 0 {
		first := req.MediaURLs[0]
		article.MediaURL = &first
	} else if req.MediaURL != nil {
		article.MediaURL = req.MediaURL
	}
	if err := s.articleRepo.UpdateGroupArticle(article); err != nil {
		return nil, errors.New("gagal memperbarui artikel")
	}

	// If new media URLs supplied, replace all old ones
	if len(req.MediaURLs) > 0 {
		// Delete existing images
		if err := s.articleRepo.DeleteArticleImages(articleID); err != nil {
			return nil, errors.New("gagal memperbarui media artikel")
		}
		// Create new images
		for _, mediaURL := range req.MediaURLs {
			if mediaURL != "" {
				img := &models.GroupArticleImage{
					ArticleID: articleID,
					ImageURL:  mediaURL,
				}
				if err := s.articleRepo.CreateArticleImage(img); err != nil {
					return nil, errors.New("gagal memperbarui media artikel")
				}
			}
		}
	}

	applyGroupUserPicture(&article.User)
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

	var parentComment *models.GroupComment
	if req.ParentCommentID != nil {
		parentComment, err = s.commentRepo.FindGroupCommentByID(*req.ParentCommentID)
		if err != nil {
			return nil, errors.New("komentar induk tidak ditemukan")
		}
		if parentComment.ArticleID != articleID {
			return nil, errors.New("komentar induk tidak valid")
		}
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
	createdComment, err := s.commentRepo.FindGroupCommentByID(comment.ID)
	if err == nil {
		applyGroupUserPicture(&createdComment.User)
		comment = createdComment
	} else {
		applyGroupUserPicture(&comment.User)
	}

	// Notify parent comment owner for replies.
	if parentComment != nil && parentComment.UserID != userID {
		_ = s.notifSvc.Notify(
			parentComment.UserID,
			"group_article_replied",
			"Balasan komentar artikel grup",
			"Komentarmu di artikel grup mendapat balasan baru",
			"group_comment",
			parentComment.ID,
		)
	}

	// Notify article owner for top-level comments.
	if req.ParentCommentID == nil && article.UserID != userID {
		_ = s.notifSvc.NotifyThrottled(
			article.UserID,
			"group_article_commented",
			"Komentar baru di artikel grup",
			"Artikel grupmu mendapat komentar baru",
			"group_article",
			articleID,
			30*time.Minute,
		)
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
	applyGroupUserPicture(&comment.User)
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
