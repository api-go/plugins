package util

import (
	"encoding/hex"
	"github.com/api-go/plugin"
	"github.com/ssgo/u"
)

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "util",
		Name: "基础工具",
		JsCode: `
util.formatDate = function (date, fmt, timezone) {
    let ret
    if (!date) return ''
    if (!fmt) fmt = 'YYYY-mm-dd HH:MM:SS'
    const opt = {
        'YYYY': date.getFullYear().toString(),
        'YY': date.getFullYear().toString().substring(2),
        'm+': (date.getMonth() + 1).toString(),
        'd+': date.getDate().toString(),
        'H+': date.getHours().toString(),
        'h+': (date.getHours() % 12).toString(),
        'APM': (date.getHours() > 12 ? 'PM' : 'AM'),
        'M+': date.getMinutes().toString(),
        'S+': date.getSeconds().toString(),
        'YW': 'W' + Math.ceil(((date.getTime() - new Date(date.getFullYear(), 0, 0))) / (24 * 60 * 60 * 1000) / 7),
    }
    for (let k in opt) {
        ret = new RegExp('(' + k + ')').exec(fmt)
        if (ret) {
            fmt = fmt.replace(ret[1], (ret[1].length == 1) ? (opt[k]) : (opt[k].padStart(ret[1].length, '0')))
        }
    }
    return fmt
}

util.parseDate = function (str) {
    if (typeof str === 'number') str = str + ''
    if (/^\d+[\\.]?\d*$/.test(str)) {
        if (str.length === 10) str += '000'
        return new Date(util.int(str))
    }
    if (str.length === 19) {
        return new Date(parseInt(str.substr(0, 4)), parseInt(str.substr(5, 2)) - 1, parseInt(str.substr(8, 2)), parseInt(str.substr(11, 2)), parseInt(str.substr(14, 2)), parseInt(str.substr(17, 2)))
    } else if (str.length === 10) {
        return new Date(parseInt(str.substr(0, 4)), parseInt(str.substr(5, 2)) - 1, parseInt(str.substr(8, 2)))
    } else if (str.length === 7) {
        return new Date(parseInt(str.substr(0, 4)), parseInt(str.substr(5, 2)) - 1, 1)
    } else if (str.length === 8) {
        let date = new Date()
        return new Date(date.getFullYear(), date.getMonth(), date.getDate(), parseInt(str.substr(0, 2)), parseInt(str.substr(3, 2)), parseInt(str.substr(6, 2)))
    } else if (str.length === 5) {
        let date = new Date()
        return new Date(date.getFullYear(), date.getMonth(), date.getDate(), parseInt(str.substr(0, 2)), parseInt(str.substr(3, 2)), 0)
    }
    return new Date(str)
}

util.int = function (v) {
    if (!v) return 0
    if (typeof v === 'number' || v instanceof Number) return v
    if (typeof v === 'object') v = v.toString ? v.toString() : ''
    if (typeof v !== 'string') return 0
    try {
        return Math.round(parseFloat(v.replace(/,/g, '').trim())) || 0
    } catch (e) {
        return 0
    }
}

util.float = function (v) {
    if (!v) return 0.0
    if (typeof v === 'number' || v instanceof Number) return v
    if (typeof v === 'object') v = v.toString ? v.toString() : ''
    if (typeof v !== 'string') return 0.0
    try {
        return parseFloat(v.replace(/,/g, '').trim()) || 0.0
    } catch (e) {
        return 0.0
    }
}

util.str = function (v) {
    if (!v) return ''
    if (typeof v === 'string') return v
    if (v.toString) return v.toString()
    return ''
}

util.keysBy = function (obj, ...fieldAndValues) {
    let keys = []
    for (let k in obj) {
        let match = true
        if (fieldAndValues.length === 1) {
            // 查找一位数组
            if (obj[k] != fieldAndValues[0]) {
                match = false
            }
        } else {
            // 查找二维数组
            for (let i = 0; i < fieldAndValues.length; i += 2) {
                if (obj[k][fieldAndValues[i]] != fieldAndValues[i + 1]) {
                    match = false
                    break
                }
            }
        }
        if (match) {
            keys.push(k)
        }
    }
    return keys
}

util.listBy = function (obj, ...fieldAndValues) {
    let list = obj instanceof Array || obj instanceof NodeList ? [] : {}
    let keys = util.keysBy(obj, ...fieldAndValues)
    for (let k of keys) {
        if (obj instanceof Array || obj instanceof NodeList) {
            list.push(obj[k])
        } else {
            list[k] = obj[k]
        }
    }
    return list
}

util.hasBy = function (obj, ...fieldAndValues) {
    let keys = util.keysBy(obj, ...fieldAndValues)
    return keys.length > 0
}

util.getBy = function (obj, ...fieldAndValues) {
    let keys = util.keysBy(obj, ...fieldAndValues)
    if (keys.length > 0) return obj[keys[0]]
    return null
}

util.setBy = function (obj, value, ...fieldAndValues) {
    let keys = util.keysBy(obj, ...fieldAndValues)
    if (keys.length > 0) obj[keys[0]] = value
}

util.indexBy = function (obj, ...fieldAndValues) {
    let keys = util.keysBy(obj, ...fieldAndValues)
    if (keys.length > 0) {
        return obj instanceof Array || obj instanceof NodeList ? util.int(keys[0]) : keys[0]
    }
    return -1
}

util.removeBy = function (obj, ...fieldAndValues) {
    let keys = util.keysBy(obj, ...fieldAndValues)
    let n = 0
    for (let i = keys.length - 1; i >= 0; i--) {
        let k = keys[i]
        if (obj instanceof Array || obj instanceof NodeList) {
            obj.splice(k, 1)
        } else {
            delete obj[k]
        }
        n++
    }
    return n
}

util.removeArrayItem = function (list, item) {
    let pos = list.indexOf(item)
    if (pos !== -1) list.splice(pos, 1)
}

util.last = function (arr) {
    if (arr && arr.length) {
        return arr[arr.length - 1]
    }
    return null
}

util.len = function (obj) {
    if (obj instanceof Array || obj instanceof NodeList) {
        return obj.length
    } else {
        let n = 0
        for (let k in obj) n++
        return n
    }
}

util.mergeBy = function (olds, news, ...fields) {
    if (!olds) return news
    for (let newItem of news) {
        let fieldAndValues = []
        for (let field of fields) {
            fieldAndValues.push(field, newItem[field])
        }
        let oldIndex = util.indexBy(olds, ...fieldAndValues)
        if (oldIndex === -1) {
            olds.push(newItem)
        } else {
            olds[oldIndex] = newItem
        }
    }
    return olds
}

util.sortBy = function (obj, field, isReverse = false, sortType) {
    let list = obj instanceof Array || obj instanceof NodeList ? [] : {}
    let sortedKeys = {}
    let sortArr = []
    for (let k in obj) {
        let v = ''
        if (field instanceof Array) {
            for (let f of field) v += obj[k][f]
        } else {
            v = obj[k][field]
        }
        if (!sortedKeys[v]) {
            sortedKeys[v] = true
            sortArr.push(v)
        }
    }
    sortArr.sort((a, b) => {
        if(sortType === 'int'){
            a = util.int(a)
            b = util.int(b)
        } else if(sortType === 'float'){
            a = util.float(a)
            b = util.float(b)
        }
        if (a == b) return 0
        if (typeof a === 'number' && typeof b === 'number') {
            return isReverse ? b - a : a - b
        } else {
            return (isReverse ? a < b : a > b) ? 1 : -1
        }
    })
    for (let sortKey of sortArr) {
        for (let k in obj) {
            let v = ''
            if (field instanceof Array) {
                for (let f of field) v += obj[k][f]
            } else {
                v = obj[k][field]
            }

            if (obj instanceof Array || obj instanceof NodeList) {
                if (v == sortKey) list.push(obj[k])
            } else {
                if (v == sortKey) list[k] = obj[k]
            }
        }
    }
    return list
}

util.in = function (v1, v2) {
    if (!(v1 instanceof Array)) v1 = util.split(v1, /,\s*/)
    return v1.indexOf(String(v2)) !== -1
}

util.uniquePush = function (arr, ...values) {
    for (let v of values) {
        if (arr.indexOf(v) === -1) arr.push(v)
    }
}

util.clearEmpty = function (arr) {
    let a = []
    for (let v of arr) if (v) a.push(v)
    return a
}

util.split = function (v1, v2) {
    return util.clearEmpty(util.str(v1).split(v2))
}

util.join = function (arr, separator) {
    return util.clearEmpty(arr).join(separator)
}

util.copy = function (obj, isDeepCopy) {
    let newObj
    if (obj instanceof Array || obj instanceof NodeList) {
        newObj = []
        for (let o of obj) {
            if (isDeepCopy && typeof o === 'object' && o) o = util.copy(o)
            newObj.push(o)
        }
    } else {
        newObj = {}
        for (let k in obj) {
            let v = obj[k]
            if (isDeepCopy && typeof v === 'object' && v) v = util.copy(v)
            newObj[k] = v
        }
    }

    return newObj
}

`,
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

			// makeId 生成指定长度的随机ID
			// makeId size ID长度(6~20)
			// makeId return 二进制数据
			"makeId": func(size int) string {
				if size > 20 {
					return u.UniqueId()
				} else if size > 14 {
					return u.UniqueId()[0:size]
				} else if size > 12 {
					return u.ShortUniqueId()[0:size]
				} else if size > 10 {
					return u.Id12()[0:size]
				} else if size > 8 {
					return u.Id10()[0:size]
				} else if size >= 6 {
					return u.Id8()[0:size]
				} else {
					return u.Id6()
				}
			},
		},
	})
}
