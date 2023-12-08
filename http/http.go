package http

import (
	"github.com/api-go/plugin"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"strings"
	"time"
)

type Client struct {
	pool          *httpclient.ClientPool
	baseURL       string
	globalHeaders map[string]string
}

var defaultClient = Client{
	pool:          httpclient.GetClient(time.Duration(0) * time.Second),
	globalHeaders: map[string]string{},
}

func init() {
	defaultClient.pool.EnableRedirect()
	plugin.Register(plugin.Plugin{
		Id:   "http",
		Name: "HTTP客户端",
		Objects: map[string]interface{}{
			"new":                   NewHTTP,
			"newH2C":                NewH2CHTTP,
			"newWithoutRedirect":    NewHTTPWithoutRedirect,
			"newH2CWithoutRedirect": NewH2CHTTPWithoutRedirect,
			"setBaseURL":            defaultClient.SetBaseURL,
			"SetGlobalHeaders":      defaultClient.SetGlobalHeaders,
			"get":                   defaultClient.Get,
			"post":                  defaultClient.Post,
			"put":                   defaultClient.Put,
			"delete":                defaultClient.Delete,
			"head":                  defaultClient.Head,
		},
	})
}

// SetBaseURL 设置一个URL前缀，后续请求中可以只提供path部分
// SetBaseURL url 以http://或https://开头的URL地址
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetGlobalHeaders 设置固定的HTTP头部信息，在每个请求中都加入这些HTTP头
// SetGlobalHeaders headers 传入一个Key-Value对象的HTTP头信息
func (c *Client) SetGlobalHeaders(headers map[string]string) {
	c.globalHeaders = headers
}

// Get 发送GET请求
// * url 以http://或https://开头的URL地址，如果设置了baseURL可以只提供path部分
// * headers 传入一个Key-Value对象的HTTP头信息，如果不指定头信息这个参数可以省略不传
// * return 返回结果对象，如果返回值是JSON格式，将自动转化为对象否则将字符串放在.result中，如发生错误将抛出异常，返回的对象中还包括：headers、statusCode、statusMessage
func (c *Client) Get(logger *log.Logger, url string, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(logger, c.pool.Get(c.makeURL(url), c.makeHeaderArray(headers)...))
}

// Post 发送POST请求
// * body 可以传入任意类型，如果不是字符串或二进制数组时会自动添加application/json头，数据将以json格式发送
func (c *Client) Post(logger *log.Logger, url string, body interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(logger, c.pool.Post(c.makeURL(url), body, c.makeHeaderArray(headers)...))
}

// Put 发送PUT请求
func (c *Client) Put(logger *log.Logger, url string, body interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(logger, c.pool.Put(c.makeURL(url), body, c.makeHeaderArray(headers)...))
}

// Delete 发送DELETE请求
func (c *Client) Delete(logger *log.Logger, url string, body interface{}, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(logger, c.pool.Delete(c.makeURL(url), body, c.makeHeaderArray(headers)...))
}

// Head 发送HEAD请求
func (c *Client) Head(logger *log.Logger, url string, headers *map[string]string) (map[string]interface{}, error) {
	return makeResult(logger, c.pool.Head(c.makeURL(url), c.makeHeaderArray(headers)...))
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

func (c *Client) makeURL(url string) string {
	if !strings.Contains(url, "://") && c.baseURL != "" {
		if strings.HasSuffix(c.baseURL, "/") && strings.HasPrefix(url, "/") {
			return c.baseURL + url[1:]
		} else if !strings.HasSuffix(c.baseURL, "/") && !strings.HasPrefix(url, "/") {
			return c.baseURL + "/" + url
		}
		return c.baseURL + url
	}
	return url
}

func (c *Client) makeHeaderArray(in *map[string]string) []string {
	out := make([]string, 0)
	if c.globalHeaders != nil {
		for k, v := range c.globalHeaders {
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

// NewHTTP 创建新的HTTP客户端
// * timeout 请求的超时时间，单位(毫秒)
// NewHTTP return HTTP客户端对象
func NewHTTP(timeout int) *Client {
	pool := httpclient.GetClient(time.Duration(timeout) * time.Second)
	pool.EnableRedirect()
	return &Client{
		pool:          pool,
		globalHeaders: map[string]string{},
	}
}

// NewHTTPWithoutRedirect 创建新的HTTP客户端（不自动跟踪301和302跳转）
// * timeout 请求的超时时间，单位(毫秒)
// NewHTTPWithoutRedirect return HTTP客户端对象
func NewHTTPWithoutRedirect(timeout int) *Client {
	pool := httpclient.GetClient(time.Duration(timeout) * time.Second)
	return &Client{
		pool:          pool,
		globalHeaders: map[string]string{},
	}
}

// NewH2CHTTP 创建新的H2C客户端
// NewH2CHTTP return H2C客户端对象
func NewH2CHTTP(timeout int) *Client {
	pool := httpclient.GetClientH2C(time.Duration(timeout) * time.Second)
	pool.EnableRedirect()
	return &Client{
		pool:          pool,
		globalHeaders: map[string]string{},
	}
}

// NewH2CHTTPWithoutRedirect 创建新的H2C客户端（不自动跟踪301和302跳转）
// NewH2CHTTPWithoutRedirect return H2C客户端对象
func NewH2CHTTPWithoutRedirect(timeout int) *Client {
	pool := httpclient.GetClientH2C(time.Duration(timeout) * time.Second)
	return &Client{
		pool:          pool,
		globalHeaders: map[string]string{},
	}
}
