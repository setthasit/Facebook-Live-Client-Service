package server

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/pion/rtp/codecs"

	"github.com/pion/webrtc/v2"
)

func StartWebRTC(body *streamBody) (conn *webrtc.PeerConnection, err error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	conn, err = webrtc.NewPeerConnection(config)
	if err != nil {
		log.Printf("cannot create peer connection: %s", err.Error())
		return
	}

	_, err = conn.AddTransceiver(webrtc.RTPCodecTypeAudio)
	if err != nil {
		log.Printf("cannot create peer connection: %s", err.Error())
		return
	}
	_, err = conn.AddTransceiver(webrtc.RTPCodecTypeVideo)
	if err != nil {
		log.Printf("cannot create peer connection: %s", err.Error())
		return
	}

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			log.Print("add ice candidate")
			err := conn.AddICECandidate(candidate.ToJSON())
			if err != nil {
				log.Print("add ice candidate failed")
			}
			log.Print("add ice candidate completed")
		}
	})

	conn.OnTrack(func(t *webrtc.Track, rcv *webrtc.RTPReceiver) {
		codec := t.Codec()
		if err != nil {
			log.Print("add ice candidate failed")
		}

		if codec.Name == webrtc.VP8 {
			// ffmpeg
			// ffmpeg -max_delay 5000 -reorder_queue_size 16384 -protocol_whitelist file,crypto,udp,rtp -re -i input.sdp -vcodec copy -acodec aac -f flv rtmp://127.0.0.1:1935/live/myStream
			cmd := exec.Command("ffmpeg", "-i", "-",
				"-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency",
				"-c:a", "aac", "-ar", "44100", "-b:a", "64k",
				"-y",
				"-use_wallclock_as_timestamps", "1",
				"-async", "1",
				"-bufsize", "1000",
				"-f", "flv", "rtmp://54.254.183.160:19350/rtmp/157657392471017?s_bl=1&s_sc=157657419137681&s_sw=0&s_vt=api-s&a=AbxYo4VY6UA6aFsE",
			)
			var stdOut, stdErr bytes.Buffer
			cmd.Stdout = &stdOut
			cmd.Stderr = &stdErr
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
					log.Printf("wait command error: %s", stdErr.String())
					return
				}
			}()

			go func() {
				for {
					if stdin == nil {
						log.Printf("wait command error: %s", err.Error())
						return
					}

					packet, err := t.ReadRTP()
					if err != nil {
						log.Printf("read rtp failed: %s", err.Error())
						return
					}

					data, err := (&codecs.VP8Packet{}).Unmarshal(packet.Payload)
					if err != nil {
						log.Printf("Unmarshal vp8 failed: %s", err.Error())
						return
					}

					log.Print(string(data))

					io.Copy(stdin, bytes.NewReader(data))
				}
			}()
		}
	})

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("connection state: %s", state.String())
	})

	conn.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("ice state: %s", state.String())
		if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateFailed {
			conn.Close()
			log.Print("webrtc close")
			return
		}
	})

	err = conn.SetRemoteDescription(body.Offer)
	if err != nil {
		log.Printf("cannot set remote desc: %s",
			err.Error())
		return
	}

	ans, err := conn.CreateAnswer(nil)
	if err != nil {
		log.Printf("cannot set remote desc: %s", err.Error())
		return
	}

	err = conn.SetLocalDescription(ans)
	if err != nil {
		log.Printf("cannot set remote desc: %s", err.Error())
		return
	}

	body.Answer = ans
	return
}

func populateStdin(file []byte) func(io.WriteCloser) {
	return func(stdin io.WriteCloser) {
		defer stdin.Close()
		io.Copy(stdin, bytes.NewReader(file))
	}
}

func runFFMPEGFromStdin(populate_stdin_func func(io.WriteCloser)) {
	cmd := exec.Command("ffmpeg", "-i", "-",
		"-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency",
		"-c:a", "aac", "-ar", "44100", "-b:a", "64k",
		"-y",
		"-use_wallclock_as_timestamps", "1",
		"-async", "1",
		"-bufsize", "1000",
		"-f", "flv", "rtmp://54.254.183.160:19350/rtmp/157657392471017?s_bl=1&s_sc=157657419137681&s_sw=0&s_vt=api-s&a=AbxYo4VY6UA6aFsE",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("%s", err.Error())
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("%s", err.Error())
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("%s", err.Error())
	}
	populate_stdin_func(stdin)
	fo, _ := os.Create("output.mp3")
	io.Copy(fo, stdout)

	err = cmd.Wait()
	if err != nil {
		log.Printf("cmd failed to wait: %s", err.Error())
	}
}

// go func() {
// 	cmd := exec.Command("ffmpeg",
// 		"-i",
// 		"-",
// 		"-re",
// 		"-fflags",
// 		"+igndts",
// 		"-vcodec",
// 		"copy",
// 		"-acodec",
// 		"copy",
// 		"-preset",
// 		"ultrafast",
// 		"-crf",
// 		"22",
// 		"-b:a",
// 		"128K",
// 		"rtmp://54.254.183.160:19350/rtmp/157657392471017?s_bl=1&s_sc=157657419137681&s_sw=0&s_vt=api-s&a=AbxYo4VY6UA6aFsE")
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		log.Printf("%s", err.Error())
// 	}
// 	defer stdin.Close()

// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Printf("%s", err.Error())
// 	}

// 	err = cmd.Start()
// 	if err != nil {
// 		log.Printf("%s", err.Error())
// 	}

// 	io.Copy(stdin, bytes.NewReader(p.Raw))
// 	log.Printf("ffmpeg stdout: %v", stdout)

// 	err = cmd.Wait()
// 	if err != nil {
// 		log.Printf("cmd failed to wait: %s", err.Error())
// 	}
// }()
