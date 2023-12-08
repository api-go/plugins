package file

import (
	"bufio"
	"errors"
	"github.com/api-go/plugin"
	"github.com/ssgo/u"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var _allowPaths = make([]string, 0)
var _allowExtensions = make([]string, 0)
var notAllowMessage = ""
var fileConfigLock = sync.RWMutex{}

var lockFile = func(f *os.File) {}
var unlockFile = func(f *os.File) {}

type File struct {
	name string
	fd   *os.File
}
type FileInfo struct {
	Name  string
	Mtime int64
	IsDir bool
	Size  int64
}

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "file",
		Name: "文件操作",
		ConfigSet: []plugin.ConfigSet{
			{Name: "_allowPaths", Type: "[]string", Memo: "允许操作的文件路径"},
			{Name: "_allowExtensions", Type: "[]string", Memo: "允许操作的文件后缀，以.开头，例如 .json .txt .db"},
			{Name: "notAllowMessage", Type: "string", Memo: "当文件路径或文件后缀不被允许时返回的错误信息"},
		},
		Init: func(conf map[string]interface{}) {
			newAllowPaths := make([]string, 0)
			newAllowExtensions := make([]string, 0)
			newNotAllowMessage := "file not allow to access"
			if conf["_allowPaths"] != nil {
				u.Convert(conf["_allowPaths"], &newAllowPaths)
			}
			if conf["_allowExtensions"] != nil {
				u.Convert(conf["_allowExtensions"], &newAllowExtensions)
			}
			if conf["notAllowMessage"] != nil {
				newNotAllowMessage = u.String(conf["notAllowMessage"])
			}
			fileConfigLock.Lock()
			_allowPaths = newAllowPaths
			_allowExtensions = newAllowExtensions
			notAllowMessage = newNotAllowMessage
			fileConfigLock.Unlock()
		},
		Objects: map[string]interface{}{
			// list 列出目录下的文件
			// * dirname 目录名称
			// list sortBy 排序依据[name|mtime|size]，默认使用名称排序
			// list limit 返回指定数量，默认返回全部
			// list return 文件列表{name:文件名,mtime:最后修改时间,size:文件尺寸}
			"list": func(dirname string, sortBy *string, limit *int) ([]FileInfo, error) {
				dirname = fixPath(dirname)
				if !checkDirAllow(dirname) {
					return nil, errors.New(getNotAllowMessage())
				}
				u.CheckPath(path.Join(dirname, "_"))

				if d, err := os.Open(dirname); err == nil {
					out := make([]FileInfo, 0)
					if files, err := d.Readdir(-1); err == nil {
						for _, f := range files {
							if !strings.HasPrefix(f.Name(), ".") {
								out = append(out, FileInfo{
									Name:  f.Name(),
									Mtime: f.ModTime().Unix(),
									IsDir: f.IsDir(),
									Size:  f.Size(),
								})
							}
						}
					}
					_ = d.Close()
					if sortBy != nil {
						sort.Slice(out, func(i, j int) bool {
							if *sortBy == "mtime" {
								return out[i].Mtime > out[j].Mtime
							} else if *sortBy == "size" {
								return out[i].Size > out[j].Size
							} else {
								return out[i].Name < out[j].Name
							}
						})
					}
					if limit != nil && *limit > 0 && *limit < len(out) {
						return out[0:*limit], nil
					}
					return out, nil
				} else {
					return []FileInfo{}, err
				}
			},
			// exists 判断文件是否存在
			// exists return 是否存在
			"exists": func(filename string) bool {
				fi, err := os.Stat(filename)
				return err == nil && fi != nil
			},
			// isDir 判断是否文件夹
			// isDir return 是否文件夹
			"isDir": func(filename string) (bool, error) {
				filename = fixPath(filename)
				if fi, err := getFileStat(filename); err == nil {
					return fi.IsDir(), nil
				} else {
					return false, err
				}
			},
			// getModTime 返回文件的修改时间（时间戳格式）
			// getModTime return 修改时间的时间戳
			"getModTime": func(filename string) (int64, error) {
				filename = fixPath(filename)
				if fi, err := getFileStat(filename); err == nil {
					return fi.ModTime().Unix(), nil
				} else {
					return 0, err
				}
			},
			// getModTimeStr 返回文件的修改时间（字符串格式）
			// getModTimeStr return 修改时间的字符串
			"getModTimeStr": func(filename string) (string, error) {
				filename = fixPath(filename)
				if fi, err := getFileStat(filename); err == nil {
					return fi.ModTime().Format("2006-01-02 15:04:05"), nil
				} else {
					return "", err
				}
			},
			// makeDir 创建文件夹
			"makeDir": func(dirname string) error {
				dirname = fixPath(dirname)
				if !checkDirAllow(dirname) {
					return errors.New(getNotAllowMessage())
				}
				return os.MkdirAll(dirname, 0700)
			},
			// read 读取一个文件
			// * filename 文件名
			// read return 文件内容，字符串格式
			"read": func(filename string) (string, error) {
				filename = fixPath(filename)
				if f, err := openFileForRead(filename); err == nil {
					defer f.Close()
					return f.ReadAll()
				} else {
					return "", err
				}
			},
			// readBytes 读取一个二进制文件
			// readBytes return 文件内容，二进制格式
			"readBytes": func(filename string) ([]byte, error) {
				filename = fixPath(filename)
				if f, err := openFileForRead(filename); err == nil {
					defer f.Close()
					return f.ReadAllBytes()
				} else {
					return nil, err
				}
			},
			// readFileLines 按行读取文件
			// readFileLines return 文件内容，返回字符串数组
			"readFileLines": func(filename string) ([]string, error) {
				filename = fixPath(filename)
				if f, err := openFileForRead(filename); err == nil {
					defer f.Close()
					return f.ReadLines()
				} else {
					return nil, err
				}
			},
			// write 写入一个文件
			// write content 文件内容，字符串格式
			// write return 写入的字节数
			"write": func(filename, content string) (int, error) {
				filename = fixPath(filename)
				if f, err := openFileForWrite(filename); err == nil {
					defer f.Close()
					return f.Write(content)
				} else {
					return 0, err
				}
			},
			// writeBytes 写入一个二进制文件
			// writeBytes content 文件内容，二进制格式
			// writeBytes return 写入的字节数
			"writeBytes": func(filename string, content []byte) (int, error) {
				filename = fixPath(filename)
				if f, err := openFileForWrite(filename); err == nil {
					defer f.Close()
					return f.WriteBytes(content)
				} else {
					return 0, err
				}
			},
			"openForRead":   openFileForRead,
			"openForWrite":  openFileForWrite,
			"openForAppend": openFileForAppend,
			"open":          openFile,
			// remove 删除文件
			"remove": func(filename string) error {
				filename = fixPath(filename)
				fi, err := getFileStat(filename)
				if err == nil && fi.IsDir() {
					if !checkDirAllow(filename) {
						return errors.New(getNotAllowMessage())
					}
					return os.RemoveAll(filename)
				} else {
					if !checkFileAllow(filename) {
						return errors.New(getNotAllowMessage())
					}
					return os.Remove(filename)
				}
			},
			// rename 修改文件名
			// * fileOldName 旧文件
			// * fileNewName 新文件
			"rename": func(fileOldName, fileNewName string) error {
				fi, err := getFileStat(fileOldName)
				if err == nil && fi.IsDir() {
					if !checkDirAllow(fileOldName) || !checkDirAllow(fileNewName) {
						return errors.New(getNotAllowMessage())
					}
				} else {
					if !checkFileAllow(fileOldName) || !checkFileAllow(fileNewName) {
						return errors.New(getNotAllowMessage())
					}
				}
				return os.Rename(fileOldName, fileNewName)
			},
			// copy 复制文件
			"copy": func(fileOldName, fileNewName string) error {
				fileNewName = fixPath(fileNewName)
				newFI, _ := getFileStat(fileNewName)
				fileOldName = fixPath(fileOldName)
				fi, err := getFileStat(fileOldName)
				if err == nil && fi.IsDir() {
					if !checkDirAllow(fileOldName) || !checkDirAllow(fileNewName) {
						return errors.New(getNotAllowMessage())
					}
					//if strings.HasSuffix(fileNewName, "/") {
					//	u.CheckPath(path.Join(fileNewName, "a.txt"))
					//}else{
					//	u.CheckPath(fileNewName)
					//}
					//_, err = u.RunCommand("cp", "-rf", fileOldName, fileNewName)
					//return err
					return copyDir(fileNewName, fileOldName)
				} else {
					if !checkFileAllow(fileOldName) || !checkFileAllow(fileNewName) {
						return errors.New(getNotAllowMessage())
					}
					if newFI != nil && newFI.IsDir() {
						fileNewName = path.Join(fileNewName, path.Base(fileOldName))
					}
					return copyFile(fileNewName, fileOldName)
				}
			},
			// saveJson 将对象存储为JSON格式的文件
			"saveJson": func(filename string, content interface{}) error {
				filename = fixPath(filename)
				if !checkFileAllow(filename) {
					return errors.New(getNotAllowMessage())
				}
				return u.SaveJsonP(filename, content)
			},
			// saveYaml 将对象存储为YAML格式的文件
			"saveYaml": func(filename string, content interface{}) error {
				filename = fixPath(filename)
				if !checkFileAllow(filename) {
					return errors.New(getNotAllowMessage())
				}
				return u.SaveYaml(filename, content)
			},
			// loadJson 读取JSON格式的文件并转化为对象
			// loadJson return 对象
			"loadJson": func(filename string) (interface{}, error) {
				filename = fixPath(filename)
				if !checkFileAllow(filename) {
					return nil, errors.New(getNotAllowMessage())
				}
				var data interface{}
				err := u.LoadJson(filename, &data)
				return data, err
			},
			// loadYaml 读取YAML格式的文件并转化为对象
			// loadYaml return 对象
			"loadYaml": func(filename string) (interface{}, error) {
				filename = fixPath(filename)
				if !checkFileAllow(filename) {
					return nil, errors.New(getNotAllowMessage())
				}
				var data interface{}
				buf, err := u.ReadFileBytes(filename)
				if err == nil {
					err = yaml.Unmarshal(buf, &data)
				}
				return data, err
			},
		},
	})
}

