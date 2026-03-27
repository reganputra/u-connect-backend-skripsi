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

// ─── Interface ────────────────────────────────────────────────────────────────

type FeedService interface {
	CreatePost(userID uint, req PostRequest) (*models.Post, error)
	GetPosts(page, limit int) ([]*PostListItem, int64, error)
	GetPostByID(id uint) (*PostDetailResponse, error)
	UpdatePost(userID uint, postID uint, req PostRequest) (*models.Post, error)
	DeletePost(userID uint, postID uint) error

	AddComment(userID uint, postID uint, req CommentRequest) (*models.Comment, error)
	UpdateComment(userID uint, commentID uint, req CommentRequest) (*models.Comment, error)
	DeleteComment(userID uint, commentID uint) error

	ReactToPost(userID uint, postID uint, req ReactionRequest) (string, error)
	ReactToComment(userID uint, commentID uint, req ReactionRequest) (string, error)

	VotePost(userID uint, postID uint, req VoteRequest) (string, error)
	VoteComment(userID uint, commentID uint, req VoteRequest) (string, error)
}

// ─── Struct ───────────────────────────────────────────────────────────────────

type feedService struct {
	postRepo     repository.PostRepository
	commentRepo  repository.CommentRepository
	reactionRepo repository.ReactionRepository
	voteRepo     repository.VoteRepository
}

func NewFeedService(
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	reactionRepo repository.ReactionRepository,
	voteRepo repository.VoteRepository,
) FeedService {
	return &feedService{
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		reactionRepo: reactionRepo,
		voteRepo:     voteRepo,
	}
}

// ─── Post ─────────────────────────────────────────────────────────────────────

func (s *feedService) CreatePost(userID uint, req PostRequest) (*models.Post, error) {
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("judul dan konten wajib diisi")
	}
	post := &models.Post{
		UserID:   userID,
		Category: req.Category,
		Title:    req.Title,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}
	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, errors.New("gagal membuat postingan")
	}
	return post, nil
}

