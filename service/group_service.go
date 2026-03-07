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
	BannerURL   *string `json:"banner_url"` // set by controller after upload
}

type GroupArticleRequest struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	MediaURL *string `json:"media_url"` // set by controller after upload
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

// ─── helpers ──────────────────────────────────────────────────────────────────

func isMember(groupID, userID uint) bool {
	_, err := repository.FindGroupMember(groupID, userID)
	return err == nil
}

func isOwner(groupID, userID uint) bool {
	m, err := repository.FindGroupMember(groupID, userID)
	return err == nil && m.Role == "owner"
}

// ─── Group CRUD ───────────────────────────────────────────────────────────────

func CreateGroup(userID uint, req GroupRequest) (*models.Group, error) {
	if req.Category == "" || req.Title == "" {
		return nil, errors.New("category and title are required")
	}
	group := &models.Group{
		OwnerID:     userID,
		Category:    req.Category,
		Title:       req.Title,
		Description: req.Description,
		Rules:       req.Rules,
		BannerURL:   req.BannerURL,
	}
	if err := repository.CreateGroup(group); err != nil {
		return nil, errors.New("failed to create group")
	}
	// Auto-join creator as owner
	member := &models.GroupMember{GroupID: group.ID, UserID: userID, Role: "owner"}
	_ = repository.AddGroupMember(member)
	return group, nil
}

func GetGroups() ([]models.Group, error) {
	return repository.FindGroups()
}

