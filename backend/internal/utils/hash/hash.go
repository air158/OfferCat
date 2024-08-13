package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "123456"

	// 生成第一个哈希值
	hash1, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	fmt.Println("Hash 1:", string(hash1))

	// 生成第二个哈希值
	hash2, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	fmt.Println("Hash 2:", string(hash2))

	err = bcrypt.CompareHashAndPassword(hash1, []byte(password))
	if err != nil {
		fmt.Println("Password does not match")
	} else {
		fmt.Println("Password matches")
	}

	// 比较两个哈希值的结果
	fmt.Println("Hash 1 and Hash 2 match:", bcrypt.CompareHashAndPassword(hash1, []byte(password)) == nil)
}
