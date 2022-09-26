package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"strings"
)

type DDLM struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Comment string
	Extra   string
	Default string
}
type TableStatus struct {
	Name    string
	Comment string
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
	flag.StringVar(&username, "u", "root", "用户名")
	flag.StringVar(&password, "p", "123456", "密码")
	flag.StringVar(&host, "h", "localhost", "主机名,默认 127.0.0.1")
	flag.IntVar(&port, "P", 3306, "端口号")
	flag.StringVar(&dbName, "d", "test", "数据库名")
	flag.StringVar(&tableName, "t", "user_info", "表名")

	flag.Parse()

	dsn := getDSN()
	mDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	tableStatus := &TableStatus{}
	err = mDB.Raw(fmt.Sprintf("show table status like '%s'", tableName)).First(&tableStatus).Error
	if err != nil {
		panic(err)
	}

	err = mDB.Raw(fmt.Sprintf("SHOW FULL FIELDS FROM  %s", tableName)).Find(&ddlms).Error
	if err != nil {
		panic(err)
	}

	var fields []string
	//var priKeyField string
	structStr := fmt.Sprintf("package dbmodel\n\n// %s %s mysql database.table: %s.%s\ntype %s struct {\n", FirstUpCase(CamelCase(tableName)), tableStatus.Comment, dbName, tableName, FirstUpCase(CamelCase(tableName)))
	for _, v := range ddlms {
		fields = append(fields, fmt.Sprintf("`%s`", v.Field))
		if strings.TrimSpace(v.Comment) != "" {
			structStr += fmt.Sprintf("	// %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\n  // "))
		}
		structStr += fmt.Sprintf("	%s %s ", FirstUpCase(CamelCase(v.Field)), getStructType(v))
		if strings.ToUpper(v.Key) == "PRI" {
			structStr += fmt.Sprintf("`gorm:\"column:%s;->\"`\n", v.Field)
		} else if strings.ToUpper(v.Type) == "DATETIME" && strings.ToUpper(v.Default) == "CURRENT_TIMESTAMP" {
			structStr += fmt.Sprintf("`gorm:\"column:%s;->\"`\n", v.Field)
		} else {
			structStr += fmt.Sprintf("`gorm:\"column:%s\"`\n", v.Field)
		}

	}

	structStr += "}\n\n"

	structFileName := fmt.Sprintf("./%s.model.go", tableName)
	_ = ioutil.WriteFile(structFileName, []byte(structStr), 0666)
	fmt.Println("struct:", structFileName)
}

func getDSN() string {
	dsnfmt := "%s:%s@tcp(%s:%d)/%s"
	dsn := fmt.Sprintf(dsnfmt, username, password, host, port, dbName)
	fmt.Printf("dsn:%s\n\n", dsn)
	return dsn
}

// TODO 完善更多
func getProtoType(s string) string {
	if strings.Contains(s, "int") {
		return "int64"
	}
	return "string"
}

func getStructType(s DDLM) string {
	types := strings.ToLower(s.Type)
	if strings.Contains(types, "int") {
		return "int64"
	}
	if strings.Contains(strings.ToLower(types), "datetime") {
		return "time.Time"
	}
	if strings.Contains(strings.ToLower(types), "timestamp") {
		return "time.Time"
	}
	if strings.Contains(strings.ToLower(types), "float") {
		return "float32"
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

func FirstUpCase(str string) string {
	if len(str) == 0 {
		return str
	}

	if !isASCIILower(str[0]) {
		return str
	}
	c := str[0]
	c -= 'a' - 'A'
	b := []byte(str)
	b[0] = c
	return string(b)
}
