package storage

import (
	"bytes"
	"dxkite.cn/storage/meta"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	UPLOAD_SUCCESS = iota
	UPLOAD_ERROR_AUTH
	UPLOAD_ERROR_HASH
	UPLOAD_ERROR_PUSH
	UPLOAD_ERROR_META
	UPLOAD_ERROR_STOR
)

const (
	PERMIT_UPLOAD = "upload"
)

// 上传
type UploadHandler struct {
	// 验证用远程服务接口
	// 支持http/https验证
	AuthRemote string
	// 验证用请求头
	// Token数据从请求头中获取
	AuthField string
	// 上传服务器
	Usn string
	// 上传块大小
	BlockSize int
	// 上传临时目录
	Temp string
	// 上传保存目录
	Root string
}

var (
	errPermit = errors.New("permit error")
)

// 请求验证
type AuthRequest struct {
	Token    string `json:"token"`
	RemoteIp string `json:"remote_ip"`
}

type UserInfo struct {
	Name   string   `json:"name"`
	Permit []string `json:"permit"`
}

// 验证响应
type AuthResponse struct {
	*UserInfo
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
	var token string
	if req.Method == http.MethodGet {
		token = req.URL.Query().Get(u.AuthField)
	} else if req.Method == http.MethodPut {
		token = req.Header.Get(u.AuthField)
	} else if req.Method == http.MethodOptions {
		resp.Header().Set("access-control-allow-origin", req.Header.Get("origin"))
		resp.Header().Set("access-control-allow-method", "OPTIONS,GET,PUT")
		resp.Header().Set("access-control-allow-header", "Content-Type,"+u.AuthField)
		return
	} else {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resp.Header().Set("access-control-allow-origin", "*")
	_ = os.MkdirAll(u.Root, os.ModePerm)
	_ = os.MkdirAll(u.Temp, os.ModePerm)

	var user *UserInfo
	if uu, err := u.auth(req, token); err != nil {
		u.respError(resp, UPLOAD_ERROR_AUTH, err)
		return
	} else {
		user = uu
	}
	if req.Method == http.MethodGet {
		u.Get(user, resp, req)
	} else {
		u.Upload(user, resp, req)
	}
}

func (u *UploadHandler) Get(user *UserInfo, resp http.ResponseWriter, req *http.Request) {
	hh := strings.ToLower(strings.TrimLeft(req.URL.Path, "/"))
	mp := path.Join(u.Root, hh+EXT_META)
	log.Println("get", hh, mp)
	if FileExist(mp) {
		if b, err := ioutil.ReadFile(mp); err == nil {
			resp.Header().Set("content-type", "application/octet-stream")
			resp.Header().Set("content-disposition", `attachment; filename="`+hh+EXT_META+`"`)
			resp.Header().Set("content-length", strconv.Itoa(len(b)))
			resp.WriteHeader(http.StatusOK)
			_, _ = resp.Write(b)
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		resp.WriteHeader(http.StatusNotFound)
	}
}

func (u *UploadHandler) Upload(user *UserInfo, resp http.ResponseWriter, req *http.Request) {
	defer func() { _ = req.Body.Close() }()
	name := filepath.Base(req.URL.Path)
	hash := strings.ToLower(req.URL.Query().Get("hash"))
	mod := time.Now().Unix()
	if m := req.URL.Query().Get("time"); len(m) > 0 {
		// Y-m-d h:i:s
		if t, er := time.Parse("2006-01-02 15:04:05", m); er == nil {
			mod = t.Unix()
		}
	}
	// 快传
	mp := path.Join(u.Root, hash+EXT_META)
	h, her := hex.DecodeString(hash)
	if FileExist(mp) {
		log.Println("fast upload", user.Name, hash)
		if b, er := ioutil.ReadFile(mp); er == nil {
			u.respData(resp, map[string]interface{}{
				"hash": h,
				"name": name,
				"meta": b,
			})
			return
		}
	}

	if user.hasPermit(PERMIT_UPLOAD) == false {
		u.respError(resp, UPLOAD_ERROR_AUTH, errPermit)
		return
	}

	var tnh string
	if her == nil {
		tnh = hash
	} else {
		tn := ByteHash([]byte(fmt.Sprintf("%s:%s:%d", user.Name, req.URL.Path, time.Now().Unix())))
		tnh = hex.EncodeToString(tn)
	}

	size := req.ContentLength

	up := NewUploader(int64(u.BlockSize), u.Usn)

	up.Processor = NewFileUploadProcessor(path.Join(u.Temp, tnh+EXT_UPLOADING), NewUploadInfo(name, size, up.Size, mod))
	up.Notify = &ConsoleNotify{}

	if mt, err := up.UploadStream(req.Body); err != nil {
		u.respError(resp, UPLOAD_ERROR_PUSH, err)
		return
	} else {
		h := hex.EncodeToString(mt.Hash)
		d := &bytes.Buffer{}
		if err := meta.EncodeToStream(d, mt); err != nil {
			u.respError(resp, UPLOAD_ERROR_META, err)
			return
		}
		b := d.Bytes()
		if err := ioutil.WriteFile(path.Join(u.Root, h+EXT_META), b, os.ModePerm); err != nil {
			u.respError(resp, UPLOAD_ERROR_STOR, err)
			return
		}
		u.respData(resp, map[string]interface{}{
			"hash": mt.Hash,
			"name": mt.Name,
			"meta": b,
		})
	}
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
		Code:    UPLOAD_SUCCESS,
		Message: "success",
		Data:    data,
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

func (u *UserInfo) hasPermit(n string) bool {
	for _, a := range u.Permit {
		if a == n {
			return true
		}
	}
	return false
}

func (u *UploadHandler) auth(req *http.Request, token string) (user *UserInfo, err error) {
	if len(u.AuthRemote) == 0 {
		log.Println("no auth")
		return &UserInfo{"", []string{PERMIT_UPLOAD}}, nil
	}
	r := AuthRequest{
		Token:    token,
		RemoteIp: req.RemoteAddr,
	}
	rj, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	resp, er := http.Post(u.AuthRemote, "application/json", bytes.NewBuffer(rj))
	if er != nil {
		return nil, err
	}
	rs := new(AuthResponse)
	defer func() { _ = resp.Body.Close() }()
	bt, rer := ioutil.ReadAll(resp.Body)
	if rer != nil {
		return nil, rer
	}
	if er := json.Unmarshal(bt, rs); er != nil {
		return nil, er
	}
	if rs.Code != UPLOAD_SUCCESS {
		return nil, errors.New(fmt.Sprintf("auth error: errno(%d): %s", rs.Code, rs.Message))
	}
	return rs.UserInfo, nil
}
