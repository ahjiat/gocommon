package captcha

import (
	"github.com/patrickmn/go-cache"
	"github.com/dchest/captcha"
	"github.com/google/uuid"
	"encoding/base64"
	"net/http"
	"fmt"
	"time"
	"bytes"
)

var localCache *cache.Cache

const (
	NoExpiration time.Duration = cache.NoExpiration
)

func init() {
	localCache = cache.New(cache.NoExpiration, 5 * time.Minute)
}

func Set(key string, value interface{}, t time.Duration) {
	localCache.Set(key, value, t)
}

func Get[T any](key string) (T, bool) {
	if value, found := localCache.Get(key); found {
		return value.(T), true
	}
	return *new(T), false
}

func Delete(key string) {
	localCache.Delete(key)
}




func WriteImage(w http.ResponseWriter) (string, string) {
	var imgValue string
	outUUID := uuid.New().String()
	digits := captcha.RandomDigits(6)
	img := captcha.NewImage(outUUID, digits, 180, 80)
	w.Header().Set("Content-Type", "image/png")
	img.WriteTo(w)
	for _, d := range digits {
		imgValue += fmt.Sprintf("%d", d)
	}
	return outUUID, imgValue
}

func GenImage() (string, string, []byte) {
	var imgValue string
	outUUID := uuid.New().String()
	imgBuf := &bytes.Buffer{}
	digits := captcha.RandomDigits(6)
	img := captcha.NewImage(outUUID, digits, 180, 80)
	img.WriteTo(imgBuf)
	for _, d := range digits {
		imgValue += fmt.Sprintf("%d", d)
	}
	return outUUID, imgValue, imgBuf.Bytes()
}

func GenBase64() (string, string, string) {
	uuid, digits, imgInByte := GenImage()
	return uuid, digits, "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgInByte)
}

func GetBase64SecureImg(expire time.Duration) (string, string) {
	uuid, digits, base64Img := GenBase64()
	Set(uuid, digits, expire)
	return uuid, base64Img
}

func Verify(uuid string, digits string) bool {
	value, found := Get[string](uuid); if ! found { return false }
	if value != digits { return false }
	return true
}
