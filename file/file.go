package file

import (
	"bufio"
	"errors"
	"github.com/api-go/plugin"
	"github.com/ssgo/u"
	"os"
	"strings"
	"sync"
)

var allowPaths = make([]string, 0)
var allowExtensions = make([]string, 0)
var notAllowMessage = ""
var fileConfigLock = sync.RWMutex{}

var lockFile = func(f *os.File) {}
var unlockFile = func(f *os.File) {}

type File struct {
	name string
	fd   *os.File
}

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api-go/plugins/file",
		Name: "file",
		ConfigSet: []plugin.ConfigSet{
			{Name: "allowPaths", Type: "[]string", Memo: "允许操作的文件路径"},
			{Name: "allowExtensions", Type: "[]string", Memo: "允许操作的文件后缀，以.开头，例如 .json .txt .db"},
			{Name: "notAllowMessage", Type: "string", Memo: "当文件路径或文件后缀不被允许时返回的错误信息"},
		},
		Init: func(conf map[string]interface{}) {
			newAllowPaths := make([]string, 0)
			newAllowExtensions := make([]string, 0)
			newNotAllowMessage := "file not allow to access"
			if conf["allowPaths"] != nil {
				u.Convert(conf["allowPaths"], &newAllowPaths)
			}
			if conf["allowExtensions"] != nil {
				u.Convert(conf["allowExtensions"], &newAllowExtensions)
			}
			if conf["notAllowMessage"] != nil {
				newNotAllowMessage = u.String(conf["notAllowMessage"])
			}
			fileConfigLock.Lock()
			allowPaths = newAllowPaths
			allowExtensions = newAllowExtensions
			notAllowMessage = newNotAllowMessage
			fileConfigLock.Unlock()
		},
		Objects: map[string]interface{}{
			// read 读取一个文件
			// * fileName 文件名
			// read return 文件内容，字符串格式
			"read": func(fileName string) (string, error) {
				if f, err := openFileForRead(fileName); err == nil {
					defer f.Close()
					return f.ReadAll()
				}else{
					return "", err
				}
			},
			// readBytes 读取一个二进制文件
			// readBytes return 文件内容，二进制格式
			"readBytes": func(fileName string) ([]byte, error) {
				if f, err := openFileForRead(fileName); err == nil {
					defer f.Close()
					return f.ReadAllBytes()
				}else{
					return nil, err
				}
			},
			// readFileLines 按行读取文件
			// readFileLines return 文件内容，返回字符串数组
			"readFileLines": func(fileName string) ([]string, error) {
				if f, err := openFileForRead(fileName); err == nil {
					defer f.Close()
					return f.ReadLines()
				}else{
					return nil, err
				}
			},
			// write 写入一个文件
			// write content 文件内容，字符串格式
			// write return 写入的字节数
			"write": func(fileName, content string) (int, error) {
				if f, err := openFileForWrite(fileName); err == nil {
					defer f.Close()
					return f.Write(content)
				}else{
					return 0, err
				}
			},
			// writeBytes 写入一个二进制文件
			// writeBytes content 文件内容，二进制格式
			// writeBytes return 写入的字节数
			"writeBytes": func(fileName string, content []byte) (int, error) {
				if f, err := openFileForWrite(fileName); err == nil {
					defer f.Close()
					return f.WriteBytes(content)
				}else{
					return 0, err
				}
			},
			"openForRead": openFileForRead,
			"openForWrite": openFileForWrite,
			"openForAppend": openFileForAppend,
			"open": openFile,
			// remove 删除文件
			"remove": func(fileName string) error {
				if !checkFileAllow(fileName) {
					return errors.New(getNotAllowMessage())
				}
				return os.Remove(fileName)
			},
			// rename 修改文件名
			// * fileOldName 旧文件
			// * fileNewName 新文件
			"rename": func(fileOldName, fileNewName string) error {
				if !checkFileAllow(fileOldName) || !checkFileAllow(fileNewName) {
					return errors.New(getNotAllowMessage())
				}
				return os.Rename(fileOldName, fileNewName)
			},
			// copy 复制文件
			"copy": func(fileOldName, fileNewName string) error {
				if !checkFileAllow(fileOldName) || !checkFileAllow(fileNewName) {
					return errors.New(getNotAllowMessage())
				}
				if buf, err := u.ReadFileBytes(fileOldName); err == nil {
					return u.WriteFileBytes(fileNewName, buf)
				} else {
					return err
				}
			},
			// saveJson 将对象存储为JSON格式的文件
			"saveJson": func(fileName string, content interface{}) error {
				if !checkFileAllow(fileName) {
					return errors.New(getNotAllowMessage())
				}
				return u.SaveJsonP(fileName, content)
			},
			// saveYaml 将对象存储为YAML格式的文件
			"saveYaml": func(fileName string, content interface{}) error {
				if !checkFileAllow(fileName) {
					return errors.New(getNotAllowMessage())
				}
				return u.SaveYaml(fileName, content)
			},
			// loadJson 读取JSON格式的文件并转化为对象
			// loadJson return 对象
			"loadJson": func(fileName string) (interface{}, error) {
				if !checkFileAllow(fileName) {
					return nil, errors.New(getNotAllowMessage())
				}
				var data interface{}
				err := u.LoadJson(fileName, &data)
				return data, err
			},
			// loadYaml 读取YAML格式的文件并转化为对象
			// loadYaml return 对象
			"loadYaml": func(fileName string) (interface{}, error) {
				if !checkFileAllow(fileName) {
					return nil, errors.New(getNotAllowMessage())
				}
				var data interface{}
				err := u.LoadYaml(fileName, &data)
				return data, err
			},
		},
	})
}

