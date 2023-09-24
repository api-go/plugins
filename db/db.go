package db

import (
	"github.com/api0-work/plugin"
	"github.com/ssgo/db"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
)

type DB struct {
	pool *db.DB
}

type Tx struct {
	conn *db.Tx
}

var dbPool = map[string]*db.DB{}
var defaultDB *db.DB

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api0-work/plugins/db",
		Name: "db",
		ConfigSet: []plugin.ConfigSet{
			{Name: "default", Type: "string", Memo: "默认的DB连接，使用 db.get() 来获得实例，格式为 db://127.0.0.1:3306/1 或 db://root:<**加密的密码**>@127.0.0.1:3306?maxIdles=0&maxLifeTime=0&maxOpens=0&logSlow=1s"},
			{Name: "configs", Type: "map[string]string", Memo: "其他DB连接，使用 db.get('name') 来获得实例"},
		},
		Init: func(conf map[string]interface{}) {
			if conf["default"] != nil {
				defaultDB = db.GetDB(u.String(conf["default"]), nil)
			}
			if conf["configs"] != nil {
				confs := map[string]string{}
				u.Convert(conf["configs"], &confs)
				for name, url := range confs {
					dbPool[name] = db.GetDB(url, nil)
				}
			}
		},
		Objects: map[string]interface{}{
			"fetch": GetDB,
		},
	})
}

// GetDB 获得数据库连接
// GetDB name 连接配置名称，如果不提供名称则使用默认连接
// GetDB return 数据库连接，对象内置连接池操作，完成后无需手动关闭连接
func GetDB(name *string, logger *log.Logger) *DB {
	if name == nil || *name == "" {
		if defaultDB != nil {
			return &DB{
				pool: defaultDB.CopyByLogger(logger),
			}
		}
	} else {
		if dbPool[*name] != nil {
			return &DB{
				pool: dbPool[*name].CopyByLogger(logger),
			}
		}
	}
	return &DB{
		pool: db.GetDB("", logger),
	}
}

// Begin 开始事务
// Begin return 事务对象，事务中的操作都在事务对象上操作，请务必在返回的事务对象上执行commit或rollback
func (db *DB) Begin() *Tx {
	return db.Begin()
}

// Exec 执行SQL
// * requestSql SQL语句
// * args SQL语句中问号变量的值，按顺序放在请求参数中
// Exec return 如果是INSERT到含有自增字段的表中返回插入的自增ID，否则返回影响的行数
func (db *DB) Exec(requestSql string, args ...interface{}) (int64, error) {
	r := db.pool.Exec(requestSql, args...)
	out := r.Id()
	if out == 0 {
		out = r.Changes()
	}
	return out, r.Error
}

// Query 查询
// Query return 返回查询到的数据，对象数组格式
func (db *DB) Query(requestSql string, args ...interface{}) ([]map[string]interface{}, error) {
	r := db.pool.Query(requestSql, args...)
	return r.MapResults(), r.Error
}

// Insert 插入数据
// * table 表名
// * data 数据对象（Key-Value格式）
// Insert return 如果是INSERT到含有自增字段的表中返回插入的自增ID，否则返回影响的行数
func (db *DB) Insert(table string, data map[string]interface{}) (int64, error) {
	r := db.pool.Insert(table, data)
	out := r.Id()
	if out == 0 {
		out = r.Changes()
	}
	return out, r.Error
}

// Replace 替换数据
// Replace return 如果是REPLACE到含有自增字段的表中返回插入的自增ID，否则返回影响的行数
func (db *DB) Replace(table string, data map[string]interface{}) (int64, error) {
	r := db.pool.Replace(table, data)
	out := r.Id()
	if out == 0 {
		out = r.Changes()
	}
	return out, r.Error
}

// Update 更新数据
// * wheres 条件（SQL中WHERE后面的部分）
// Update return 返回影响的行数
func (db *DB) Update(table string, data map[string]interface{}, wheres string, args ...interface{}) (int64, error) {
	r := db.pool.Update(table, data, wheres, args...)
	return r.Changes(), r.Error
}

// Delete 删除数据
// Delete return 返回影响的行数
func (db *DB) Delete(table string, wheres string, args ...interface{}) (int64, error) {
	r := db.pool.Delete(table, wheres, args...)
	return r.Changes(), r.Error
}

// Commit 提交事务
func (tx *Tx) Commit() error {
	return tx.conn.Commit()
}

// Rollback 回滚事务
func (tx *Tx) Rollback() error {
	return tx.conn.Rollback()
}

// Finish 根据传入的成功标识提交或回滚事务
// Finish ok 事务是否执行成功
func (tx *Tx) Finish(ok bool) error {
	return tx.conn.Finish(ok)
}

// CheckFinished 检查事务是否已经提交或回滚，如果事务没有结束则执行回滚操作
func (tx *Tx) CheckFinished() error {
	return tx.conn.CheckFinished()
}

func (tx *Tx) Exec(requestSql string, args ...interface{}) (int64, error) {
	r := tx.conn.Exec(requestSql, args...)
	return r.Changes(), r.Error
}

func (tx *Tx) Query(requestSql string, args ...interface{}) ([]map[string]interface{}, error) {
	r := tx.conn.Query(requestSql, args...)
	return r.MapResults(), r.Error
}

func (tx *Tx) Insert(table string, data map[string]interface{}) (int64, error) {
	r := tx.conn.Insert(table, data)
	return r.Id(), r.Error
}

func (tx *Tx) Replace(table string, data map[string]interface{}) (int64, error) {
	r := tx.conn.Replace(table, data)
	return r.Id(), r.Error
}

func (tx *Tx) Update(table string, data map[string]interface{}, wheres string, args ...interface{}) (int64, error) {
	r := tx.conn.Update(table, data, wheres, args...)
	return r.Changes(), r.Error
}

func (tx *Tx) Delete(table string, wheres string, args ...interface{}) (int64, error) {
	r := tx.conn.Delete(table, wheres, args...)
	return r.Changes(), r.Error
}
