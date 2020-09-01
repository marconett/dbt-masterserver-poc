package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"strings"
)

var joinString = ""

func main() {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := "0.0.0.0:32100"
	listener, err := tls.Listen("tcp4", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		go handleClient(conn)
	}
}

func respond(conn net.Conn, s string) {
	_, err := conn.Write([]byte(s + "\n"))
	log.Printf("reponse: %s", s)

	if err != nil {
		log.Printf("reponse error: %s", err)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("server: conn: read: %s", err)
			break
		}

		request := string(buf[:n])
		log.Printf("request: %s", request)

		// server messages:
		// 2 == thx for playing banner -> quit
		// , == logged in from elsewhere -> quit
		// . == client is outdated -> quit
		// 1 == EGS verification failed -> quit
		// - == banned -> quit
		// 0 == ?
		// 4 == authed, load eggbot
		// # == authed, allowed to load maps
		// <space> == ?
		// ) == ?
		// % == start game with parameters (twice)
		// & == start game with parameters (twice)
		// / == start game with parameters
		// ' == start game with parameters (twice)

		// client messages:
		// ! == game launch
		// R == list custom games
		// Q == quit match
		// O == chat

		if strings.HasPrefix(request, "!") {
			go respond(conn, "4")
			go respond(conn, "#")

		} else if strings.HasPrefix(request, "O join") {
			// go respond(conn, "J s lobby-match-confirmed")
			go respond(conn, joinString)

		} else if strings.HasPrefix(request, "O start") {
			split := strings.Split(request, " ")
			if len(split) >= 4 {
				gameMode := split[2]
				gameMap := split[3]
				gameJSON := "{\"private\":false,\"name\":\"hehe\",\"mode\":\"" + gameMode + "\",\"game_mode\":\"" + gameMode + "\",\"map\":\"" + gameMap + "\",\"location\":\"lan\",\"lan_ip\":\"192.168.222.233\",\"colors\":[\"23c841\",\"8438f6\"],\"time_limit\":960,\"score_limit\":4,\"team_count\":2,\"team_size\":5,\"continuous\":1,\"intro\":0,\"team_switching\":2,\"physics\":0,\"spawn_logic\":1,\"warmup_time\":-1,\"min_players\":3,\"max_clients\":13,\"ready_percentage\":1,\"instagib\":false}"

				addr, ok := conn.RemoteAddr().(*net.TCPAddr)
				if ok == false {
					log.Printf("server: can't get client ip")
					break
				}
				gameServerIP := addr.IP.String()

				// the int after game mode set's self_host = true (clients hosts the game itself)
				startgame := "/ 00000000000000000000000000000000 " + gameServerIP + ":32123 " + gameMap + " " + gameMode + " 1 " + gameJSON

				joinString = "/ 00000000000000000000000000000000 " + gameServerIP + ":32123 " + gameMap + " " + gameMode + " 0 " + gameJSON
				go respond(conn, startgame)
			}
		}
	}
	log.Println("server: conn: closed")
}
