package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type DDLM struct {
	Field   string
	Type    string
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
	flag.StringVar(&username, "u", "", "用户名,默认为空")
	flag.StringVar(&password, "p", "", "密码,默认为空")
	flag.StringVar(&host, "h", "127.0.0.1", "主机名,默认 127.0.0.1")
	flag.IntVar(&port, "P", 3306, "端口号,默认为空")

	// 从arguments中解析注册的flag。必须在所有flag都注册好而未访问其值时执行。未注册却使用flag -help时，会返回ErrHelp。
	flag.Parse()

	dsn := getDSN()
	mDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = mDB.Raw("SHOW FULL FIELDS FROM  user_info").Find(&ddlms).Error
	if err != nil {
		panic(err)
	}

	protoStr := fmt.Sprintf("message %s{\n", tableName)
	for index, v := range ddlms {
		if strings.TrimSpace(v.Comment) != "" {
			protoStr += fmt.Sprintf("	// %s\n", strings.ReplaceAll(strings.TrimSpace(v.Comment), "\n", "\\\\ \n"))
		}
		protoStr += fmt.Sprintf("	%s %s = %d;\n", v.Type, v.Field, index)
	}
	protoStr += "\n}"

	fmt.Printf("\n\n")
	fmt.Println(protoStr)
	fmt.Printf("\n\n")
}

func getDSN() string {
	dsn := "%s:%s@tcp(%s:%d)/%s"
	return fmt.Sprintf(dsn, username, password, host, port, dbName)
}
