package radius

import (
	"bytes"
	"crypto/md5"
)

// PAP 密码验证
func pap(shareSecret, password string, rp RadiusPackage) bool {
	hash := md5.New()
	hash.Write([]byte(shareSecret))
	hash.Write(rp.Authenticator[:])
	b := hash.Sum(nil)

	value := []byte(password)
	length := len(value)

	shortOf := length % 16
	times := length/16 + ifVal(shortOf > 0, 1, 0).(int)
	supplement := make([]byte, shortOf)
	value = append(value, supplement...)
	result := make([]byte, 0)
	for i := 0; i < times; i++ {
		left := value[i*16 : (i+1)*16]
		ret := [16]byte{}
		for j := 0; j < 16; j++ {
			ret[j] = left[j] ^ b[j]
		}
		result = append(result, ret[:]...)
	}

	attr := rp.RadiusAttrMap[AttrKey{Standard, 2}]
	return bytes.Equal(result, attr.AttrValue)
}

// CHAP认证 MD5(ID + PASSWORD明文 + CHALLENGE)
func chap(password string, rp *RadiusPackage) bool {
	var chapPassword []byte
	for _, attr := range rp.RadiusAttrs {
		if attr.AttrType == 3 {
			chapPassword = attr.AttrValue
		}
	}

	if len(chapPassword) != 17 {
		return false
	}

	var chapId = chapPassword[0]
	challenge := rp.challenge
	if len(challenge) != 16 {
		challenge = rp.Authenticator[:]
	}

	hashPassword := chapPassword[1:]

	buffer := bytes.NewBuffer(nil)
	buffer.WriteByte(chapId)
	buffer.Write([]byte(password))
	buffer.Write(challenge)
	sum := md5.Sum(buffer.Bytes())
	return bytes.Equal(sum[:], hashPassword)
}

func ifVal(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

