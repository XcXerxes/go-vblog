/*
 * @Description: 主程序
 * @Author: leo
 * @Date: 2020-02-16 11:20:31
 * @LastEditors: leo
 * @LastEditTime: 2020-02-16 11:21:03
 */
package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MASIZE int64 = 10 * 1024 * 1024

// 输出hello world
func sayHello(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world"))
}

// 上传视频接口
func uploadHandler(w http.ResponseWriter, req *http.Request) {
	// 1、限制客户端上传视频文件的大小
	req.Body = http.MaxBytesReader(w, req.Body, MASIZE)
	fmt.Println(req)
	if err := req.ParseMultipartForm(MASIZE); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 2、获取上传的文件
	file, fileHeader, err := req.FormFile("uploadFile")
	// 3、检查文件类型
	if ret := strings.HasSuffix(fileHeader.Filename, ".mp4"); ret == false {
		http.Error(w, "not mp4", http.StatusInternalServerError)
	}
	// 4、获取随机名称
	md5Byte := md5.Sum([]byte(fileHeader.Filename + time.Now().String()))
	md5Str := fmt.Sprintf("%x", md5Byte)
	newFileName := md5Str + ".mp4"

	// 5、写入文件
	fmt.Println(newFileName)
	dst, err := os.Create("video/" + newFileName)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

// 获取视频列表
func getFileListHandler(w http.ResponseWriter, req *http.Request) {
	// 1、解析 video/ 下面所有的 .mp4后缀的文件
	files, _ := filepath.Glob("video/*.mp4")
	var ret []string
	// 2、循环遍历解析出来的 切片
	for _, file := range files {
		// 将 协议 host 还有文件名一起保存到一个切片中
		var s strings.Builder
		s.WriteString("http://")
		s.WriteString(req.Host)
		s.WriteString(filepath.Base(file))
		ret = append(ret, s.String())
	}
	retJson, _ := json.Marshal(ret)
	// 返回到客户端
	w.Write(retJson)
	return
}
func main() {
	// 实现读取文件 handler
	fileHandler := http.FileServer(http.Dir("./video"))
	http.Handle("/video/", http.StripPrefix("/video/", fileHandler))
	http.HandleFunc("/hello", sayHello)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/list", getFileListHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
