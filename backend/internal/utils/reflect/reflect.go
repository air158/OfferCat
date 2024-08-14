package main

import (
	"fmt"
	"reflect"
)

type Person struct {
	Name string
	Age  int
	City string
}

func printStructFields(s interface{}) {
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()

	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		fmt.Printf("Field Name: %s, Field Value: %v\n", fieldType.Name, field.Interface())
	}
}

func main() {
	p := Person{Name: "Alice", Age: 30, City: "New York"}
	printStructFields(p)
}
