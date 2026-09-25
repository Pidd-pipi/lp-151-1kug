package repository

import (
	"testing"
	"time"

	"github.com/gbtreehole/backend/internal/model"
)

func TestPostAliasRepository(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostAliasRepository(db)

	alias := &model.PostAlias{PostID: 1, IdentityID: 2, Nickname: "沉默树懒-a1b", Avatar: "https://example.com/a.svg", CreatedAt: time.Now()}
	if err := repo.Create(alias); err != nil {
		t.Fatalf("create alias: %v", err)
	}

	got, err := repo.FindByPostAndIdentity(1, 2)
	if err != nil {
		t.Fatalf("find by post and identity: %v", err)
	}
	if got.Nickname != "沉默树懒-a1b" {
		t.Fatalf("nickname mismatch: %s", got.Nickname)
	}

	if _, err := repo.FindByPostAndIdentity(1, 3); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := repo.FindByPostAndIdentity(2, 2); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for other post, got %v", err)
	}

	other := &model.PostAlias{PostID: 2, IdentityID: 2, Nickname: "勇敢企鹅-c3d", Avatar: "https://example.com/b.svg", CreatedAt: time.Now()}
	if err := repo.Create(other); err != nil {
		t.Fatalf("create alias for other post: %v", err)
	}

	aliases, err := repo.ListByPostIDs([]uint{1, 2})
	if err != nil {
		t.Fatalf("list by post ids: %v", err)
	}
	if len(aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(aliases))
	}

	aliases, err = repo.ListByPostIDs([]uint{1})
	if err != nil {
		t.Fatalf("list by single post id: %v", err)
	}
	if len(aliases) != 1 || aliases[0].PostID != 1 {
		t.Fatalf("expected 1 alias of post 1, got %+v", aliases)
	}

	// 同一 (帖子, 身份) 组合不允许重复
	dup := &model.PostAlias{PostID: 1, IdentityID: 2, Nickname: "重复-x0y", CreatedAt: time.Now()}
	if err := repo.Create(dup); err == nil {
		t.Fatalf("expected unique constraint violation, got nil")
	}
}
