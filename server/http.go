package server

// import (
// 	"bytes"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool { return true },
// }

// func facebook(c *gin.Context) {
// 	tokenID := c.Param("token_id")
// 	if tokenID == "" {
// 		log.Print("no token id")
// 		c.Status(400)
// 		return
// 	}
// 	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		log.Print("upgrade:", err)
// 		c.Status(500)
// 		return
// 	}
// 	defer conn.Close()

// 	cmd := sendRTMP("ffmpeg -vcodec libvpx -i - -c:v libx264 -preset veryfast -tune zerolatency -an -bufsize 1000 -f flv 'rtmp://54.254.183.160:19350/rtmp/143570563879700?s_bl=1&s_sc=143570600546363&s_sw=0&s_vt=api-s&a=Abz537tK6d1PF9HH'")
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		log.Print("stream stdin:", err)
// 		c.Status(500)
// 		return
// 	}

// 	go cmd.Start()

// 	for {
// 		mt, message, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println("read:", err)
// 			err = nil
// 			break
// 		}

// 		log.Print(string(message))
// 		stdin.Write(message)

// 		err = conn.WriteMessage(mt, message)
// 		if err != nil {
// 			log.Println("write:", err)
// 			err = nil
// 			break
// 		}
// 	}

// 	b, _ := cmd.CombinedOutput()
// 	log.Print("log from ffmpeg: ", string(b))

// 	if err != nil {
// 		log.Println("wait:", err)
// 		c.Status(500)
// 		return
// 	}

// 	if err := cmd.Wait(); err != nil {
// 		log.Println("wait:", err)
// 		c.Status(500)
// 		return
// 	}

// 	c.Status(200)
// 	return
// }

// func sendRTMP(cmdStr string) (cmd *exec.Cmd) {
// 	// "ffmpeg -vcodec libvpx -i - -c:v libx264 -preset veryfast -tune zerolatency -an -bufsize 1000 -f flv 'rtmp://54.254.183.160:19350/rtmp/143570563879700?s_bl=1&s_sc=143570600546363&s_sw=0&s_vt=api-s&a=Abz537tK6d1PF9HH'"
// 	// ffmpeg -r 30 -f lavfi -i - -c:v -vcodec libx264 -profile:v baseline -pix_fmt yuv420p -f flv 'rtmp://54.254.183.160:19350/rtmp/143570563879700?s_bl=1&s_sc=143570600546363&s_sw=0&s_vt=api-s&a=Abz537tK6d1PF9HH'
// 	args := strings.Split(cmdStr, " ")
// 	cmd = exec.Command(args[0], args[1:]...)
// 	return
// }

// func videoConvert(in string, out string) {
// 	os.Remove(out)
// 	//fmt.Println(in, out)
// 	cmdStr := fmt.Sprintf("ffmpeg -i %s -loglevel quiet -c copy -bsf:v h264_mp4toannexb -f mpegts %s", in, out)
// 	args := strings.Split(cmdStr, " ")
// 	msg, err := Cmd(args[0], args[1:])
// 	if err != nil {
// 		fmt.Printf("videoConvert failed, %v, output: %v\n", err, msg)
// 		return
// 	}
// }

// func Cmd(commandName string, params []string) (cmd *exec.Cmd, err error) {
// 	cmd = exec.Command(commandName, params...)
// 	//fmt.Println("Cmd", cmd.Args)
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = os.Stderr
// 	err = cmd.Start()
// 	if err != nil {
// 		return
// 	}
// 	return
// }

// func StartHTTP() {
// 	app := gin.Default()
// 	app.GET("/stream/facebook/:token_id", facebook)

// 	app.Run(":8080")
// }
