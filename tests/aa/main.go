package main


import (
	"fmt"
)

type A struct {
	name string
}

type Test struct {
	a []A
}

func test()  {
	t := Test{}
	t.a = make([]A, 0, 0 )

	for i := 0; i < 10; i++ {
		k := A{
			name: "aaa",
		}
		t.a = append(t.a, k)
	}

	//fmt.Printf("%#v\n", t.a)

	for _, k := range t.a {
		k.name = "bbb"
	}

	for _, k := range t.a {
		fmt.Printf("%#v\n", k.name)
	}



}

func main()  {
	test()
}