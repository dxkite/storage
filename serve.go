package storage

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	UPLOAD_SUCCESS = iota
	UPLOAD_ERROR_AUTH
	UPLOAD_ERROR_SIZE
	UPLOAD_ERROR_HASH
	UPLOAD_ERROR_PUSH
)

// 上传
type UploadHandler struct {
	// 验证用远程服务接口
	// 支持http/https验证
	AuthRemote string
	// 验证用请求头
	// Token数据从请求头中获取
	AuthHeader string
	// 上传服务器
	Usn string
	// 上传块大小
	BlockSize int
}

// 请求验证
type AuthRequest struct {
	Token    string `json:"token"`
	RemoteIp string `json:"remote_ip"`
}

// 验证响应
type AuthResponse struct {
	// 0 验证成功
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 响应
type UploadResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (u *UploadHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if err := u.auth(req); err != nil {
		u.respError(resp, UPLOAD_ERROR_AUTH, err)
		return
	}
	name := "./" + req.RequestURI
	hash64 := req.Header.Get("hash")
	var hash []byte
	if b, err := hex.DecodeString(hash64); err != nil {
		u.respError(resp, UPLOAD_ERROR_HASH, err)
		return
	} else {
		hash = b
	}
	defer func() { _ = req.Body.Close() }()
	up := NewUploader(int64(u.BlockSize), u.Usn)
	if err := up.UploadStream(name, hash, req.ContentLength, req.Body); err != nil {
		u.respError(resp, UPLOAD_ERROR_PUSH, err)
		return
	}
	u.respData(resp, "ok")
}

func (u *UploadHandler) respError(resp http.ResponseWriter, code int, err error) {
	r := UploadResponse{
		Code:    code,
		Message: err.Error(),
	}
	if rj, err := json.Marshal(r); err != nil {
		resp.Header().Set("Error", err.Error())
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Header().Set("content-type", "application/json")
		resp.WriteHeader(http.StatusOK)
		_, _ = resp.Write(rj)
	}
}

func (u *UploadHandler) respData(resp http.ResponseWriter, data interface{}) {
	r := UploadResponse{
		Code: UPLOAD_SUCCESS,
		Data: data,
	}
	if rj, err := json.Marshal(r); err != nil {
		resp.Header().Set("Error", err.Error())
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Header().Set("content-type", "application/json")
		resp.WriteHeader(http.StatusOK)
		_, _ = resp.Write(rj)
	}
}

func (u *UploadHandler) auth(req *http.Request) error {
	if len(u.AuthRemote) == 0 {
		log.Println("on auth")
		return nil
	}
	r := AuthRequest{
		Token:    req.Header.Get(u.AuthHeader),
		RemoteIp: req.RemoteAddr,
	}
	rj, err := json.Marshal(r)
	if err != nil {
		return err
	}
	resp, er := http.Post(u.AuthRemote, "application/json", bytes.NewBuffer(rj))
	if er != nil {
		return er
	}
	rs := new(AuthResponse)
	defer func() { _ = resp.Body.Close() }()
	bt, rer := ioutil.ReadAll(resp.Body)
	if rer != nil {
		return nil
	}
	if er := json.Unmarshal(bt, rs); er != nil {
		return er
	}
	if rs.Code != UPLOAD_SUCCESS {
		return errors.New(fmt.Sprintf("auth error: errno(%d): %s", rs.Code, rs.Message))
	}
	return nil
}
