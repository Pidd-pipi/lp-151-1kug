package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// 化名词组：化名由“形容词 + 动物 + 编号后缀”构成。
// 只作为每篇帖子内部的展示名称，不与用户身份表中的原始昵称挂钩。
var aliasAdjectives = []string{
	"快乐", "沉默", "勇敢", "温柔", "机智", "神秘", "阳光", "清冷",
	"调皮", "安静", "热情", "腼腆", "呆萌", "洒脱", "认真", "迷糊",
	"活泼", "淡定", "好奇", "可靠", "古怪", "浪漫", "倔强", "坦率",
}

var aliasAnimals = []string{
	"树懒", "刺猬", "海豚", "企鹅", "狐狸", "小鹿", "鲸鱼", "松鼠",
	"水獭", "考拉", "熊猫", "柴犬", "橘猫", "兔子", "浣熊", "海豹",
	"猫头鹰", "北极熊", "小狼", "山羊", "水豚", "仓鼠", "鹦鹉", "河豚",
}

// 头像仅在“帖子 + 身份”维度可复现，跨帖无法拼接头像 URL。
const aliasAvatarURL = "https://api.dicebear.com/9.x/bottts-neutral/svg?seed="

// Alias 是某个身份在某篇帖子内使用的树洞化名。
type Alias struct {
	Nickname string
	Avatar   string
}

// AliasService 依据（帖子ID, 身份ID）派生一次性化名。
//
// 派生过程使用服务端密钥签名：同一个身份在同一篇帖子的发帖与评论始终得到
// 同一组昵称与头像；换到另一篇帖子即得到全新一组。接口只暴露派生结果，
// 不返回任何可跨帖关联的稳定身份编号，历史数据升级后同样即时生效。
type AliasService interface {
	// For 返回身份 identityID 在帖子 postID 内的化名。
	For(postID, identityID uint) Alias
}

type aliasService struct {
	secret []byte
}

func NewAliasService(secret string) AliasService {
	return &aliasService{secret: []byte(secret)}
}

func (s *aliasService) For(postID, identityID uint) Alias {
	mac := hmac.New(sha256.New, s.secret)
	fmt.Fprintf(mac, "treehole-alias:post=%d:identity=%d", postID, identityID)
	sum := mac.Sum(nil)

	adj := aliasAdjectives[int(binary.BigEndian.Uint32(sum[0:4])%uint32(len(aliasAdjectives)))]
	animal := aliasAnimals[int(binary.BigEndian.Uint32(sum[4:8])%uint32(len(aliasAnimals)))]
	suffix := hex.EncodeToString(sum[8:10])
	// 头像种子独立取一段，且只放每帖内可复现的派生值，杜绝跨帖比对。
	seed := hex.EncodeToString(sum[16:28])

	return Alias{
		Nickname: adj + animal + "-" + suffix,
		Avatar:   aliasAvatarURL + seed,
	}
}
