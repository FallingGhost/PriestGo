package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var LogFilename = fmt.Sprintf("%s.log", time.Now().Format("01.02"))

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func StrfTime(format string) string {
	format = strings.ReplaceAll(format, "%Y", "2006")
	format = strings.ReplaceAll(format, "%m", "01")
	format = strings.ReplaceAll(format, "%B", "January")
	format = strings.ReplaceAll(format, "%b", "Jan")
	format = strings.ReplaceAll(format, "%d", "02")
	format = strings.ReplaceAll(format, "%A", "Monday")
	format = strings.ReplaceAll(format, "%a", "Mon")
	format = strings.ReplaceAll(format, "%H", "15")
	format = strings.ReplaceAll(format, "%I", "03")
	format = strings.ReplaceAll(format, "%p", "PM")
	format = strings.ReplaceAll(format, "%M", "04")
	format = strings.ReplaceAll(format, "%S", "05")
	return time.Now().Format(format)
}

func ReleasePort() error {
	result, err := Popen("netstat -ano | findstr 60724")
	if err != nil {
		return err
	}
	fmt.Println(result)
	re, err := regexp.Compile(`(\d)+`)
	if err != nil {
		return err
	}
	match := re.FindAllString(string(result), -1)

	if match[len(match)-1] != "" {
		_, err := Popen("taskkill /f /pid " + match[len(match)-1])
		if err != nil {
			return err
		}
	}
	return nil
}

func Popen(command string) ([]byte, error) {
	err := exec.Command("chcp", "65001").Run()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("cmd", "/c", fmt.Sprintf(`%s`, command))
	cmd.Stderr = os.Stderr
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Cd(args []string) (string, error) {
	if len(args) == 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return " ", err
		}
		return workDir, nil
	} else {
		err := os.Chdir(strings.Join(args, " "))
		if err != nil {
			return " ", err
		}
	}
	return " ", nil
}

func LogWrite(msg string, currentPath string) error {
	log, err := os.OpenFile(fmt.Sprintf("%s/.log/%s", currentPath, LogFilename),
		os.O_CREATE|os.O_APPEND,
		0)
	if err != nil {
		return err
	}
	_, err = log.Write([]byte(
		fmt.Sprintf("[%s] %s\r\n",
			StrfTime("%H:%M:%S"), msg)))
	if err != nil {
		return err
	}
	return nil
}

func Encode(orig string, key string) string {
	origData := []byte(orig)
	k := []byte(key)
	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()
	origData = padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	cryted := make([]byte, len(origData))
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)
}

func Decode(cryted string, key string) string {
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	k := []byte(key)
	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = unPadding(orig)
	return string(orig)
}

func padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func unPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
