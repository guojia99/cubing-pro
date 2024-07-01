package utils

//// Encrypt 加密函数
//func Encrypt(plaintext string, key []byte) (string, error) {
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return "", err
//	}
//
//	// 创建一个GCM
//	gcm, err := cipher.NewGCM(block)
//	if err != nil {
//		return "", err
//	}
//
//	// 创建一个随机nonce
//	nonce := make([]byte, gcm.NonceSize())
//	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
//		return "", err
//	}
//	// 将时间戳编码成字符串并加密
//	timestamp := time.Now().Unix()
//	plaintext += fmt.Sprintf(":%d", timestamp)
//	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
//	return base64.StdEncoding.EncodeToString(ciphertext), nil
//}
//
//func EncryptToByte(plaintext string, key []byte) ([]byte, error) {
//	out, err := Encrypt(plaintext, key)
//	return []byte(out), err
//}
//
//// Decrypt 解密函数
//func Decrypt(ciphertext string, key []byte) (string, error) {
//	// 先将密文解码
//	data, err := base64.StdEncoding.DecodeString(ciphertext)
//	if err != nil {
//		return "", err
//	}
//
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return "", err
//	}
//
//	gcm, err := cipher.NewGCM(block)
//	if err != nil {
//		return "", err
//	}
//
//	// 先解密
//	nonceSize := gcm.NonceSize()
//	if len(data) < nonceSize {
//		return "", fmt.Errorf("ciphertext too short")
//	}
//
//	nonce, ciphertext := data[:nonceSize], string(data[nonceSize:])
//	plaintext, err := gcm.Open(nil, nonce, []byte(ciphertext), nil)
//	if err != nil {
//		return "", err
//	}
//
//	// 分割出时间戳并返回
//	parts := bytes.Split(plaintext, []byte(":"))
//	if len(parts) < 2 {
//		return "", fmt.Errorf("invalid plaintext format")
//	}
//
//	return string(parts[0]), nil
//}

//参考文档
//http://www.topgoer.com/%E5%85%B6%E4%BB%96/%E5%8A%A0%E5%AF%86%E8%A7%A3%E5%AF%86/%E5%8A%A0%E5%AF%86%E8%A7%A3%E5%AF%86.html
//高级加密标准（Adevanced Encryption Standard ,AES）

//// PKCS7 填充模式
//func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
//	padding := blockSize - len(ciphertext)%blockSize
//	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
//	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
//	return append(ciphertext, padtext...)
//}
//
//// 填充的反向操作，删除填充字符串
//func PKCS7UnPadding1(origData []byte) ([]byte, error) {
//	//获取数据长度
//	length := len(origData)
//	if length == 0 {
//		return nil, errors.New("加密字符串错误！")
//	} else {
//		//获取填充字符串长度
//		unpadding := int(origData[length-1])
//		//截取切片，删除填充字节，并且返回明文
//		return origData[:(length - unpadding)], nil
//	}
//}
//
//// 实现加密
//func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
//	//创建加密算法实例
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return nil, err
//	}
//	//获取块的大小
//	blockSize := block.BlockSize()
//	//对数据进行填充，让数据长度满足需求
//	origData = PKCS7Padding(origData, blockSize)
//	//采用AES加密方法中CBC加密模式
//	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
//	crypted := make([]byte, len(origData))
//	//执行加密
//	blocMode.CryptBlocks(crypted, origData)
//	return crypted, nil
//}
//
//// 实现解密
//func AesDeCrypt(cypted []byte, key []byte) (string, error) {
//	//创建加密算法实例
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return "", err
//	}
//	//获取块大小
//	blockSize := block.BlockSize()
//	//创建加密客户端实例
//	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
//	origData := make([]byte, len(cypted))
//	//这个函数也可以用来解密
//	blockMode.CryptBlocks(origData, cypted)
//	//去除填充字符串
//	origData, err = PKCS7UnPadding1(origData)
//	if err != nil {
//		return "", err
//	}
//	return string(origData), err
//}

//func EnPwdCode(pwdStr string, ts int64) (string, error) {
//	pwd := []byte(pwdStr)
//	result, err := AesEcrypt(pwd, GenerateRandomKey(ts))
//	if err != nil {
//		return "", err
//	}
//	return hex.EncodeToString(result), nil
//}
//
//func DePwdCode(pwd string, ts int64) (string, error) {
//	temp, err := hex.DecodeString(pwd)
//	if err != nil {
//		return "", err
//	}
//	//执行AES解密
//	res, err := AesDeCrypt(temp, GenerateRandomKey(ts))
//	if err != nil {
//		return "", err
//	}
//	return res, err
//}

func EnPwdCode(pwdStr string, ts int64) (string, error) {
	//e := cryptor.AesSimpleEncrypt("Hello World!", key)
	return "", nil
}
func DePwdCode(pwd string, ts int64) (string, error) {
	return "", nil
}
