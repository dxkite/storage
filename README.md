# storage

[![Build Status](https://travis-ci.org/dxkite/storage.svg?branch=master)](https://travis-ci.org/dxkite/storage) 
[![Go Report Card](https://goreportcard.com/badge/github.com/dxkite/storage)](https://goreportcard.com/report/github.com/dxkite/storage)

a tool use free CDN store file

## how it works

many website provider api to store image on CDN without authority check, this tool store file pieces as image on CDN. 

## usage

```bash
# upload
storage -usn 'plugin://cmd?exec=python&args=./vim-cn.py' /path/to/file # upload file to CDN, generated pieces meta for download
# start a upload server
storage -usn 'plugin://cmd?exec=python&args=./vim-cn.py' -addr ':8080'
# download
storage /path/to/file.meta
# download url
storage http://storage.dxkite.cn/meta/mp4.meta
# download meta uri
storage storage://meta?dl=aHR0cDovL3N0b3JhZ2UuZHhraXRlLmNuL21ldGEvbXA0Lm1ldGE=
```

## web downloader

- [Web Downloader](http://storage.dxkite.cn/)
- Test 
    - [mp4 download test](http://storage.dxkite.cn/meta/mp4.meta)
    
## Extra

- [API](./docs/api-server.md)
- [Plugin](./docs/plugin.md)
