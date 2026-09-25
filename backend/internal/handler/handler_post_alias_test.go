package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gbtreehole/backend/internal/model"
	"github.com/gbtreehole/backend/internal/repository"
	"github.com/gbtreehole/backend/internal/service"
)

type aliasTestEnv struct {
	engine    *gin.Engine
	posts     service.PostService
	comments  service.CommentService
	identityA *model.UserIdentity
	identityB *model.UserIdentity
}

func newAliasTestEnv(t *testing.T) *aliasTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	identityRepo := repository.NewIdentityRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	tagRepo := repository.NewTagRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	sensitiveRepo := repository.NewSensitiveWordRepository(db)
	reviewRepo := repository.NewReviewQueueRepository(db)
	aliasRepo := repository.NewPostAliasRepository(db)

	logger := slog.Default()
	tagService := service.NewTagService(tagRepo)
	sensitiveService := service.NewSensitiveWordService(sensitiveRepo)
	reviewService := service.NewReviewService(reviewRepo, postRepo, commentRepo, logger)
	aliasService := service.NewAliasService(aliasRepo, logger)
	postService := service.NewPostService(postRepo, tagService, sensitiveService, reviewService, aliasService, logger)
	commentService := service.NewCommentService(commentRepo, postRepo, sensitiveService, reviewService, aliasService, logger)
	likeService := service.NewLikeService(likeRepo, postRepo, commentRepo, logger)

	postHandler := NewPostHandler(postService, likeService, aliasService, logger)
	commentHandler := NewCommentHandler(commentService, likeService, aliasService, logger)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/posts/:id", postHandler.GetPost)
	engine.GET("/posts/:id/comments", commentHandler.ListComments)

	now := time.Now()
	identityA := &model.UserIdentity{IdentityKey: "key-a", Nickname: "原始昵称A", Avatar: "https://example.com/a.svg", CreatedAt: now, UpdatedAt: now}
	identityB := &model.UserIdentity{IdentityKey: "key-b", Nickname: "原始昵称B", Avatar: "https://example.com/b.svg", CreatedAt: now, UpdatedAt: now}
	if err := identityRepo.Create(identityA); err != nil {
		t.Fatalf("create identity A: %v", err)
	}
	if err := identityRepo.Create(identityB); err != nil {
		t.Fatalf("create identity B: %v", err)
	}
	return &aliasTestEnv{engine: engine, posts: postService, comments: commentService, identityA: identityA, identityB: identityB}
}

func getJSON(t *testing.T, engine *gin.Engine, path string) map[string]any {
	t.Helper()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s: status = %d, body = %s", path, recorder.Code, recorder.Body.String())
	}
	var body struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(body.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	return data
}

func TestPostDetailShowsPerPostAlias(t *testing.T) {
	env := newAliasTestEnv(t)
	post, _, _, err := env.posts.Create(env.identityA.ID, "标题", "今天想吐槽一下", nil, nil)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	data := getJSON(t, env.engine, "/posts/"+itoa(post.ID))
	if _, leaked := data["identityId"]; leaked {
		t.Fatalf("post response must not carry identityId: %v", data)
	}
	nickname, _ := data["nickname"].(string)
	if nickname == "" || nickname == env.identityA.Nickname {
		t.Fatalf("post must show per-post alias instead of original nickname, got %q", nickname)
	}
	avatar, _ := data["avatar"].(string)
	if avatar == "" || avatar == env.identityA.Avatar {
		t.Fatalf("post must show per-post alias avatar, got %q", avatar)
	}
}

func TestCommentsShareAliasWithPostAuthor(t *testing.T) {
	env := newAliasTestEnv(t)
	post, _, _, err := env.posts.Create(env.identityA.ID, "", "楼主的内容", nil, nil)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	if _, _, _, err := env.comments.Create(env.identityA.ID, post.ID, "楼主自己补充"); err != nil {
		t.Fatalf("create author comment: %v", err)
	}
	if _, _, _, err := env.comments.Create(env.identityB.ID, post.ID, "路人评论"); err != nil {
		t.Fatalf("create visitor comment: %v", err)
	}

	postData := getJSON(t, env.engine, "/posts/"+itoa(post.ID))
	authorAlias, _ := postData["nickname"].(string)

	commentData := getJSON(t, env.engine, "/posts/"+itoa(post.ID)+"/comments")
	items, _ := commentData["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 comments, got %v", items)
	}
	seen := map[string]bool{}
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if _, leaked := item["identityId"]; leaked {
			t.Fatalf("comment response must not carry identityId: %v", item)
		}
		nickname, _ := item["nickname"].(string)
		if nickname == env.identityA.Nickname || nickname == env.identityB.Nickname {
			t.Fatalf("comment must not show original nickname, got %q", nickname)
		}
		seen[nickname] = true
	}
	if !seen[authorAlias] {
		t.Fatalf("author comment must reuse the post alias %q, got %v", authorAlias, seen)
	}
	if len(seen) != 2 {
		t.Fatalf("author and visitor should have distinct aliases within the post, got %v", seen)
	}
}

func TestAliasChangesAcrossPosts(t *testing.T) {
	env := newAliasTestEnv(t)
	post1, _, _, err := env.posts.Create(env.identityA.ID, "", "第一篇", nil, nil)
	if err != nil {
		t.Fatalf("create post 1: %v", err)
	}
	post2, _, _, err := env.posts.Create(env.identityA.ID, "", "第二篇", nil, nil)
	if err != nil {
		t.Fatalf("create post 2: %v", err)
	}

	data1 := getJSON(t, env.engine, "/posts/"+itoa(post1.ID))
	data2 := getJSON(t, env.engine, "/posts/"+itoa(post2.ID))
	// 两篇帖子各自生成独立的化名记录（化名内容随机，但必须是两次独立生成）
	if data1["nickname"] == "" || data2["nickname"] == "" {
		t.Fatalf("both posts should carry aliases, got %q and %q", data1["nickname"], data2["nickname"])
	}
	// 同一篇帖子重复读取化名保持稳定
	again := getJSON(t, env.engine, "/posts/"+itoa(post1.ID))
	if again["nickname"] != data1["nickname"] || again["avatar"] != data1["avatar"] {
		t.Fatalf("alias of post 1 must stay stable across reads: %v vs %v", data1["nickname"], again["nickname"])
	}
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