func copyDir(dst, src string) error{
	if d, err := os.Open(src); err == nil {
		defer d.Close()
		if files, err := d.Readdir(-1); err == nil {
			for _, f := range files {
				if f.IsDir() {
					if err2 := copyDir(path.Join(dst, f.Name()), path.Join(src, f.Name())); err2 != nil {
						return err2
					}
				}else{
					if err2 := copyFile(path.Join(dst, f.Name()), path.Join(src, f.Name())); err2 != nil {
						return err2
					}
				}
			}
		}
		return nil
	} else {
		return err
	}
}

func copyFile(dst, src string) error{
	if f, err := openFileForRead(src); err == nil {
		defer f.Close()
		if f2, err2 := openFileForWrite(dst); err == nil {
			defer f2.Close()
			for {
				if buf, err3 := f.ReadBytes(10240); err3 != nil {
					break
				} else {
					_, _ = f2.WriteBytes(buf)
				}
			}
			return nil
		} else {
			return err2
		}
	} else {
		return err
	}
}

func fixPath(filename string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(filename, "/", "\\")
	}
	return filename
}

func getFileStat(filename string) (os.FileInfo, error) {
	fi, err := os.Stat(filename)
	if err == nil && fi.IsDir() {
		if !checkDirAllow(filename) {
			return nil, errors.New(getNotAllowMessage())
		}
	} else {
		if !checkFileAllow(filename) {
			return nil, errors.New(getNotAllowMessage())
		}
	}
	return fi, err
}

