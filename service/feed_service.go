package service

import (
	"errors"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validReactionTypes = map[string]bool{
	"like": true, "love": true, "haha": true,
	"wow": true, "sad": true, "angry": true,
}

// ─── Response types ───────────────────────────────────────────────────────────

// CommentNode is a recursive structure for infinitely nested replies.
type CommentNode struct {
	ID              uint              `json:"id"`
	PostID          uint              `json:"post_id"`
	UserID          uint              `json:"user_id"`
	User            models.User       `json:"user"`
	Content         string            `json:"content"`
	ParentCommentID *uint             `json:"parent_comment_id"`
	Reactions       []models.Reaction `json:"reactions"`
	Votes           []models.Vote     `json:"votes"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Replies         []*CommentNode    `json:"replies"`
}

type PostDetailResponse struct {
	ID        uint              `json:"id"`
	UserID    uint              `json:"user_id"`
	User      models.User       `json:"user"`
	Category  *string           `json:"category"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	ImageURL  *string           `json:"image_url"`
	Reactions []models.Reaction `json:"reactions"`
	Votes     []models.Vote     `json:"votes"`
	Comments  []*CommentNode    `json:"comments"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PostListItem is a lightweight summary for the feed list — shows counts only.
type PostListItem struct {
	ID            uint        `json:"id"`
	UserID        uint        `json:"user_id"`
	User          models.User `json:"user"`
	Category      *string     `json:"category"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	ImageURL      *string     `json:"image_url"`
	CommentCount  int         `json:"comment_count"`
	ReactionCount int         `json:"reaction_count"`
	VoteCount     int         `json:"vote_count"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// buildCommentTree converts a flat comment list into an infinitely nested tree.
func buildCommentTree(comments []models.Comment) []*CommentNode {
	nodeMap := make(map[uint]*CommentNode, len(comments))

	// Pass 1: create all nodes
	for i := range comments {
		c := &comments[i]
		nodeMap[c.ID] = &CommentNode{
			ID:              c.ID,
			PostID:          c.PostID,
			UserID:          c.UserID,
			User:            c.User,
			Content:         c.Content,
			ParentCommentID: c.ParentCommentID,
			Reactions:       c.Reactions,
			Votes:           c.Votes,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
			Replies:         []*CommentNode{},
		}
	}

	// Pass 2: attach children to parents
	var roots []*CommentNode
	for i := range comments {
		c := &comments[i]
		node := nodeMap[c.ID]
		if c.ParentCommentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*c.ParentCommentID]; ok {
				parent.Replies = append(parent.Replies, node)
			} else {
				// Orphaned (parent deleted) — show as root
				roots = append(roots, node)
			}
		}
	}

	if roots == nil {
		roots = []*CommentNode{}
	}
	return roots
}

// ─── DTOs ────────────────────────────────────────────────────────────────────

type PostRequest struct {
	Category *string `json:"category"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	ImageURL *string `json:"image_url"` // set by controller after upload
}

type CommentRequest struct {
	Content         string `json:"content"`
	ParentCommentID *uint  `json:"parent_comment_id"`
}

type ReactionRequest struct {
	Type string `json:"type"` // like|love|haha|wow|sad|angry
}

type VoteRequest struct {
	Value int `json:"value"` // 1 or -1
}

// ─── Post ─────────────────────────────────────────────────────────────────────

func CreatePost(userID uint, req PostRequest) (*models.Post, error) {
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("title and content are required")
	}
	post := &models.Post{
		UserID:   userID,
		Category: req.Category,
		Title:    req.Title,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}
	if err := repository.CreatePost(post); err != nil {
		return nil, errors.New("failed to create post")
	}
	return post, nil
}

func GetPosts(page, limit int) ([]*PostListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	posts, total, err := repository.FindPosts(page, limit)
	if err != nil {
		return nil, 0, err
	}
	var result []*PostListItem
	for _, p := range posts {
		result = append(result, &PostListItem{
			ID:            p.ID,
			UserID:        p.UserID,
			User:          p.User,
			Category:      p.Category,
			Title:         p.Title,
			Content:       p.Content,
			ImageURL:      p.ImageURL,
			CommentCount:  len(p.Comments),
			ReactionCount: len(p.Reactions),
			VoteCount:     len(p.Votes),
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		})
	}
	if result == nil {
		result = []*PostListItem{}
	}
	return result, total, nil
}

