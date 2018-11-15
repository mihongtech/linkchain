package main

type inner struct{
	name string
	age int
}

type outer struct{
	inner
	gender int
	location string
}
//
//func main() {
//	o := outer{inner{"pop", 10}, 1, "aganda"}
//	fmt.Println(o.name)
//}
