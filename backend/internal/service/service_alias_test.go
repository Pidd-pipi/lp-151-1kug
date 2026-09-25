package service

import (
	"strings"
	"testing"
)

func TestAliasFor(t *testing.T) {
	svc := NewAliasService("test-alias-secret")

	base := svc.For(1, 10)
	if base.Nickname == "" || !strings.Contains(base.Nickname, "-") {
		t.Fatalf("化名格式异常: %q", base.Nickname)
	}
	if base.Avatar == "" || !strings.HasPrefix(base.Avatar, aliasAvatarURL) {
		t.Fatalf("化名头像格式异常: %q", base.Avatar)
	}

	cases := []struct {
		name   string
		postID uint
		userID uint
		want   Alias
		same   bool
	}{
		{"同帖同人重复派生保持稳定", 1, 10, base, true},
		{"同帖不同身份应得到另一组化名", 1, 11, base, false},
		{"同一身份换到另一帖应得到另一组化名", 2, 10, base, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.For(tc.postID, tc.userID)
			if got == tc.want != tc.same {
				t.Fatalf("For(%d,%d) = %+v, base = %+v, 期望相等=%v", tc.postID, tc.userID, got, tc.want, tc.same)
			}
		})
	}
}

func TestAliasDeterministicAcrossInstances(t *testing.T) {
	// 相同密钥下，新建实例结果应完全一致，保证重启/历史数据稳定。
	s1 := NewAliasService("stable-secret")
	s2 := NewAliasService("stable-secret")
	if got := s1.For(77, 88); got != s2.For(77, 88) {
		t.Fatalf("相同密钥派生化名不一致: %+v != %+v", got, s2.For(77, 88))
	}

	// 不同密钥应产生不同结果，避免外部根据公开 ID 自行离线枚举。
	s3 := NewAliasService("another-secret")
	if s1.For(77, 88) == s3.For(77, 88) {
		t.Fatal("不同密钥不应派生出相同化名")
	}
}