func GetPostByID(id uint) (*PostDetailResponse, error) {
	post, err := repository.FindPostByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Load all comments flat (any depth) with their reactions & votes
	comments, err := repository.FindAllCommentsByPostID(id)
	if err != nil {
		return nil, errors.New("failed to load comments")
	}

	return &PostDetailResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		User:      post.User,
		Category:  post.Category,
		Title:     post.Title,
		Content:   post.Content,
		ImageURL:  post.ImageURL,
		Reactions: post.Reactions,
		Votes:     post.Votes,
		Comments:  buildCommentTree(comments),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func UpdatePost(userID uint, postID uint, req PostRequest) (*models.Post, error) {
	post, err := repository.FindPostByID(postID)
	if err != nil {
		return nil, errors.New("post not found")
	}
	if post.UserID != userID {
		return nil, errors.New("access denied")
	}
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Category != nil {
		post.Category = req.Category
	}
	if req.ImageURL != nil {
		post.ImageURL = req.ImageURL
	}
	if err := repository.UpdatePost(post); err != nil {
		return nil, errors.New("failed to update post")
	}
	return post, nil
}

func DeletePost(userID uint, postID uint) error {
	post, err := repository.FindPostByID(postID)
	if err != nil {
		return errors.New("post not found")
	}
	if post.UserID != userID {
		return errors.New("access denied")
	}
	return repository.DeletePost(postID)
}

// ─── Comment ──────────────────────────────────────────────────────────────────

func AddComment(userID uint, postID uint, req CommentRequest) (*models.Comment, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	// Verify post exists
	if _, err := repository.FindPostByID(postID); err != nil {
		return nil, errors.New("post not found")
	}
	comment := &models.Comment{
		PostID:          postID,
		UserID:          userID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
	}
	if err := repository.CreateComment(comment); err != nil {
		return nil, errors.New("failed to add comment")
	}
	return comment, nil
}

func UpdateComment(userID uint, commentID uint, req CommentRequest) (*models.Comment, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	comment, err := repository.FindCommentByID(commentID)
	if err != nil {
		return nil, errors.New("comment not found")
	}
	if comment.UserID != userID {
		return nil, errors.New("access denied")
	}
	comment.Content = req.Content
	if err := repository.UpdateComment(comment); err != nil {
		return nil, errors.New("failed to update comment")
	}
	return comment, nil
}

func DeleteComment(userID uint, commentID uint) error {
	comment, err := repository.FindCommentByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}
	if comment.UserID != userID {
		return errors.New("access denied")
	}
	return repository.DeleteComment(commentID)
}

// ─── Reaction ─────────────────────────────────────────────────────────────────

func ReactToPost(userID uint, postID uint, req ReactionRequest) (string, error) {
	if !validReactionTypes[req.Type] {
		return "", errors.New("invalid reaction type: must be like, love, haha, wow, sad, or angry")
	}

	existing, err := repository.FindReaction(userID, &postID, nil)
	if err == nil {
		// Exists — same type: toggle off; different type: update
		if existing.Type == req.Type {
			_ = repository.DeleteReaction(existing.ID)
			return "removed", nil
		}
		existing.Type = req.Type
		_ = repository.UpdateReaction(existing)
		return "updated", nil
	}
	// Does not exist — create
	reaction := &models.Reaction{UserID: userID, PostID: &postID, Type: req.Type}
	_ = repository.CreateReaction(reaction)
	return "added", nil
}

func ReactToComment(userID uint, commentID uint, req ReactionRequest) (string, error) {
	if !validReactionTypes[req.Type] {
		return "", errors.New("invalid reaction type: must be like, love, haha, wow, sad, or angry")
	}

	existing, err := repository.FindReaction(userID, nil, &commentID)
	if err == nil {
		if existing.Type == req.Type {
			_ = repository.DeleteReaction(existing.ID)
			return "removed", nil
		}
		existing.Type = req.Type
		_ = repository.UpdateReaction(existing)
		return "updated", nil
	}
	reaction := &models.Reaction{UserID: userID, CommentID: &commentID, Type: req.Type}
	_ = repository.CreateReaction(reaction)
	return "added", nil
}

// ─── Vote ─────────────────────────────────────────────────────────────────────

func VotePost(userID uint, postID uint, req VoteRequest) (string, error) {
	if req.Value != 1 && req.Value != -1 {
		return "", errors.New("value must be 1 (upvote) or -1 (downvote)")
	}

	existing, err := repository.FindVote(userID, &postID, nil)
	if err == nil {
		if existing.Value == req.Value {
			_ = repository.DeleteVote(existing.ID)
			return "removed", nil
		}
		existing.Value = req.Value
		_ = repository.UpdateVote(existing)
		return "flipped", nil
	}
	vote := &models.Vote{UserID: userID, PostID: &postID, Value: req.Value}
	_ = repository.CreateVote(vote)
	return "added", nil
}

func VoteComment(userID uint, commentID uint, req VoteRequest) (string, error) {
	if req.Value != 1 && req.Value != -1 {
		return "", errors.New("value must be 1 (upvote) or -1 (downvote)")
	}

	existing, err := repository.FindVote(userID, nil, &commentID)
	if err == nil {
		if existing.Value == req.Value {
			_ = repository.DeleteVote(existing.ID)
			return "removed", nil
		}
		existing.Value = req.Value
		_ = repository.UpdateVote(existing)
		return "flipped", nil
	}
	vote := &models.Vote{UserID: userID, CommentID: &commentID, Value: req.Value}
	_ = repository.CreateVote(vote)
	return "added", nil
}
