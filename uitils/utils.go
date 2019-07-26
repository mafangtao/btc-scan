package utils

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	utf8 "unicode/utf8"
)

func FilterEmoji(content string) string {
	new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new_content += string(value)
		}
	}
	return new_content
}

func JsonMap(from, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}

func GenPwd(pwd, salt string) string {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(pwd)), 10)
	if err != nil {
		//return fmt.Sprintf("%x", hashPwd)
		return ""
	}

	return fmt.Sprintf(string(hashPwd))
}

func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

var exeDir string

func GetExeDir() string {
	if exeDir != "" {
		return exeDir
	}
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	exeDir = filepath.Dir(path)

	return exeDir
}

func IsRelactivePath(path string) bool {
	if strings.Index(path, ".") == 0 {
		return true
	}
	return false
}

func AbsPath(path string) string {
	if IsRelactivePath(path) {
		path = GetExeDir() + string(os.PathSeparator) + path
		return path
	}
	return path
}

func Xor(num1, num2 string) (string, error) {
	var bigNum1 big.Int
	var bigNum2 big.Int
	var result big.Int

	if _, ok := bigNum1.SetString(num1, 10); !ok {
		return "", fmt.Errorf("invalid num1")
	}
	if _, ok := bigNum2.SetString(num2, 10); !ok {
		return "", fmt.Errorf("invalid num2")
	}

	result.Xor(&bigNum1, &bigNum2)

	return result.String(), nil
}

func Ip2Int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func Int2Ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func Struct2Map(r interface{}) (s map[string]string, err error) {
	var temp map[string]interface{}
	var result = make(map[string]string)

	bin, err := json.Marshal(r)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(bin, &temp); err != nil {
		return nil, err
	}
	for k, v := range temp {
		v2, ok := v.(string)
		if ok {
			result[k] = v2
		}
	}
	return result, nil
}

func Round(v float64) int64 {
	return int64(v + 0.5)
}

func Md5(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func GenUuidString() string {
	uuid, err := uuid.NewV1()
	if err != nil {
		return ""
	}
	return strings.Replace(uuid.String(), "-", "", -1)
}
