package captcha

import (
	"github.com/dchest/captcha"
	"github.com/google/uuid"
	"net/http"
	"fmt"
	"bytes"
)

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
