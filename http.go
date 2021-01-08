package main

import (
	"bytes"
	"fmt"
	httpclient "github.com/ddliu/go-httpclient"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func init() {
	httpclient.Defaults(httpclient.Map{
		httpclient.OPT_UNSAFE_TLS: true,
		httpclient.OPT_MAXREDIRS:  0,
	})
}

func PostJson(url string, v interface{}) (string, error) {
	resp, err := httpclient.PostJson(url, v)
	if err != nil {
		return "", err
	}

	respBody, err := resp.ToString()
	return respBody, nil
}

func PostForm(url string, headers map[string]string, params map[string]string) (string, []*http.Cookie, error) {
	resp, err := httpclient.WithHeaders(headers).Post(url, params)
	if err != nil {
		return "", nil, err
	}

	respBody, err := resp.ToString()
	return respBody, nil, err
}

func PostFormParams(url string, params map[string]string) (string, error) {
	resp, err := httpclient.Post(url, params)
	if err != nil {
		return "", err
	}

	respBody, err := resp.ToString()
	return respBody, nil
}

func Get(url string, params map[string]interface{}) (string, error) {
	resp, err := httpclient.Get(url, params)
	if err != nil {
		return "", err
	}

	respBody, err := resp.ToString()
	return respBody, nil
}


func GetWithCookieAndHeader(url string, headers map[string]string, cookie []*http.Cookie) (string, []*http.Cookie, error) {
	//for i := 0; i < len(cookieStr); i ++ {
	//	cookie := cookieStr[i]
	//	index := strings.Index(cookie, "=")
	//	key := cookie[0:index]
	//	value := cookie[index+1:len(cookie)]
	//
	//	fmt.Println(key)
	//	fmt.Println(value)
	//
	//}
	//
	//return
	resp, err := httpclient.WithHeaders(headers).WithCookie(cookie[0]).Get(url)
	if err != nil {
		return "", nil, err
	}

	fmt.Println("cookie: ", resp.Cookies())

	respBody, err := resp.ToString()
	return respBody, resp.Cookies(), nil
}

func GetUrl(url string) (string, error) {
	resp, err := httpclient.Get(url)
	if err != nil {
		return "", err
	}

	respBody, err := resp.ToString()
	return respBody, nil
}

func GetUrlWithHeaders(url string, headers map[string]string) (string, []*http.Cookie, error) {
	resp, err := httpclient.WithHeaders(headers).Get(url)
	if err != nil {
		return "", nil, err
	}

	fmt.Println("cookie: ", resp.Cookies())

	respBody, err := resp.ToString()
	return respBody, resp.Cookies(), nil
}

//PostFile 上传文件
func PostFile(fieldname, filename, uri string) ([]byte, error) {
	fields := []MultipartFormField{
		{
			IsFile:    true,
			Fieldname: fieldname,
			Filename:  filename,
		},
	}
	return PostMultipartForm(fields, uri)
}

//MultipartFormField 保存文件或其他字段信息
type MultipartFormField struct {
	IsFile    bool
	Fieldname string
	Value     []byte
	Filename  string
}

//PostMultipartForm 上传文件或其他多个字段
func PostMultipartForm(fields []MultipartFormField, uri string) (respBody []byte, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	for _, field := range fields {
		if field.IsFile {
			fileWriter, e := bodyWriter.CreateFormFile(field.Fieldname, field.Filename)
			if e != nil {
				err = fmt.Errorf("error writing to buffer , err=%v", e)
				return
			}

			fh, e := os.Open(field.Filename)
			if e != nil {
				err = fmt.Errorf("error opening file , err=%v", e)
				return
			}
			defer fh.Close()

			if _, err = io.Copy(fileWriter, fh); err != nil {
				return
			}
		} else {
			partWriter, e := bodyWriter.CreateFormField(field.Fieldname)
			if e != nil {
				err = e
				return
			}
			valueReader := bytes.NewReader(field.Value)
			if _, err = io.Copy(partWriter, valueReader); err != nil {
				return
			}
		}
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, e := http.Post(uri, contentType, bodyBuf)
	if e != nil {
		err = e
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	return
}
