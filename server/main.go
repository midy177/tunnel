package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/midy177/remotedialer"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func main() {
	listen := flag.String("listen", ":1323", "Listen address")
	token := flag.String("token", "79aspb8r3t786gs4ryqj26b78fcmxg9r", "Tunnel service client establishes connection token")
	peerToken := flag.String("peerToken", "nbyugqgya76bke9x3rbkhcdsm7yh3z84", "Tunnel server cluster node token")
	peerID := flag.String("peerID", "", "Tunnel server node ID(peerID)")
	peers := flag.String("peers", "", "Peers format id:token:url,id:token:url")
	debuge := flag.Bool("debug", true, "Debug logging")

	flag.Parse()

	handler := remotedialer.New(func(req *http.Request) (clientKey string, authed bool, err error) {
		id := req.Header.Get("x-tunnel-token")
		return strconv.FormatInt(time.Now().Unix(), 10), id == *token, nil
	}, remotedialer.DefaultErrorWriter)
	handler.ClientConnectAuthorizer = func(proto, address string) bool {
		log.Printf("remotedialer: %s %s\n", proto, address)
		return true
	}
	handler.PeerToken = *peerToken
	if *peerID == "" {
		log.Fatal("Tunnel server node ID(peerID) can't be null")
	}
	handler.PeerID = *peerID
	if *debuge {
		logrus.SetLevel(logrus.DebugLevel)
	}
	for _, peer := range strings.Split(*peers, ",") {
		parts := strings.SplitN(strings.TrimSpace(peer), ":", 3)
		if len(parts) != 3 {
			continue
		}
		handler.AddPeer(parts[2], parts[0], parts[1])
	}

	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 日志记录或错误处理
					fmt.Printf("Recovered from a panic: %v\n%s", err, debug.Stack())
					// 返回服务器错误响应
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	})
	router.Handle("/connect", handler)
	fmt.Println("Listening on ", *listen)
	err := http.ListenAndServe(*listen, router)
	if err != nil {
		panic(err)
	}
}