func GetGroupByID(id uint) (*models.Group, error) {
	group, err := repository.FindGroupByID(id)
	if err != nil {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func UpdateGroup(userID uint, groupID uint, req GroupRequest) (*models.Group, error) {
	if !isOwner(groupID, userID) {
		return nil, errors.New("access denied: owner only")
	}
	group, err := repository.FindGroupByID(groupID)
	if err != nil {
		return nil, errors.New("group not found")
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
	if err := repository.UpdateGroup(group); err != nil {
		return nil, errors.New("failed to update group")
	}
	return group, nil
}

func DeleteGroup(userID uint, groupID uint) error {
	if !isOwner(groupID, userID) {
		return errors.New("access denied: owner only")
	}
	return repository.DeleteGroup(groupID)
}

// ─── Membership ───────────────────────────────────────────────────────────────

func JoinGroup(userID uint, groupID uint) error {
	if _, err := repository.FindGroupByID(groupID); err != nil {
		return errors.New("group not found")
	}
	if isMember(groupID, userID) {
		return errors.New("already a member of this group")
	}
	return repository.AddGroupMember(&models.GroupMember{GroupID: groupID, UserID: userID, Role: "member"})
}

func LeaveGroup(userID uint, groupID uint) error {
	if isOwner(groupID, userID) {
		return errors.New("owner cannot leave the group")
	}
	if !isMember(groupID, userID) {
		return errors.New("not a member of this group")
	}
	return repository.RemoveGroupMember(groupID, userID)
}

func GetGroupMembers(groupID uint) ([]models.GroupMember, error) {
	return repository.FindGroupMembers(groupID)
}

func GetJoinedGroups(userID uint) ([]models.Group, error) {
	return repository.FindJoinedGroups(userID)
}

func KickMember(ownerID uint, groupID uint, targetUserID uint) error {
	if !isOwner(groupID, ownerID) {
		return errors.New("access denied: owner only")
	}
	if ownerID == targetUserID {
		return errors.New("owner cannot kick themselves")
	}
	if !isMember(groupID, targetUserID) {
		return errors.New("target user is not a member")
	}
	return repository.RemoveGroupMember(groupID, targetUserID)
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func CreateGroupArticle(userID uint, groupID uint, req GroupArticleRequest) (*models.GroupArticle, error) {
	if !isMember(groupID, userID) {
		return nil, errors.New("access denied: members only")
	}
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("title and content are required")
	}
	article := &models.GroupArticle{
		GroupID:  groupID,
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		MediaURL: req.MediaURL,
	}
	if err := repository.CreateGroupArticle(article); err != nil {
		return nil, errors.New("failed to create article")
	}
	return article, nil
}

func GetGroupArticleDetail(userID uint, articleID uint) (*GroupArticleDetailResponse, error) {
	article, err := repository.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("article not found")
	}
	// Non-members: return article but no comments
	comments, _ := repository.FindAllCommentsByArticleID(articleID)
	if !isMember(article.GroupID, userID) {
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

func UpdateGroupArticle(userID uint, articleID uint, req GroupArticleRequest) (*models.GroupArticle, error) {
	article, err := repository.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("article not found")
	}
	// Article owner or group owner can update
	if article.UserID != userID && !isOwner(article.GroupID, userID) {
		return nil, errors.New("access denied")
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
	if err := repository.UpdateGroupArticle(article); err != nil {
		return nil, errors.New("failed to update article")
	}
	return article, nil
}

func DeleteGroupArticle(userID uint, articleID uint) error {
	article, err := repository.FindGroupArticleByID(articleID)
	if err != nil {
		return errors.New("article not found")
	}
	if article.UserID != userID && !isOwner(article.GroupID, userID) {
		return errors.New("access denied")
	}
	return repository.DeleteGroupArticle(articleID)
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func AddGroupComment(userID uint, articleID uint, req GroupCommentRequest) (*models.GroupComment, error) {
	article, err := repository.FindGroupArticleByID(articleID)
	if err != nil {
		return nil, errors.New("article not found")
	}
	if !isMember(article.GroupID, userID) {
		return nil, errors.New("access denied: members only")
	}
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	comment := &models.GroupComment{
		ArticleID:       articleID,
		UserID:          userID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
	}
	if err := repository.CreateGroupComment(comment); err != nil {
		return nil, errors.New("failed to add comment")
	}
	return comment, nil
}

func UpdateGroupComment(userID uint, commentID uint, req GroupCommentRequest) (*models.GroupComment, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	comment, err := repository.FindGroupCommentByID(commentID)
	if err != nil {
		return nil, errors.New("comment not found")
	}
	if comment.UserID != userID {
		return nil, errors.New("access denied")
	}
	comment.Content = req.Content
	if err := repository.UpdateGroupComment(comment); err != nil {
		return nil, errors.New("failed to update comment")
	}
	return comment, nil
}

func DeleteGroupComment(userID uint, commentID uint) error {
	comment, err := repository.FindGroupCommentByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}
	// Comment owner or group owner can delete
	article, _ := repository.FindGroupArticleByID(comment.ArticleID)
	if comment.UserID != userID && (article == nil || !isOwner(article.GroupID, userID)) {
		return errors.New("access denied")
	}
	return repository.DeleteGroupComment(commentID)
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func ReactToGroupArticle(userID uint, articleID uint, req GroupReactionRequest) (string, error) {
	if !validGroupReactionTypes[req.Type] {
		return "", errors.New("invalid reaction type")
	}
	article, err := repository.FindGroupArticleByID(articleID)
	if err != nil {
		return "", errors.New("article not found")
	}
	if !isMember(article.GroupID, userID) {
		return "", errors.New("access denied: members only")
	}
	existing, err := repository.FindGroupReaction(userID, &articleID, nil)
	if err == nil {
		if existing.Type == req.Type {
			_ = repository.DeleteGroupReaction(existing.ID)
			return "removed", nil
		}
		existing.Type = req.Type
		_ = repository.UpdateGroupReaction(existing)
		return "updated", nil
	}
	_ = repository.CreateGroupReaction(&models.GroupReaction{UserID: userID, ArticleID: &articleID, Type: req.Type})
	return "added", nil
}

func ReactToGroupComment(userID uint, commentID uint, req GroupReactionRequest) (string, error) {
	if !validGroupReactionTypes[req.Type] {
		return "", errors.New("invalid reaction type")
	}
	comment, err := repository.FindGroupCommentByID(commentID)
	if err != nil {
		return "", errors.New("comment not found")
	}
	article, _ := repository.FindGroupArticleByID(comment.ArticleID)
	if article != nil && !isMember(article.GroupID, userID) {
		return "", errors.New("access denied: members only")
	}
	existing, err := repository.FindGroupReaction(userID, nil, &commentID)
	if err == nil {
		if existing.Type == req.Type {
			_ = repository.DeleteGroupReaction(existing.ID)
			return "removed", nil
		}
		existing.Type = req.Type
		_ = repository.UpdateGroupReaction(existing)
		return "updated", nil
	}
	_ = repository.CreateGroupReaction(&models.GroupReaction{UserID: userID, CommentID: &commentID, Type: req.Type})
	return "added", nil
}
