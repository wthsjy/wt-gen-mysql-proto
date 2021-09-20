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
	structStr := fmt.Sprintf("package dbmodel\n\n// %s mysql database.table: %s.%s\ntype %s struct{\n", FirstUpCase(CamelCase(tableName)), dbName, tableName, FirstUpCase(CamelCase(tableName)))
	protoStr := "syntax = \"proto3\";\n\n"
	protoStr += "package wtmicro.pbgen.service.model;\n"
	protoStr += "option go_package = \"wtmicro/pbgen/service/model;modelpb\";\n"
	protoStr += fmt.Sprintf("// mysql database.table: %s.%s\nmessage %s{\n", dbName, tableName, FirstUpCase(CamelCase(tableName)))
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

		if strings.TrimSpace(v.Comment) != "" {
			structStr += fmt.Sprintf("  // %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\n  // "))
		}
		structStr += fmt.Sprintf("  %s %s  ", FirstUpCase(CamelCase(v.Field)), getProtoType(v.Type))
		structStr += fmt.Sprintf("`json:\"%s\" gorm:\"column:%s\"`\n", CamelCase(v.Field), v.Field)
	}

	protoStr += "}"

	structStr += "}\n\n"
	structStr += "const(\n"
	structStr += fmt.Sprintf("  sqlAdd%s =\"insert into `%s`(%s)values(%s)\"\n", FirstUpCase(CamelCase(tableName)), tableName, strings.Join(fields, ","), strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ","))
	structStr += fmt.Sprintf("  sqlDel%sByIds =\"delete from `%s` where `%s` in ?\"\n", FirstUpCase(CamelCase(tableName)), tableName, priKeyField)
	structStr += fmt.Sprintf("  sqlGet%sByIds =\"select %s from `%s` where `%s` in ?\"\n", FirstUpCase(CamelCase(tableName)), strings.Join(fields, ","), tableName, priKeyField)
	structStr += ")"

	protoFileName := fmt.Sprintf("./%s.model.proto", tableName)
	_ = ioutil.WriteFile(protoFileName, []byte(protoStr), 0666)
	fmt.Println("proto:", protoFileName)

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

// UserInfo mysql database.table: demo.user_info
type UserInfo struct {
	// 用户id
	UserId uint64 `json:"userId" gorm:"column:user_id"`
	// 用户昵称
	UserName string `json:"userName" gorm:"column:user_name"`
	// 创建时间
	CreateTime string `json:"createTime" gorm:"column:create_time"`
	// 更新时间1
	// 更新时间2
	UpdateTime string `json:"updateTime" gorm:"column:update_time"`
}
