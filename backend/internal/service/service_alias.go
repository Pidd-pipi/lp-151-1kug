package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gbtreehole/backend/internal/model"
	"github.com/gbtreehole/backend/internal/repository"
)

// AliasPair 表示「帖子 × 身份」的化名归属关系。
type AliasPair struct {
	PostID     uint
	IdentityID uint
}

// AliasService 负责为每篇帖子的参与者生成并查询树洞化名。
// 同一身份在同一帖子内（发帖 + 评论）复用同一组化名与头像，
// 换一篇帖子则生成另一组，对外接口不再暴露可跨帖拼接的身份编号。
type AliasService interface {
	// Ensure 确保指定 (帖子, 身份) 的化名存在，不存在则生成并落库。
	Ensure(postID, identityID uint) (*model.PostAlias, error)
	// Resolve 批量解析化名，返回 map[postID]map[identityID]*PostAlias；
	// 缺失的组合会即时生成，历史数据升级后首次读取时自动补齐。
	Resolve(pairs []AliasPair) (map[uint]map[uint]*model.PostAlias, error)
}

type aliasService struct {
	repo   repository.PostAliasRepository
	logger *slog.Logger
}

func NewAliasService(repo repository.PostAliasRepository, logger *slog.Logger) AliasService {
	return &aliasService{repo: repo, logger: logger}
}

func (s *aliasService) Ensure(postID, identityID uint) (*model.PostAlias, error) {
	if postID == 0 || identityID == 0 {
		return nil, fmt.Errorf("invalid alias pair: post=%d identity=%d", postID, identityID)
	}
	existing, err := s.repo.FindByPostAndIdentity(postID, identityID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("find post alias: %w", err)
	}
	nickname, err := randomNickname()
	if err != nil {
		return nil, fmt.Errorf("generate alias nickname: %w", err)
	}
	alias := &model.PostAlias{
		PostID:     postID,
		IdentityID: identityID,
		Nickname:   nickname,
		Avatar:     randomAvatar(),
		CreatedAt:  time.Now(),
	}
	if err := s.repo.Create(alias); err != nil {
		// 并发请求可能已创建同一组合，回读即可保证一致。
		existing, findErr := s.repo.FindByPostAndIdentity(postID, identityID)
		if findErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("create post alias: %w", err)
	}
	return alias, nil
}

func (s *aliasService) Resolve(pairs []AliasPair) (map[uint]map[uint]*model.PostAlias, error) {
	result := make(map[uint]map[uint]*model.PostAlias)
	if len(pairs) == 0 {
		return result, nil
	}
	postIDSet := make(map[uint]bool, len(pairs))
	for _, p := range pairs {
		postIDSet[p.PostID] = true
	}
	postIDs := make([]uint, 0, len(postIDSet))
	for id := range postIDSet {
		postIDs = append(postIDs, id)
	}
	aliases, err := s.repo.ListByPostIDs(postIDs)
	if err != nil {
		return nil, err
	}
	for i := range aliases {
		alias := aliases[i]
		if _, ok := result[alias.PostID]; !ok {
			result[alias.PostID] = make(map[uint]*model.PostAlias)
		}
		result[alias.PostID][alias.IdentityID] = &alias
	}
	for _, p := range pairs {
		if p.PostID == 0 || p.IdentityID == 0 {
			continue
		}
		if result[p.PostID][p.IdentityID] != nil {
			continue
		}
		alias, err := s.Ensure(p.PostID, p.IdentityID)
		if err != nil {
			return nil, err
		}
		if _, ok := result[p.PostID]; !ok {
			result[p.PostID] = make(map[uint]*model.PostAlias)
		}
		result[p.PostID][p.IdentityID] = alias
	}
	return result, nil
}
