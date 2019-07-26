package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"time"
)

var session *gocql.Session

//初始化
func init() {
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Keyspace = "mycas"
	cluster.Consistency = gocql.Consistency(1)
	cluster.NumConns = 3
	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		fmt.Printf("start %v\n", err)
		return
	}
}

//创建表
func TestScylla_Create() {
	//query := fmt.Sprintf(`DROP table user`)
	//session.Query(query).Exec()

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS user(id int PRIMARY KEY, user_name varchar);`)
	session.Query(query).Exec()
}


//插入数据
func TestScylla_Insert() {
	t0 := time.Now().UnixNano() / 1000000
	t1 := t0

	for i := 1; i < 10000;i++ {
		query := fmt.Sprintf(`INSERT INTO user (id,user_name) VALUES (%d,'wang %010d')`, i, i)
		err := session.Query(query).Exec()
		if err != nil {
			fmt.Println(err)
			break
		}
		if i % 1000 == 0 {
			t1 = time.Now().UnixNano() / 1000000
			fmt.Printf("[%s] i=%010d t=%d\n", time.Now().String(), i, int(1000.0 * 1000/(t1-t0)))

			t0 = t1
		}
	}
}

//删除表
func TestScylla_Drop() {
	query := fmt.Sprintf(`drop table user;`)
	err := session.Query(query).Exec()

	if err != nil {
		fmt.Println(err)
	}
}

//查询数据
func TestScylla_Select() {
	query := fmt.Sprintf("SELECT * from user;")
	iter := session.Query(query).Iter()
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()
	var id int
	var name string
	for iter.Scan(&id, &name) {
		fmt.Println(id, name)
	}
}

//批量执行数据
func TestScylla_Batch() {
	query := fmt.Sprintf(`BEGIN BATCH
            UPDATE user SET user_name = 'asdqw' where id = %d;
            INSERT INTO user (id,user_name) VALUES (2,'zhangsan');
            APPLY BATCH;`, 1)
	err := session.Query(query).Exec()
	if err != nil {
		fmt.Println(err)
	}
}

func main()  {
	TestScylla_Create()
	TestScylla_Insert()
	//TestScylla_Select()
}