// openFileForRead 打开一个用于读取的文件，若不存在会抛出异常
// openFileForRead return 文件对象，请务必在使用完成后关闭文件
func openFileForRead(filename string) (*File, error) {
	return _openFile(filename, os.O_RDONLY, 0400)
}

// openFileForWrite 打开一个用于写入的文件，若不存在会自动创建
// openFileForWrite return 文件对象，请务必在使用完成后关闭文件
func openFileForWrite(filename string) (*File, error) {
	return _openFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
}

// openFileForAppend 打开一个用于追加写入的文件，若不存在会自动创建
// openFileForAppend return 文件对象，请务必在使用完成后关闭文件
func openFileForAppend(filename string) (*File, error) {
	return _openFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
}

// openFile 打开一个用于追加写入的文件，若不存在会自动创建
// openFile return 文件对象，请务必在使用完成后关闭文件
func openFile(filename string) (*File, error) {
	return _openFile(filename, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0600)
}
func _openFile(filename string, flag int, perm os.FileMode) (*File, error) {
	if !checkFileAllow(filename) {
		return nil, errors.New(getNotAllowMessage())
	}
	u.CheckPath(filename)
	fd, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	lockFile(fd)
	return &File{name: filename, fd: fd}, nil
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

func getAllowPaths() []string {
	fileConfigLock.RLock()
	defer fileConfigLock.RUnlock()
	allowPaths := make([]string, len(_allowPaths))
	for i, v := range _allowPaths {
		allowPaths[i] = v
	}
	return allowPaths
}

func getAllowExtensions() []string {
	fileConfigLock.RLock()
	defer fileConfigLock.RUnlock()
	allowExtensions := make([]string, len(_allowExtensions))
	for i, v := range _allowPaths {
		allowExtensions[i] = v
	}
	return allowExtensions
}

func checkDirAllow(filename string) bool {
	allowPaths := getAllowPaths()
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
	return true
}

func checkFileAllow(filename string) bool {
	if !checkDirAllow(filename) {
		return false
	}

	allowExtensions := getAllowExtensions()
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
