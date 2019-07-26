package main

import (
	"github.com/liyue201/btc-scan/store/leveldb"
	"github.com/pborman/uuid"
	"fmt"
	"time"
)

func main()  {
	s, _ := leveldb.NewLevelDbStorage("./data")

	t1 := time.Now().Unix()

	for i := 0; i  < 10000000; i++ {
		key := uuid.New()
		key = key + key
		value := key + key
		value = value + value

		s.Set(key, value)

		if i % 10000 == 0 {
			fmt.Printf("%v %d\n", time.Now().String(), i)
		}
	}

	s.Close()

	t2 := time.Now().Unix()

	fmt.Printf("d = %d\n", t2 - t1)

}