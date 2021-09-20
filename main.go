package main

import (
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
	port      string
	user      string
	passwd    string
	dbName    string
	tableName string
)

func main() {
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
	dsn := "%s:%s@tcp(%s:%s)/%s"
	return fmt.Sprintf(dsn, user, passwd, host, port, dbName)
}
