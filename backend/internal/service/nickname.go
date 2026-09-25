package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var adjectiveList = []string{"快乐", "沉默", "勇敢", "温柔", "机智", "神秘", "阳光", "清冷"}
var nounList = []string{"树懒", "刺猬", "海豚", "企鹅", "狐狸", "小鹿", "鲸鱼", "松鼠"}

// randomNickname 生成「形容词 + 动物 + 短码」风格的匿名昵称。
func randomNickname() (string, error) {
	ai, err := randInt(len(adjectiveList))
	if err != nil {
		return "", err
	}
	ni, err := randInt(len(nounList))
	if err != nil {
		return "", err
	}
	return adjectiveList[ai] + nounList[ni] + "-" + shortCode(3), nil
}

// randomAvatar 生成基于随机种子的匿名头像地址。
func randomAvatar() string {
	return "https://api.dicebear.com/9.x/bottts-neutral/svg?seed=" + shortCode(12)
}

func randInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("rand int: %w", err)
	}
	return int(n.Int64()), nil
}

func shortCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
