package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v3"
)

// 配置结构体
type Config struct {
	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"smtp"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

// 邮件请求结构体
type MailRequest struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
}

var conf Config

func loadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
}

func sendMailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req MailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", conf.SMTP.Username)
	m.SetHeader("To", req.To...)
	m.SetHeader("Subject", req.Subject)
	if req.IsHTML {
		m.SetBody("text/html", req.Body)
	} else {
		m.SetBody("text/plain", req.Body)
	}

	d := gomail.NewDialer(
		conf.SMTP.Host,
		conf.SMTP.Port,
		conf.SMTP.Username,
		conf.SMTP.Password,
	)

	if err := d.DialAndSend(m); err != nil {
		http.Error(w, fmt.Sprintf("发送失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
	loadConfig()

	http.HandleFunc("/send", sendMailHandler)
	log.Printf("邮件服务启动，监听端口 %d", conf.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Server.Port), nil))
}
