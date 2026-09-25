package service

import (
	"log/slog"
	"testing"

	"github.com/gbtreehole/backend/internal/model"
	"github.com/gbtreehole/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAliasService(t *testing.T) AliasService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.PostAlias{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewAliasService(repository.NewPostAliasRepository(db), slog.Default())
}

func TestAliasEnsureStableWithinPost(t *testing.T) {
	svc := newAliasService(t)
	first, err := svc.Ensure(1, 10)
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if first.Nickname == "" || first.Avatar == "" {
		t.Fatalf("alias should carry nickname and avatar: %+v", first)
	}
	second, err := svc.Ensure(1, 10)
	if err != nil {
		t.Fatalf("ensure again: %v", err)
	}
	if first.Nickname != second.Nickname || first.Avatar != second.Avatar {
		t.Fatalf("alias must stay stable within one post: %+v vs %+v", first, second)
	}
}

func TestAliasEnsurePerPost(t *testing.T) {
	svc := newAliasService(t)
	postA, err := svc.Ensure(1, 10)
	if err != nil {
		t.Fatalf("ensure post 1: %v", err)
	}
	postB, err := svc.Ensure(2, 10)
	if err != nil {
		t.Fatalf("ensure post 2: %v", err)
	}
	if postA.ID == postB.ID {
		t.Fatalf("each post must get its own alias row")
	}
	// 各自帖子内保持稳定
	againA, err := svc.Ensure(1, 10)
	if err != nil {
		t.Fatalf("ensure post 1 again: %v", err)
	}
	if againA.Nickname != postA.Nickname {
		t.Fatalf("alias for post 1 changed: %s -> %s", postA.Nickname, againA.Nickname)
	}
}

func TestAliasResolveBackfillsHistoricalData(t *testing.T) {
	svc := newAliasService(t)
	// 模拟升级前的历史数据：帖子与评论已存在，但没有任何化名记录。
	pairs := []AliasPair{
		{PostID: 7, IdentityID: 1}, // 楼主
		{PostID: 7, IdentityID: 2}, // 评论者 A
		{PostID: 7, IdentityID: 3}, // 评论者 B
		{PostID: 8, IdentityID: 1}, // 同一身份在另一篇帖子
	}
	resolved, err := svc.Resolve(pairs)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	for _, p := range pairs {
		alias := resolved[p.PostID][p.IdentityID]
		if alias == nil {
			t.Fatalf("missing alias for pair %+v", p)
		}
		if alias.Nickname == "" || alias.Avatar == "" {
			t.Fatalf("alias for pair %+v is incomplete: %+v", p, alias)
		}
	}
	// 同一身份在不同帖子使用不同化名记录
	if resolved[7][1].ID == resolved[8][1].ID {
		t.Fatalf("identity 1 should have distinct aliases per post")
	}
	// 再次解析结果保持一致（已落库，不会重新生成）
	again, err := svc.Resolve(pairs)
	if err != nil {
		t.Fatalf("resolve again: %v", err)
	}
	for _, p := range pairs {
		if again[p.PostID][p.IdentityID].Nickname != resolved[p.PostID][p.IdentityID].Nickname {
			t.Fatalf("alias for pair %+v changed between reads", p)
		}
	}
}

func TestAliasResolveEmpty(t *testing.T) {
	svc := newAliasService(t)
	resolved, err := svc.Resolve(nil)
	if err != nil {
		t.Fatalf("resolve empty: %v", err)
	}
	if len(resolved) != 0 {
		t.Fatalf("expected empty result, got %v", resolved)
	}
}
