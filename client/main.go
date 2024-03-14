package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/midy177/remotedialer"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	addr := flag.String("connect", "ws://localhost:8123/connect", "要连接的地址")
	token := flag.String("token", "79aspb8r3t786gs4ryqj26b78fcmxg9r", "连接建立用token")
	debug := flag.Bool("debug", true, "Debug logging")
	proto := flag.String("proto", "tcp", "隧道转发协议")
	listen := flag.Int64("listen", 3306, "隧道转发到本地的端口")
	dst := flag.String("dst", "localhost:22", "隧道转发远程目的地址")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		"x-tunnel-token": []string{*token},
	}
	err := remotedialer.ClientConnect(context.Background(), *addr, headers,
		nil,
		func(proto, address string) bool {
			log.Printf("remotedialer: %s %s\n", proto, address)
			return true
		},
		func(ctx context.Context, session *remotedialer.Session) error {
			// 监听本地端口8080
			l, err := net.Listen(*proto, fmt.Sprintf(":%d", *listen))
			if err != nil {
				log.Panic(err)
			}
			defer l.Close()
			log.Printf("Listening on : %d", *listen)

			for {
				// 接受连接
				source, err := l.Accept()
				if err != nil {
					log.Println(err)
					continue
				}
				target, err := session.Dial(context.Background(), *proto, *dst)
				if err != nil {
					_, _ = source.Write([]byte(err.Error()))
					source.Close()
					log.Println(err)
					continue
				}
				// 对每个连接启动一个goroutine处理
				go handleClientRequest(source, target)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func handleClientRequest(source, target net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("recover in func handleClientRequest")
			log.Println(fmt.Sprintf("%T %v", err, err))
		}
	}()

	defer source.Close()
	defer target.Close()

	// 将客户端数据转发到目标
	go io.Copy(target, source)
	// 将目标数据转发回客户端
	io.Copy(source, target)
}
