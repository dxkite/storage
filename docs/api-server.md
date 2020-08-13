## Api Server

a storage server provider a put api to upload file

### run command

```bash
storage -usn 'plugin://cmd?exec=python&args=./vim-cn.py' -addr ':8080' -auth http://127.0.0.1/auth.php
```

### auth server

```php
<?php
$request = json_decode(file_get_contents('php://input'), true);
if ($request['token'] === 'dxkite') {
    echo json_encode([
        'code' => 0,
        'message' => 'success',
        'name' => 'dxkite',
        'permit' => ['upload'] // can upload
    ]);
} else {
    echo json_encode([
        'code' => 1,
        'message' => 'auth error'
    ]);
}
```

### upload test

```bash
curl --location --request PUT 'http://127.0.0.1:8080/storage/陈奕迅 - 十年.mp3' \
--header 'token: dxkite' \
--header 'Content-Type: audio/mpeg' \
--data-binary '@/E:/陈奕迅 - 十年.mp3'
```

#### response

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "hash": "kYg/Rptb0mZy/m921cB7QrL+Xss=",
        "meta": "FFNNRgGc+Kmm/vDz//fw+K2m+KqppvTo6OzvprOz/fmsrbL98PX/+PKy//Pxs/f6s9SurK79+fr5+P3//f+oqq2u/qv5qqqs+az9+Kyurq3++Mqy9uz7rab0rqymchIVU8Uwk1C+/wLxuJXHKW35wvitpvX1rPn5+K2m+KqppvTo6OzvprOz/fmsrbL98PX/+PKy//Pxs/f6s9SvqK6q+Kmsr6yr/q+orKmo/v6t+KmtpP6t/qStqa2q+cuy9uz7rab0rqymqjKnAJ+hkaWMed2qsCy8gU0GK3+tpvX1rfn5+a2spv7w8//3w+/15vn1rqylq62prvmqpvny//P4+fWt+aim9P3v9K6sppGIP0abW9Jmcv5vdtXAe0Ky/l7LqKby/fH5rq6mdQUUeTkJdCMZvLG8eREdeSUosvHsr6im7/Xm+fWvrqiupK6u+aqm7+j96Onv9a75qKbo5ez59az5+Q==",
        "name": "陈奕迅 - 十年.mp3"
    }
}
```
