package runtime

import (
	"github.com/api-go/plugin"
	"runtime"
	"time"
)

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api-go/plugins/runtime",
		Name: "runtime",
		Objects: map[string]interface{}{
			// sleep 程序等待指定时间
			// sleep ms 休眠时长，单位ms
			"sleep": func(ms int) {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			},
			// os 获取操作系统名称
			// os return 操作系统名称
			"os": func() string {
				return runtime.GOOS
			},
			// arch 获取操作系统构建版本
			// arch return 操作系统构建版本
			"arch": func() string {
				return runtime.GOARCH
			},
		},
	})
}
