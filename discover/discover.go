package discover

import (
	"github.com/api0-work/plugin"
	"github.com/ssgo/discover"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"net/http"
	"strings"
)

type DiscoverApp struct {
	app    string
	token  string
	caller *discover.Caller
	logger *log.Logger
	globalHeaders map[string]string
}

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api0-work/plugins/discover",
		Name: "discover",
		Objects: map[string]interface{}{
			"fetch": GetDiscoverApp,
		},
	})
}

// GetDiscoverApp 获得一个服务发现的客户端，如果在配置(server>calls)中指定了AccessToken、超时时间或者HTTP协议(如:iZg753bnsBxTOqHjaeEdt2szvov95eLq34G6jiHBoeE=:1:s:60s)会自动在获得的客户端中设置好
// GetDiscoverApp app 需要访问的服务名称
// GetDiscoverApp return 服务发现客户端对象，支持的方法：get、post、put、delete、head、setGlobalHeaders
func GetDiscoverApp(app string, request *http.Request, logger *log.Logger) *DiscoverApp {
	return &DiscoverApp{
		app:    app,
		logger: logger,
		caller: discover.NewCaller(request, logger),
		globalHeaders: map[string]string{},
	}
}

// SetGlobalHeaders 设置固定的HTTP头部信息，在每个请求中都加入这些HTTP头
// * SetGlobalHeaders 传入一个Key-Value对象的HTTP头信息
func (dApp *DiscoverApp) SetGlobalHeaders(headers map[string]string) {
	dApp.globalHeaders = headers
}

// Get 发送GET请求
// * path /开头的请求路径，调用时会自动加上负载均衡到的目标节点的URL前缀发送HTTP请求
// * headers 传入一个Key-Value对象的HTTP头信息，如果不指定头信息这个参数可以省略不传
// * return 返回结果对象，如果返回值是JSON格式，将自动转化为对象否则将字符串放在.result中，如发生错误将抛出异常，返回的对象中还包括：headers、statusCode、statusMessage
func (dApp *DiscoverApp) Get(path string, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(dApp.logger, dApp.caller.Get(dApp.app, fixHTTPPath(path), dApp.makeHeaderArray(headers)...))
}

// Post 发送POST请求
// * data 可以传入任意类型，如果不是字符串或二进制数组时会自动添加application/json头，数据将以json格式发送
func (dApp *DiscoverApp) Post(path string, data *map[string]interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(dApp.logger, dApp.caller.Post(dApp.app, fixHTTPPath(path), data, dApp.makeHeaderArray(headers)...))
}

// Put 发送PUT请求
func (dApp *DiscoverApp) Put(path string, data *map[string]interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(dApp.logger, dApp.caller.Put(dApp.app, fixHTTPPath(path), data, dApp.makeHeaderArray(headers)...))
}

// Delete 发送DELETE请求
func (dApp *DiscoverApp) Delete(path string, data *map[string]interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(dApp.logger, dApp.caller.Delete(dApp.app, fixHTTPPath(path), data, dApp.makeHeaderArray(headers)...))
}

// Head 发送HEAD请求
func (dApp *DiscoverApp) Head(path string, data *map[string]interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(dApp.logger, dApp.caller.Head(dApp.app, fixHTTPPath(path), data, dApp.makeHeaderArray(headers)...))
}

func (dApp *DiscoverApp) makeHeaderArray(in *map[string]string) []string {
	out := make([]string, 0)
	if dApp.globalHeaders != nil {
		for k, v := range dApp.globalHeaders {
			out = append(out, k, v)
		}
	}
	if in != nil {
		for k, v := range *in {
			out = append(out, k, v)
		}
	}
	return out
}

func fixHTTPPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}


func makeResult(logger *log.Logger, result *httpclient.Result) (map[string]interface{}, error) {
	r := map[string]interface{}{}
	if result.Error != nil {
		logger.Error(result.Error.Error())
		return nil, result.Error
	}

	if result.Response != nil {
		headers := map[string]string{}
		for k, v := range result.Response.Header {
			if len(v) == 1 {
				headers[k] = v[0]
			} else {
				headers[k] = strings.Join(v, " ")
			}
		}
		r["headers"] = headers
		r["statusCode"] = result.Response.StatusCode
		r["statusMessage"] = result.Response.Status

		if strings.Contains(result.Response.Header.Get("Content-Type"), "application/json") {
			u.UnJson(result.String(), &r)
		} else {
			r["result"] = result.String()
		}
	}
	return r, nil
}
