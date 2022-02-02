package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gofrs/uuid"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsNumber(data string) bool {
	_, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return false
	}
	return true
}

func IsPortBind(protocol string, hostname string, port int, timeout int) bool {
	p := strconv.Itoa(port)
	addr := net.JoinHostPort(hostname, p)
	_, err := net.DialTimeout(protocol, addr, time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	return true
}

func GetTimestamp_S_String() string {
	timeUnix := time.Now().Unix()
	return strconv.FormatInt(timeUnix, 10)
}

func GenDate_String() string {
	formatTimeStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	return formatTimeStr
}

func GenRandomString(length uint) string {
	str := `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	bytes := []byte(str)
	var result []byte
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < int(length); i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

func JoinString(str ...string) string {
	var ReturnStr string
	for _, v := range str {
		ReturnStr += v
	}
	return ReturnStr
}

func Base64ToString(str_base64 string) string {
	decodeBytes, err := base64.StdEncoding.DecodeString(str_base64)
	if err != nil {
		return ""
	}
	return string(decodeBytes)
}

func StringToBase64(str string) string {
	encodeString := base64.StdEncoding.EncodeToString([]byte(str))
	return encodeString
}

func BigIntReduce(numstr string, num string) string {
	n, _ := new(big.Int).SetString(numstr, 10)
	numInt64, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return "nil"
	}
	m := new(big.Int)
	m.SetInt64(-numInt64)
	m.Add(n, m)
	return m.String()
}

func Sha256Sum(str string) string {
	StrSecret := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%x", StrSecret)
}

func StringNumCompare(string1 string, string2 string) int {
	a, _ := new(big.Int).SetString(string1, 10)
	b, _ := new(big.Int).SetString(string2, 10)
	return a.Cmp(b)
}

func GenUUID() string {
	UUIDGen, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", UUIDGen)
}

func TimestampToDate(timestamp int64) string {
	return fmt.Sprintf("%v", time.Unix(timestamp, 0))
}
