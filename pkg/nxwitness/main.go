package nxwitness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type NxCloud struct {
	SystemId string
	Ip       string
	Port     int
	Server   string
	User     string
	Pass     string
	Token    string
}

func NewNxCloud(systemId, user, pass, ip string, port int) (*NxCloud, error) {
	var nx NxCloud
	nx.SystemId, nx.User, nx.Pass, nx.Ip, nx.Port = systemId, user, pass, ip, port

	if nx.SystemId != "" {
		nx.Server = fmt.Sprintf("https://%v.relay.vmsproxy.com", nx.SystemId)
	} else if nx.Ip != "" && nx.Port != 0 {
		nx.Server = fmt.Sprintf("http://%v:%d", nx.Ip, nx.Port)
	} else {
		return nil, fmt.Errorf("Datos de acceso al cloud inválidos.")
	}

	err := nx.setLoginToken()
	if err != nil {
		return nil, fmt.Errorf("Autenticación con el cloud fallida: %v", err)
	}

	return &nx, nil
}

type LoginPost struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Token      string `json:"token"`
	AgeS       int    `json:"ageS"`
	ExpiresInS int    `json:"expiresInS"`
}

func (nx *NxCloud) setLoginToken() error {
	body := []byte(fmt.Sprintf(`{
        "username": "%v",
        "password": "%v",
        "setCookie": true
    }`, nx.User, nx.Pass))

	postUrl := fmt.Sprintf("%v/rest/v3/login/sessions", nx.Server)

	res, err := http.Post(
		postUrl,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	post := &LoginPost{}
	derr := json.NewDecoder(res.Body).Decode(post)
	if derr != nil {
		return derr
	}

	if res.StatusCode != 200 {
		fmt.Println(res.Status)
	}
	nx.Token = post.Token

	return nil
}
