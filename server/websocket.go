package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
)

// Stream -
type Stream struct {
	TokenID string `json:"token_id" form:"token_id"`
	Sbl     string `json:"s_bl" form:"s_bl" binding:"required"`
	Ssc     string `json:"s_sc" form:"s_sc" binding:"required"`
	Ssw     string `json:"s_sw" form:"s_sw" binding:"required"`
	Svt     string `json:"s_vt" form:"s_vt" binding:"required"`
	A       string `json:"a" form:"a" binding:"required"`
}

type rtcConn struct {
	conn *webrtc.PeerConnection
}

type streamBody struct {
	Type    string                    `json:"type"`
	Message string                    `json:"message"`
	Offer   webrtc.SessionDescription `json:"offer"`
	Answer  webrtc.SessionDescription `json:"answer"`
	State   webrtc.StatsReport        `json:"state"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func facebook(c *gin.Context) {
	var req Stream
	if err := c.BindQuery(&req); err != nil {
		log.Print(err.Error())
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}
	req.TokenID = c.Param("token_id")
	if req.TokenID == "" {
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}

	sock, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "upgrade request failed",
		})
		return
	}
	defer sock.Close()

	var body streamBody
	var conn *webrtc.PeerConnection

	// socket & webrtc
	for {
		err = sock.ReadJSON(&body)
		if err != nil {
			log.Print(err)
			err = nil
			break
		}
		switch body.Type {
		case "offer":
			conn, err = StartWebRTC(&body)
			body.Type = "answer"
			log.Printf("send answer to %s", req.TokenID)
			sock.WriteJSON(body)
		case "state":
			body.Type = "state_answer"
			body.State = conn.GetStats()
			log.Printf("send state to %s", req.TokenID)
			sock.WriteJSON(body)
		}
	}

	if err != nil {
		c.JSON(500, gin.H{
			"message": "upgrade request failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "end of session",
	})
	return
}

func facebookBinary(c *gin.Context) {
	var req Stream
	if err := c.BindQuery(&req); err != nil {
		log.Print(err.Error())
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}
	req.TokenID = c.Param("token_id")
	if req.TokenID == "" {
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}

	sock, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "upgrade request failed",
		})
		return
	}
	defer sock.Close()

	// ffmpeg
	cmd := exec.Command("ffmpeg", "-i", "-",
		"-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency",
		"-c:a", "aac", "-ar", "44100", "-b:a", "64k",
		"-y",
		"-use_wallclock_as_timestamps", "1",
		"-async", "1",
		"-bufsize", "1000",
		"-f", "flv", "rtmps://live-api-s.facebook.com:443/rtmp/158220232414733?s_bl=1&s_sc=158220255748064&s_sw=0&s_vt=api-s&a=AbxcLvHeTwLmLzeq",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("%s", err.Error())
	}
	defer stdin.Close()
	cmd.Start()
	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Printf("wait command error: %s", err.Error())
		}
	}()
	// socket
	for {
		_, msg, err := sock.ReadMessage()
		if err != nil {
			log.Print(err)
			err = nil
			break
		}
		_, err = io.Copy(stdin, bytes.NewReader([]byte(msg)))
		if err != nil {
			log.Printf("error write to stdin: %s", err.Error())
		}
	}

	if err != nil {
		c.JSON(500, gin.H{
			"message": "upgrade request failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "end of session",
	})
	return
}

// StartWebSocket -
func StartWebSocket() {
	app := gin.Default()
	app.GET("/stream/webrtc/:token_id", facebook)
	app.GET("/stream/binary/:token_id", facebookBinary)
	app.Run(":8080")
}
