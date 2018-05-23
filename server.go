package sshme

import (
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// Server is an ssh server.
type Server struct {
	Bind string
	Key  string
}

// Run starts the server.
func (s *Server) Run() {
	config := ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(s.privateKey())

	l, err := net.Listen("tcp", s.Bind)
	if err != nil {
		log.Fatalf(`failed to listen on "%s": %v`, s.Bind, err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}

		_, chs, rs, err := ssh.NewServerConn(c, &config)
		if err != nil {
			log.Printf("failed to handshake: %v", err)
			continue
		}

		go ssh.DiscardRequests(rs)
		go handleChannels(chs)
	}
}

func (s *Server) privateKey() ssh.Signer {
	key, err := ioutil.ReadFile(s.Key)
	if err != nil {
		log.Fatalf("failed to read key file %s: %v", s.Key, err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("failed to parse key: %v", err)
	}
	return signer
}

func handleChannels(chs <-chan ssh.NewChannel) {
	for ch := range chs {
		go handle(ch)
	}
}

func handle(ch ssh.NewChannel) {
	s, err := NewSession(ch)
	if err != nil {
		log.Printf("failed to create a new session: %v", err)
		return
	}
	defer s.Close()

	s.Run()
}
