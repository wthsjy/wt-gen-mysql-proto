package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"strings"
	"time"
)

type DDLM struct {
	Field   string
	Type    string
	Comment string
	Key     string
}

var (
	err       error
	mDB       *gorm.DB
	ddlms     []DDLM
	host      string
	port      int
	username  string
	password  string
	dbName    string
	tableName string
)

func main() {
	flag.StringVar(&username, "u", "", "用户名")
	flag.StringVar(&password, "p", "", "密码")
	flag.StringVar(&host, "h", "127.0.0.1", "主机名,默认 127.0.0.1")
	flag.IntVar(&port, "P", 3306, "端口号")
	flag.StringVar(&dbName, "d", "", "数据库名")
	flag.StringVar(&tableName, "t", "", "表名")

	// 从arguments中解析注册的flag。必须在所有flag都注册好而未访问其值时执行。未注册却使用flag -help时，会返回ErrHelp。
	flag.Parse()

	dsn := getDSN()
	mDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = mDB.Raw(fmt.Sprintf("SHOW FULL FIELDS FROM  %s", tableName)).Find(&ddlms).Error
	if err != nil {
		panic(err)
	}

	var fields []string
	var priKeyField string
	protoStr := fmt.Sprintf("// mysql database.table: %s.%s\nmessage %s{\n", dbName, tableName, FirstUpCase(CamelCase(tableName)))
	for index, v := range ddlms {
		fields = append(fields, fmt.Sprintf("`%s`", v.Field))
		if strings.ToUpper(v.Key) == "PRI" {
			priKeyField = v.Field
		}
		if strings.TrimSpace(v.Comment) != "" {
			protoStr += fmt.Sprintf("  // %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\n  // "))
			protoStr += fmt.Sprintf("  //\n  // table field:\n  // @gotags: gorm:\"column:%s\"\n", v.Field)
		}
		protoStr += fmt.Sprintf("  %s %s = %d;\n", getProtoType(v.Type), CamelCase(v.Field), index+1)
	}
	protoStr += "}"

	protoStr += "\n\n"
	protoStr += fmt.Sprintf("insert into `%s`(%s)values(%s)\n", tableName, strings.Join(fields, ","), strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ","))
	protoStr += fmt.Sprintf("delete from `%s` where `%s` in ?\n", tableName, priKeyField)
	protoStr += fmt.Sprintf("select %s from `%s` where `%s` in ?\n", strings.Join(fields, ","), tableName, priKeyField)

	fmt.Println(protoStr)
	ts := time.Now().Unix()
	fname := fmt.Sprintf("./%s_%d.txt", tableName, ts)
	_ = ioutil.WriteFile(fname, []byte(protoStr), 0666)
}

func getDSN() string {
	dsnfmt := "%s:%s@tcp(%s:%d)/%s"
	dsn := fmt.Sprintf(dsnfmt, username, password, host, port, dbName)
	fmt.Printf("dsn:%s\n\n", dsn)
	return dsn
}

// TODO 完善更多
func getProtoType(s string) string {
	if strings.Contains(s, "bigint") {
		if strings.Contains(s, "unsigned") {
			return "uint64"
		} else {
			return "int64"
		}
	}
	if strings.Contains(s, "int") {
		if strings.Contains(s, "unsigned") {
			return "uint32"
		} else {
			return "int32"
		}
	}
	return "string"
}

func CamelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A'
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func FirstUpCase(s string) string {
	if len(s) == 0 {
		return s
	}

	if !isASCIILower(s[0]) {
		return s
	}
	c := s[0]
	c -= 'a' - 'A'
	b := []byte(s)
	b[0] = c
	return string(b)
}
