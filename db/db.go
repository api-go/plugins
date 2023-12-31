package db

import (
	"fmt"
	"github.com/api-go/plugin"
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
		Id:   "db",
		Name: "数据库操作",
		ConfigSample: `default: mysql://root:<**encrypted_password**>@127.0.0.1:3306/1?maxIdles=0&maxLifeTime=0&maxOpens=0&logSlow=1s # set default db connection pool, used by db.xxx
configs:
  conn1: sqlite3://conn1.db # set a named connection pool, used by db.get('conn1').xxx
  conn2: mysql://root:@127.0.0.1:3306/1?sslCa=<**encrypted**>&sslCert=<**encrypted**>&sslKey=<**encrypted**>&sslSkipVerify=true # set ssl connection pool for mysql
  conn3: mysql://root:@127.0.0.1:3306/1?timeout=90s&readTimeout=5s&writeTimeout=3s&charset=utf8mb4,utf8 # set more option for mysql
`,

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
		// 实现直接使用db.xxx操作默认的数据库
		JsCode: `_db = db
db = _db.fetch()
db.fetch = _db.fetch
`,
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

// Query1 查询
// Query1 return 返回查询到的第一行数据，对象格式
func (db *DB) Query1(requestSql string, args ...interface{}) (map[string]interface{}, error) {
	r := db.pool.Query(requestSql, args...)
	results := r.MapResults()
	if len(results) > 0 {
		return results[0], r.Error
	} else {
		return map[string]interface{}{}, r.Error
	}
}

// Query11 查询
// Query11 return 返回查询到的第一行第一列数据，字段类型对应的格式
func (db *DB) Query11(requestSql string, args ...interface{}) (interface{}, error) {
	r := db.pool.Query(requestSql, args...)
	results := r.SliceResults()
	if len(results) > 0 {
		if len(results[0]) > 0 {
			return results[0][0], r.Error
		} else {
			return nil, r.Error
		}
	} else {
		return nil, r.Error
	}
}

// Query1a 查询
// Query1a return 返回查询到的第一列数据，数组格式
func (db *DB) Query1a(requestSql string, args ...interface{}) ([]interface{}, error) {
	r := db.pool.Query(requestSql, args...)
	results := r.SliceResults()
	a := make([]interface{}, 0)
	for _, row := range results {
		if len(results[0]) > 0 {
			a = append(a, row[0])
		}
	}
	return a, r.Error
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

// MakeId 生成指定字段不唯一的ID
// MakeId idField ID字段
// MakeId idSize ID长度
// MakeId return 新的ID
func (db *DB) MakeId(table string, idField string, idSize uint) (string, error) {
	var id string
	var err error
	for i:=0; i<100; i++ {
		if idSize > 20 {
			id = u.UniqueId()
		} else if idSize > 14 {
			id = u.UniqueId()[0:idSize]
		} else if idSize > 12 {
			id = u.ShortUniqueId()[0:idSize]
		} else if idSize > 10 {
			id = u.Id12()[0:idSize]
		} else if idSize > 8 {
			id = u.Id10()[0:idSize]
		} else if idSize >= 6 {
			id = u.Id8()[0:idSize]
		} else {
			id = u.Id6()
		}
		r := db.pool.Query(fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s`=?", table, idField), id)
		err = r.Error
		if r.IntOnR1C1() == 0 {
			break
		}
	}
	return id, err
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

func (tx *Tx) Query1(requestSql string, args ...interface{}) (map[string]interface{}, error) {
	r := tx.conn.Query(requestSql, args...)
	results := r.MapResults()
	if len(results) > 0 {
		return results[0], r.Error
	} else {
		return map[string]interface{}{}, r.Error
	}
}

func (tx *Tx) Query11(requestSql string, args ...interface{}) (interface{}, error) {
	r := tx.conn.Query(requestSql, args...)
	results := r.SliceResults()
	if len(results) > 0 {
		if len(results[0]) > 0 {
			return results[0][0], r.Error
		} else {
			return nil, r.Error
		}
	} else {
		return nil, r.Error
	}
}

func (tx *Tx) Query1a(requestSql string, args ...interface{}) ([]interface{}, error) {
	r := tx.conn.Query(requestSql, args...)
	results := r.SliceResults()
	a := make([]interface{}, 0)
	for _, row := range results {
		if len(results[0]) > 0 {
			a = append(a, row[0])
		}
	}
	return a, r.Error
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