// openFileForRead 打开一个用于读取的文件，若不存在会抛出异常
// openFileForRead return 文件对象，请务必在使用完成后关闭文件
func openFileForRead(fileName string) (*File, error) {
	return _openFile(fileName, os.O_RDONLY, 0400)
}

// openFileForWrite 打开一个用于写入的文件，若不存在会自动创建
// openFileForWrite return 文件对象，请务必在使用完成后关闭文件
func openFileForWrite(fileName string) (*File, error) {
	return _openFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
}

// openFileForAppend 打开一个用于追加写入的文件，若不存在会自动创建
// openFileForAppend return 文件对象，请务必在使用完成后关闭文件
func openFileForAppend(fileName string) (*File, error) {
	return _openFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
}

// openFile 打开一个用于追加写入的文件，若不存在会自动创建
// openFile return 文件对象，请务必在使用完成后关闭文件
func openFile(fileName string) (*File, error) {
	return _openFile(fileName, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0600)
}
func _openFile(fileName string, flag int, perm os.FileMode) (*File, error) {
	if !checkFileAllow(fileName) {
		return nil, errors.New(getNotAllowMessage())
	}
	u.CheckPath(fileName)
	fd, err := os.OpenFile(fileName, flag, perm)
	if err != nil {
		return nil, err
	}
	lockFile(fd)
	return &File{name: fileName, fd: fd}, nil
}

// Close 关闭文件
func (f *File) Close() error {
	unlockFile(f.fd)
	return f.fd.Close()
}

// Read 从文件中读取指定长度的内容
// * size 长度
// Read return 读取的内容，字符串格式
func (f *File) Read(size int) (string, error) {
	str, err := f.ReadBytes(size)
	return string(str), err
}

// ReadBytes 从二进制文件中读取指定长度的内容
// ReadBytes return 读取的内容，二进制格式
func (f *File) ReadBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := f.fd.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[0:n], nil
}

// ReadAll 从文件中读取全部内容
// ReadAll return 读取的内容，字符串格式
func (f *File) ReadAll() (string, error) {
	str, err := f.ReadAllBytes()
	return string(str), err
}

// ReadAllBytes 从二进制文件中读取全部内容
// ReadAllBytes return 读取的内容，二进制格式
func (f *File) ReadAllBytes() ([]byte, error) {
	var maxLen int
	if fi, _ := os.Stat(f.name); fi != nil {
		maxLen = int(fi.Size())
	} else {
		maxLen = 1024000
	}
	return f.ReadBytes(maxLen)
}

// ReadLines 逐行文件中读取全部内容
// ReadLines return 读取的内容，字符串数组格式
func (f *File) ReadLines() ([]string, error) {
	outs := make([]string, 0)
	inputReader := bufio.NewReader(f.fd)
	for {
		line, err := inputReader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		outs = append(outs, line)
		if err != nil {
			break
		}
	}
	return outs, nil
}

// Write 写入字符串
// Write return 写入的长度
func (f *File) Write(content string) (int, error) {
	return f.WriteBytes([]byte(content))
}

// WriteBytes 写入二进制
// WriteBytes return 写入的长度
func (f *File) WriteBytes(content []byte) (int, error) {
	return f.fd.Write(content)
}

// SeekStart 将文件指针移动到开头
func (f *File) SeekStart() error {
	_, err := f.fd.Seek(0, 0)
	return err
}

// SeekEnd 将文件指针移动到末尾
func (f *File) SeekEnd() error {
	_, err := f.fd.Seek(0, 2)
	return err
}

// Seek 将文件指针移动到指定位置（从文件开头计算）
func (f *File) Seek(offset int64) error {
	_, err := f.fd.Seek(offset, 0)
	return err
}

func getNotAllowMessage() string {
	fileConfigLock.RLock()
	defer fileConfigLock.RUnlock()
	return notAllowMessage
}

func checkFileAllow(filename string) bool {
	fileConfigLock.RLock()
	defer fileConfigLock.RUnlock()
	if len(allowPaths) > 0 {
		ok := false
		for _, allowPath := range allowPaths {
			if strings.HasPrefix(filename, allowPath) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(allowExtensions) > 0 {
		ok := false
		for _, allowExtension := range allowExtensions {
			if strings.HasSuffix(filename, allowExtension) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}
