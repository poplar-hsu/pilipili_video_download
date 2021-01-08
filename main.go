package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var url = ""
var  urlReg = `base_url":.*?,"backupUrl`
var titleReg = `title=".*?"`
var jsonReg = `<script>window.__playinfo__=.*?</script>`

type MyJsonName struct {
	Code int64 `json:"code"`
	Data struct {
		AcceptDescription []string `json:"accept_description"`
		AcceptFormat      string   `json:"accept_format"`
		AcceptQuality     []int64  `json:"accept_quality"`
		Dash              struct {
			Audio []struct {
				//SegmentBase struct {
				//	Initialization string `json:"Initialization"`
				//	IndexRange     string `json:"indexRange"`
				//} `json:"SegmentBase"`
				//BackupURL   []string `json:"backupUrl"`
				BackupURL   []string `json:"backup_url"`
				Bandwidth   int64    `json:"bandwidth"`
				//BaseURL     string   `json:"baseUrl"`
				BaseURL     string   `json:"base_url"`
				Codecid     int64    `json:"codecid"`
				Codecs      string   `json:"codecs"`
				//FrameRate   string   `json:"frameRate"`
				FrameRate   string   `json:"frame_rate"`
				Height      int64    `json:"height"`
				ID          int64    `json:"id"`
				//MimeType    string   `json:"mimeType"`
				MimeType    string   `json:"mime_type"`
				Sar         string   `json:"sar"`
				SegmentBase struct {
					IndexRange     string `json:"index_range"`
					Initialization string `json:"initialization"`
				} `json:"segment_base"`
				//StartWithSap int64 `json:"startWithSap"`
				StartWithSap int64 `json:"start_with_sap"`
				Width        int64 `json:"width"`
			} `json:"audio"`
			Duration      int64   `json:"duration"`
			//MinBufferTime float64 `json:"minBufferTime"`
			MinBufferTime float64 `json:"min_buffer_time"`
			Video         []struct {
				//SegmentBase struct {
				//	Initialization string `json:"Initialization"`
				//	IndexRange     string `json:"indexRange"`
				//} `json:"SegmentBase"`
				//BackupURL   []string `json:"backupUrl"`
				BackupURL   []string `json:"backup_url"`
				Bandwidth   int64    `json:"bandwidth"`
				//BaseURL     string   `json:"baseUrl"`
				BaseURL     string   `json:"base_url"`
				Codecid     int64    `json:"codecid"`
				Codecs      string   `json:"codecs"`
				//FrameRate   string   `json:"frameRate"`
				FrameRate   string   `json:"frame_rate"`
				Height      int64    `json:"height"`
				ID          int64    `json:"id"`
				//MimeType    string   `json:"mimeType"`
				MimeType    string   `json:"mime_type"`
				Sar         string   `json:"sar"`
				SegmentBase struct {
					IndexRange     string `json:"index_range"`
					Initialization string `json:"initialization"`
				} `json:"segment_base"`
				//StartWithSap int64 `json:"startWithSap"`
				//StartWithSap int64 `json:"start_with_sap"`
				Width        int64 `json:"width"`
			} `json:"video"`
		} `json:"dash"`
		Format         string `json:"format"`
		From           string `json:"from"`
		Message        string `json:"message"`
		Quality        int64  `json:"quality"`
		Result         string `json:"result"`
		SeekParam      string `json:"seek_param"`
		SeekType       string `json:"seek_type"`
		SupportFormats []struct {
			DisplayDesc    string `json:"display_desc"`
			Format         string `json:"format"`
			NewDescription string `json:"new_description"`
			Quality        int64  `json:"quality"`
			Superscript    string `json:"superscript"`
		} `json:"support_formats"`
		Timelength   int64 `json:"timelength"`
		VideoCodecid int64 `json:"video_codecid"`
	} `json:"data"`
	Message string `json:"message"`
	Session string `json:"session"`
	TTL     int64  `json:"ttl"`
}
func main()  {
	for {
		if url != "" {
			break
		}

		fmt.Println("请输入视频地址（按回车键确认）：")
		fmt.Scanln(&url)
	}

	video,audio, err := download(url)
	fmt.Println(video,audio)
	if err != nil {
		return
	}

	fmt.Println("开始进行视频音频合并")
	err = mergeVA(video,audio)
	if err != nil {
		return
	}
}


