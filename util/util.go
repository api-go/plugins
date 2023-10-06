package util

import (
	"encoding/hex"
	"github.com/api-go/plugin"
	"github.com/ssgo/u"
)

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api-go/plugins/util",
		Name: "util",
		Objects: map[string]interface{}{
			// makeToken 生成指定长度的随机二进制数组
			// makeToken size token长度
			// makeToken return Hex编码的字符串
			"makeToken": func(size int) string {
				return hex.EncodeToString(u.MakeToken(size))
			},

			// makeTokenBytes 生成指定长度的随机二进制数组
			// makeTokenBytes size token长度
			// makeTokenBytes return 二进制数据
			"makeTokenBytes": u.MakeToken,
		},
	})
}
