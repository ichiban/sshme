package sshme

import (
	"errors"
	"fmt"
	"log"

	"bufio"

	"bytes"

	"encoding/binary"

	"io"

	"golang.org/x/crypto/ssh"
)

// Session is a session.
type Session struct {
	channel  ssh.Channel
	requests <-chan *ssh.Request
	width    int
	height   int
}

// NewSession creates a session.
func NewSession(ch ssh.NewChannel) (*Session, error) {
	if t := ch.ChannelType(); t != "session" {
		msg := fmt.Sprintf("unknown channel type: %s", t)
		if err := ch.Reject(ssh.UnknownChannelType, msg); err != nil {
			log.Printf("failed to reject channel: %v", err)
			return nil, err
		}
		return nil, errors.New(msg)
	}

	c, rs, err := ch.Accept()
	if err != nil {
		log.Printf("failed to accept channel: %v", err)
		return nil, err
	}

	return &Session{
		channel:  c,
		requests: rs,
		width:    80,
		height:   25,
	}, nil
}

// Run executes a session
func (s *Session) Run() {
	go s.handleRequests()
	s.run()
}

func (s *Session) handleRequests() {
	for {
		req := <-s.requests
		if req == nil {
			return
		}

		switch req.Type {
		case "pty-req":
			s.handlePtyReq(req)
		case "window-change":
			s.handleWindowChange(req)
		}
	}
}

func (s *Session) handlePtyReq(req *ssh.Request) {
	r := bytes.NewReader(req.Payload)
	if _, err := parseTerm(r); err != nil {
		log.Printf("failed to parse term: %v", err)
		return
	}
	var err error
	if s.width, s.height, err = parseSize(r); err != nil {
		log.Printf("failed to parse size: %v", err)
		return
	}
	if req.WantReply {
		if err := req.Reply(true, nil); err != nil {
			log.Printf("failed to reply: %v", err)
		}
	}
}

func (s *Session) handleWindowChange(req *ssh.Request) {
	r := bytes.NewReader(req.Payload)
	var err error
	if s.width, s.height, err = parseSize(r); err != nil {
		log.Printf("failed to parse size: %v", err)
		return
	}
	if req.WantReply {
		if err := req.Reply(true, nil); err != nil {
			log.Printf("failed to reply: %v", err)
			return
		}
	}
	if err := s.render(); err != nil {
		log.Printf("failed to render: %v", err)
	}
}

func parseTerm(r io.Reader) (string, error) {
	var termLen uint32
	if err := binary.Read(r, binary.BigEndian, &termLen); err != nil {
		return "", err
	}
	term := make([]byte, termLen)
	if _, err := r.Read(term); err != nil {
		return "", err
	}
	return string(term), nil
}

func parseSize(r io.Reader) (int, int, error) {
	var width, height uint32
	if err := binary.Read(r, binary.BigEndian, &width); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(r, binary.BigEndian, &height); err != nil {
		return 0, 0, err
	}
	return int(width), int(height), nil
}

func (s *Session) run() {
	if err := s.render(); err != nil {
		log.Printf("failed to render: %v", err)
		return
	}

	in := bufio.NewReader(s.channel)
	_, _, err := in.ReadRune()
	if err != nil {
		log.Printf("failed to read rune: %v", err)
		return
	}
}

// Close closes a session.
func (s *Session) Close() {
	if err := s.channel.Close(); err != nil {
		log.Printf("failed to close: %v", err)
	}
}

func (s *Session) render() error {
	w := bufio.NewWriter(s.channel)

	parts := []string{
		"\033[2J", // Erase Screen
		"\033[H",  // Cursor Home
		"╔══════════════════════════════════════════════════════════════════════════════╗\r\n",
		"║                                                                              ║\r\n",
		"║                                                                              ║\r\n",
		"║     \033[1mYutaka Ichibangase\033[0m                                                       ║\r\n",
		"║                                                                              ║\r\n",
		"║                                                                              ║\r\n",
		"║     I'm a software developer specialized in \033[4msolving business problems\033[0m.       ║\r\n",
		"║     Since I moved to \033[4mTokyo, Japan\033[0m, I've worked on \033[4mbusiness process \033[0m          ║\r\n",
		"║     \033[4mautomation\033[0m for offices, \033[4mcontent creation platform\033[0m for a content          ║\r\n",
		"║     writing business, \033[4msocial media marketing tool\033[0m for international          ║\r\n",
		"║     fashion brands, and so on.                                               ║\r\n",
		"║                                                                              ║\r\n",
		"║                                                                              ║\r\n",
		"║                 GitHub  https://github.com/ichiban                           ║\r\n",
		"║                Twitter  https://twitter.com/1ban                             ║\r\n",
		"║              AngelList  https://angel.co/yutaka-ichibangase                  ║\r\n",
		"║               LinkedIn  https://www.linkedin.com/in/yutakaichibangase        ║\r\n",
		"║                   Blog  http://www.y1ban.com/                                ║\r\n",
		"║                  Email  yichiban@gmail.com                                   ║\r\n",
		"║                                                                              ║\r\n",
		"║                                                                              ║\r\n",
		"║     Press any key to quit.                                                   ║\r\n",
		"║                                                                              ║\r\n",
		"╚══════════════════════════════════════════════════════════════════════════════╝",
		"\033[0m", // Reset all attributes
	}

	for _, p := range parts {
		if _, err := fmt.Fprint(w, p); err != nil {
			return err
		}
	}

	return w.Flush()
}
