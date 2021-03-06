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
	flag.StringVar(&password, "p", "root", "密码")
	flag.StringVar(&host, "h", "127.0.0.1", "主机名,默认 127.0.0.1")
	flag.IntVar(&port, "P", 3306, "端口号")
	flag.StringVar(&dbName, "d", "demo", "数据库名")
	flag.StringVar(&tableName, "t", "admin_account", "表名")

	// 从arguments中解析注册的flag。必须在所有flag都注册好而未访问其值时执行。未注册却使用flag -help时，会返回ErrHelp。
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
	var priKeyField string
	structStr := fmt.Sprintf("package dbmodel\n\n// %s %s mysql database.table: %s.%s\ntype %s struct {\n", FirstUpCase(CamelCase(tableName)), tableStatus.Comment, dbName, tableName, FirstUpCase(CamelCase(tableName)))
	protoStr := "syntax = \"proto3\";\n\n"
	protoStr += "package dodo.go.pbgen.service.model;\n"
	protoStr += "option go_package = \"dodo-go/pbgen/service/model;modelpb\";\n\n"
	protoStr += fmt.Sprintf("//  %s mysql database.table: %s.%s\nmessage %s{\n", tableStatus.Comment, dbName, tableName, FirstUpCase(CamelCase(tableName)))
	for index, v := range ddlms {
		fields = append(fields, fmt.Sprintf("`%s`", v.Field))
		if strings.ToUpper(v.Key) == "PRI" {
			priKeyField = v.Field
		}
		if strings.TrimSpace(v.Comment) != "" {
			protoStr += fmt.Sprintf("  // %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\n  // "))
		}
		protoStr += fmt.Sprintf("  //\n  // table field:\n  // @gotags: gorm:\"column:%s\"\n", v.Field)
		protoStr += fmt.Sprintf("  %s %s = %d;\n", getProtoType(v.Type), CamelCase(v.Field), index+1)

		if strings.TrimSpace(v.Comment) != "" {
			structStr += fmt.Sprintf("	// %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\n  // "))
		}
		structStr += fmt.Sprintf("	%s %s ", FirstUpCase(CamelCase(v.Field)), getStructType(v))
		structStr += fmt.Sprintf("`json:\"%s\" gorm:\"column:%s\"`\n", CamelCase(v.Field), v.Field)
	}

	protoStr += "}"

	structStr += "}\n\n"
	structStr += "const (\n"
	structStr += fmt.Sprintf("	sqlAdd%s = \"insert into `%s`(%s)values(%s)\"\n", FirstUpCase(CamelCase(tableName)), tableName, strings.Join(fields, ","), strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ","))
	structStr += fmt.Sprintf("	sqlDel%sByIds = \"delete from `%s` where `%s` in ?\"\n", FirstUpCase(CamelCase(tableName)), tableName, priKeyField)
	structStr += fmt.Sprintf("	sqlGet%sByIds = \"select %s from `%s` where `%s` in ?\"\n", FirstUpCase(CamelCase(tableName)), strings.Join(fields, ","), tableName, priKeyField)
	structStr += ")"

	protoFileName := fmt.Sprintf("./%s.model.proto", tableName)
	structFileName := fmt.Sprintf("./%s.model.go", tableName)
	_ = ioutil.WriteFile(protoFileName, []byte(protoStr), 0666)
	_ = ioutil.WriteFile(structFileName, []byte(structStr), 0666)
	fmt.Println("proto:", protoFileName)
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

// TODO 完善更多
func getStructType(s DDLM) string {
	types := strings.ToLower(s.Type)
	nullable := strings.ToUpper(s.Null) == "YES"
	if strings.Contains(types, "bigint") {
		if nullable {
			return "sqlnull2json.NullInt64"
		}
		if strings.Contains(types, "unsigned") {
			return "uint64"
		} else {

			return "int64"
		}
	}
	if strings.Contains(types, "int") {
		if nullable {
			return "sqlnull2json.NullInt32"
		}
		if strings.Contains(types, "unsigned") {
			return "uint32"
		} else {
			return "int32"
		}
	}
	if strings.Contains(strings.ToLower(types), "datetime") {
		if nullable {
			return "sqlnull2json.NullTime"
		}
		return "time.Time"
	}
	if strings.Contains(strings.ToLower(types), "float") {
		if nullable {
			return "sqlnull2json.NullFloat64"
		}
		return "float64"
	}
	if nullable {
		return "sqlnull2json.NullString"
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
