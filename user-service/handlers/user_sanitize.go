package handlers

// sanitizeUserResult 移除用户信息中的敏感字段，避免泄露到前端
func sanitizeUserResult(m map[string]interface{}) {
	if m == nil {
		return
	}

	// 常见密码字段命名
	delete(m, "password")
	delete(m, "Password")
	delete(m, "hashed_password")
	delete(m, "hashedPassword")
}


