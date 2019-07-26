package api

import (
	"testing"
	"net/http"
	"io/ioutil"
	"sync"
)

var getTxUrl =  "http://192.168.8.54:8888/api/v1/txs?address="

//go test -v -run='TestGetTxs'
func TestGetTxs(t *testing.T) {
	addr := "1Gz1AFqFMnabMiAYicATrH8dV1D7dsz5Pu"
	_, err := Get(getTxUrl+ addr)
	if err != nil{
		t.Error(err)
		return
	}
	t.Logf("txs ok",)
}

//go test -v -run='TestGetTxsCurrency'
func TestGetTxsCurrency(t *testing.T) {
	addrs := []string{"1Gz1AFqFMnabMiAYicATrH8dV1D7dsz5Pu", "1WK3tBYqp31QoLujxwLjBMWNaxTgYEjLD", "1K5JKs8PvJv4vHjXtVHMQuLZJ2Vids2HpN"}
	addrNum := len(addrs)

	w := sync.WaitGroup{}

	testNum := 10000
	countCh := make(chan int, 100)
	waitCh :=  make(chan int, 100)

	go func() {
		for i := 0; i < 100; i++ {
			waitCh <- 1
		}
	}()

	go func() {
		for i := 0; i < testNum; i++ {
			<-waitCh

			go func(i int) {
				defer func() {
					countCh <-1
					waitCh<-1
				}()

				resp, err := Get(getTxUrl+ addrs[i % addrNum])
				if err != nil{
					t.Error(err)
					return
				}
				if i == -1 {
					t.Logf("resp= %v", resp)
				}
			}(i)
		}
	}()

	//等待
	w.Add(1)
	go func() {
		count := 0
		for  {
			select {
			case c := <-countCh:
				count+= c
			default:
			}
			if count == testNum {
				t.Logf("ok-----------%d", count)
				break
			}
		}
		w.Done()
	}()

	w.Wait()

	t.Logf("txs ok",)
}


func Get(url string) (string, error) {
	cli := http.DefaultClient
	resp, err := cli.Get(url)
	if err != nil{
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return "", err
	}
	return string(data), err
}