func (s *feedService) GetPosts(page, limit int) ([]*PostListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	posts, total, err := s.postRepo.FindPosts(page, limit)
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

func (s *feedService) GetPostByID(id uint) (*PostDetailResponse, error) {
	post, err := s.postRepo.FindPostByID(id)
	if err != nil {
		return nil, errors.New("postingan tidak ditemukan")
	}

	comments, err := s.postRepo.FindAllCommentsByPostID(id)
	if err != nil {
		return nil, errors.New("gagal memuat komentar")
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

func (s *feedService) UpdatePost(userID uint, postID uint, req PostRequest) (*models.Post, error) {
	post, err := s.postRepo.FindPostByID(postID)
	if err != nil {
		return nil, errors.New("postingan tidak ditemukan")
	}
	if post.UserID != userID {
		return nil, errors.New("akses ditolak")
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
	if err := s.postRepo.UpdatePost(post); err != nil {
		return nil, errors.New("gagal memperbarui postingan")
	}
	return post, nil
}

func (s *feedService) DeletePost(userID uint, postID uint) error {
	post, err := s.postRepo.FindPostByID(postID)
	if err != nil {
		return errors.New("postingan tidak ditemukan")
	}
	if post.UserID != userID {
		return errors.New("akses ditolak")
	}
	return s.postRepo.DeletePost(postID)
}

// ─── Comment ──────────────────────────────────────────────────────────────────

func (s *feedService) AddComment(userID uint, postID uint, req CommentRequest) (*models.Comment, error) {
	if req.Content == "" {
		return nil, errors.New("konten wajib diisi")
	}
	if _, err := s.postRepo.FindPostByID(postID); err != nil {
		return nil, errors.New("postingan tidak ditemukan")
	}
	comment := &models.Comment{
		PostID:          postID,
		UserID:          userID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
	}
	if err := s.commentRepo.CreateComment(comment); err != nil {
		return nil, errors.New("gagal menambahkan komentar")
	}
	return comment, nil
}

func (s *feedService) UpdateComment(userID uint, commentID uint, req CommentRequest) (*models.Comment, error) {
	if req.Content == "" {
		return nil, errors.New("konten wajib diisi")
	}
	comment, err := s.commentRepo.FindCommentByID(commentID)
	if err != nil {
		return nil, errors.New("komentar tidak ditemukan")
	}
	if comment.UserID != userID {
		return nil, errors.New("akses ditolak")
	}
	comment.Content = req.Content
	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, errors.New("gagal memperbarui komentar")
	}
	return comment, nil
}

func (s *feedService) DeleteComment(userID uint, commentID uint) error {
	comment, err := s.commentRepo.FindCommentByID(commentID)
	if err != nil {
		return errors.New("komentar tidak ditemukan")
	}
	if comment.UserID != userID {
		return errors.New("akses ditolak")
	}
	return s.commentRepo.DeleteComment(commentID)
}

// ─── Reaction ─────────────────────────────────────────────────────────────────

func (s *feedService) ReactToPost(userID uint, postID uint, req ReactionRequest) (string, error) {
	if !validReactionTypes[req.Type] {
		return "", errors.New("jenis reaksi tidak valid: harus like, love, haha, wow, sad, atau angry")
	}

	existing, err := s.reactionRepo.FindReaction(userID, &postID, nil)
	if err == nil {
		if existing.Type == req.Type {
			if err := s.reactionRepo.DeleteReaction(existing.ID); err != nil {
				return "", errors.New("gagal menghapus reaksi")
			}
			return "removed", nil
		}
		existing.Type = req.Type
		if err := s.reactionRepo.UpdateReaction(existing); err != nil {
			return "", errors.New("gagal memperbarui reaksi")
		}
		return "updated", nil
	}
	reaction := &models.Reaction{UserID: userID, PostID: &postID, Type: req.Type}
	if err := s.reactionRepo.CreateReaction(reaction); err != nil {
		return "", errors.New("gagal menambahkan reaksi")
	}
	return "added", nil
}

func (s *feedService) ReactToComment(userID uint, commentID uint, req ReactionRequest) (string, error) {
	if !validReactionTypes[req.Type] {
		return "", errors.New("jenis reaksi tidak valid: harus like, love, haha, wow, sad, atau angry")
	}

	existing, err := s.reactionRepo.FindReaction(userID, nil, &commentID)
	if err == nil {
		if existing.Type == req.Type {
			if err := s.reactionRepo.DeleteReaction(existing.ID); err != nil {
				return "", errors.New("gagal menghapus reaksi")
			}
			return "removed", nil
		}
		existing.Type = req.Type
		if err := s.reactionRepo.UpdateReaction(existing); err != nil {
			return "", errors.New("gagal memperbarui reaksi")
		}
		return "updated", nil
	}
	reaction := &models.Reaction{UserID: userID, CommentID: &commentID, Type: req.Type}
	if err := s.reactionRepo.CreateReaction(reaction); err != nil {
		return "", errors.New("gagal menambahkan reaksi")
	}
	return "added", nil
}

// ─── Vote ─────────────────────────────────────────────────────────────────────

func (s *feedService) VotePost(userID uint, postID uint, req VoteRequest) (string, error) {
	if req.Value != 1 && req.Value != -1 {
		return "", errors.New("nilai harus 1 (upvote) atau -1 (downvote)")
	}

	existing, err := s.voteRepo.FindVote(userID, &postID, nil)
	if err == nil {
		if existing.Value == req.Value {
			if err := s.voteRepo.DeleteVote(existing.ID); err != nil {
				return "", errors.New("gagal menghapus vote")
			}
			return "removed", nil
		}
		existing.Value = req.Value
		if err := s.voteRepo.UpdateVote(existing); err != nil {
			return "", errors.New("gagal memperbarui vote")
		}
		return "flipped", nil
	}
	vote := &models.Vote{UserID: userID, PostID: &postID, Value: req.Value}
	if err := s.voteRepo.CreateVote(vote); err != nil {
		return "", errors.New("gagal menambahkan vote")
	}
	return "added", nil
}

func (s *feedService) VoteComment(userID uint, commentID uint, req VoteRequest) (string, error) {
	if req.Value != 1 && req.Value != -1 {
		return "", errors.New("nilai harus 1 (upvote) atau -1 (downvote)")
	}

	existing, err := s.voteRepo.FindVote(userID, nil, &commentID)
	if err == nil {
		if existing.Value == req.Value {
			if err := s.voteRepo.DeleteVote(existing.ID); err != nil {
				return "", errors.New("gagal menghapus vote")
			}
			return "removed", nil
		}
		existing.Value = req.Value
		if err := s.voteRepo.UpdateVote(existing); err != nil {
			return "", errors.New("gagal memperbarui vote")
		}
		return "flipped", nil
	}
	vote := &models.Vote{UserID: userID, CommentID: &commentID, Value: req.Value}
	if err := s.voteRepo.CreateVote(vote); err != nil {
		return "", errors.New("gagal menambahkan vote")
	}
	return "added", nil
}
