package main

import "encoding/json"

type TestDaughter struct {
	ChildName string
}

type TestSon struct {
	ChildName string
}

type TestParent struct {
	ParentName string
	Son        *TestSon
	Daughter   *TestDaughter
}

func main() {
	//fmt.Println("TEST")
	//var list []int
	//for i := range list {
	//	fmt.Println(i)
	//}

	testSon := TestSon{ChildName: "A"}
	var testParent TestParent
	testParent.ParentName = "Parent"
	testParent.Son = &testSon
	bytes, err := json.Marshal(testParent)
	if err != nil {
		panic(err)
	}
	println(string(bytes))
}