func download(url string) (string, string, error){
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36",
	}
	resp, cookies, err:= GetUrlWithHeaders(url, headers)
	if err != nil {
		return "", "", err
	}

	compiles := regexp.MustCompile(jsonReg)
	submatchs := compiles.FindString(resp)
	resJsonstr := submatchs[28:len(submatchs)-9]

	resJsons := MyJsonName{}
	err = json.Unmarshal([]byte(resJsonstr),&resJsons)
	if err != nil {
		fmt.Println("解析失败")
		return "", "", err
	}

	hostReg := `"baseUrl":"https://(.*?)/upgcxcode`
	compile2 := regexp.MustCompile(hostReg)
	submatch2 := compile2.FindAllSubmatch([]byte(resp), -1)
	host := string(submatch2[1][0])[11:]
	var videoTitle string
	var audioTitle string
	fmt.Println("开始下载源视频文件——————————")
	for key,video := range resJsons.Data.Dash.Video {
		videoTitle, err = downloadVideo(key, 1, video.BaseURL, host, cookies)
		if err != nil {
			fmt.Println("视频下载失败")
			continue
		}
		break
	}
	fmt.Println("开始下载源音频文件——————————")
	for key,audio := range resJsons.Data.Dash.Audio {
		audioTitle, err = downloadVideo(key, 2, audio.BaseURL, host, cookies)
		if err != nil {
			fmt.Println("音频下载失败")
			continue
		}
		break
	}

	return videoTitle, audioTitle, nil

}

func downloadVideo(key, types int, videoUrl, host string, cookies []*http.Cookie) (string, error){
	fmt.Println("第",key+1,"次尝试")
	var title string
	if(types == 1) {
		title = strconv.Itoa(key+1)+".mp4"
	}else {
		title = strconv.Itoa(key+1)+".mp3"
	}

	headers1 := map[string]string {
		"Host": host,
		"Connection": "keep-alive",
		"Origin": "https://www.bilibili.com",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36",
		"Accept": "*/*",
		"Referer": url + "/",

		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-CN,zh;q=0.9",
	}

	//fmt.Println(headers1)
	//fmt.Println(len(cookies))
	//res,_, err := GetWithCookieAndHeader(videoUrl,headers1,cookies)
	////fmt.Println(res)
	//content := []byte(res)
	//
	//err = ioutil.WriteFile(title, content, 07777)
	//if err != nil {
	//	fmt.Println("文档下载失败")
	//	fmt.Println(err)
	//	return "", err
	//}
	down(videoUrl, title, headers1, cookies, func(length, downLen int64) {
		totalLen := float64(length)/1024/1024
		totalLen = math.Trunc(totalLen*100)/100
		downLeng := float64(downLen)/1024/1024
		downLeng = math.Trunc(downLeng*100)/100
		pro := float64(downLen)/float64(length)

		pro = math.Trunc(pro*10000)/100
		//fmt.Println("文件总大小： ", totalLen, "MB, 当前已下载： ", downLeng, "MB, 下载进度： ", pro, "%")
		fmt.Printf("文件总大小： %.2f MB, 当前已下载： %.2f MB, 下载进度： %.2f %%\r", totalLen,downLeng,pro)
	})
	fmt.Println("文档下载成功")
	return title, nil
}


func mergeVA(video, audio string) error {
	now  := time.Now()
	title := now.Format("20060102150405") + ".mp4"
	cmd := exec.Command("ffmpeg", "-i", video, "-i", audio, "-c", "copy", title)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	//fmt.Println("Result: " + out.String())


	//合并成功，删除源视频音频
	if err = os.Remove(video); err != nil {
		fmt.Println("源视频文件删除失败")
	}

	if err = os.Remove(audio); err != nil {
		fmt.Println("源音频文件删除失败")
	}

	fmt.Println("视频下载成功！")

	return nil
}

func down(url,title string,headers map[string]string, cookies []*http.Cookie, fb func(length, downLen int64)) error{
	var (
		fsize int64
		buf = make([]byte, 32*1024)
		written int64
	)

	req, _ := http.NewRequest("GET", url, nil)
	for key, header := range headers {
		req.Header.Set(key, header)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//读取服务器返回的文件大小
	fsize, err = strconv.ParseInt(resp.Header.Get("content-Length"), 10, 32)
	if err != nil {
		return err
	}

	file, err := os.Create(title)
	if err != nil {
		return err
	}
	defer file.Close()

	if resp.Body == nil {
		return errors.New("body is null")
	}

	for {
		// 读取bytes
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			// 写入bytes
			nw,ew := file.Write(buf[0:nr])
			// 数据长度大于0
			if nw > 0 {
				written += int64(nw)
			}
			//写入错误
			if ew != nil {
				err = ew
				break
			}
			// 读取是数据长度不等于写入的数据长度
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}

		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		fb(fsize, written)
	}

	return err
}