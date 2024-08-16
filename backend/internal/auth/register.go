package auth

//func Register(c *gin.Context) {
//	var user User
//	if err := c.ShouldBindJSON(&user); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		lib.Err()
//		return
//	}
//
//	// 密码哈希处理
//	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
//		return
//	}
//	user.PasswordHash = string(hashedPassword)
//
//	// 保存用户信息到数据库
//	if err := db.DB.Create(&user).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
//
//}
