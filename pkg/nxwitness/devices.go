package nxwitness

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func GetDeviceThumbnailB64(cl *NxCloud, format, deviceId string, timestamp int64) (string, error) {
	postUrl := fmt.Sprintf(
		"%v/rest/v2/devices/%v/image?timestampUs=%d&rotation=10&format=%v&size=720x480&tolerant=true",
		cl.Server,
		deviceId,
		timestamp,
		format,
	)
	contentType := fmt.Sprintf("image/%v", format)
	auth := fmt.Sprintf("Bearer %v", cl.Token)

	req, err := http.NewRequest(
		http.MethodGet,
		postUrl,
		nil,
	)

	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", contentType)

	// fmt.Println(req)
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error al leer la respuesta: %v", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error al leer la respuesta: %v", err)
	}

	if res.StatusCode != 200 {
		fmt.Println("status:", res.Status)
		return "", fmt.Errorf("Error al obtener la imagen, status: %v", res.StatusCode)
	}

	img := base64.StdEncoding.EncodeToString(body)
	// img = "data:image/jpeg;base64," + img
	return img, nil
}
