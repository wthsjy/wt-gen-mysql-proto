# install
```
go install -v github.com/wthsjy/wt-gen-mysql-proto@latest
```

```shell
$GOBIN/wt-gen-mysql-proto -h

  -P int
        端口号 (default 3306)
  -d string
        数据库名
  -h string
        主机名,默认 127.0.0.1 (default "127.0.0.1")
  -p string
        密码
  -t string
        表名
  -u string
        用户名
```


```protobuf
// mysql database.table: demo.user_info
message UserInfo{
  // 用户id (field comment by table ddl)
  uint64 userId = 1; // @gotags: gorm:"column:user_id"
  // 用户昵称
  string userName = 2; // @gotags: gorm:"column:user_name"
  // 创建时间
  string createTime = 3; // @gotags: gorm:"column:create_time"
  // 更新时间1
  // 更新时间2
  string updateTime = 4; // @gotags: gorm:"column:update_time"
}


```