package main

import (
	"bufio"
	"fmt"
	iofs "io/fs"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/devchat-ai/gopool"

	smb2 "github.com/hirochachacha/go-smb2"
)

func smbConnect(host string, port int) (error, error) {

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "TestUser",
			Password: "TestUser",
		},
	}

	log.Printf("Hello connecting ! %s:%d", host, port)
	target := net.JoinHostPort(strings.ReplaceAll(host, "\n", ""), "445")
	fmt.Println(target)
	time, _ := time.ParseDuration("1m")
	conn, err := net.DialTimeout("tcp", target, time)
	if err != nil {
		// panic(err)
		log.Printf("Initial Gateway closed for host: %s%d ", host, 445)
		return nil, nil
		_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, 139), time)
		if err != nil {
			log.Printf("Gateway closed for host: %s%d ", host, 139)
			return nil, nil

		} else {
			target = fmt.Sprintf("%s:%d", host, port)
			conn2, err2 := net.DialTimeout("tcp", target, time)
			if err2 != nil {
				log.Printf("Gateway Closed for host with conn 139 ", err2)
				return nil, nil
			}
			conn = conn2
		}
	}
	defer conn.Close()

	s, err := d.Dial(conn)
	if err != nil {
		log.Printf("Unable to dial host: %s ", host)
		return nil, nil
	}
	defer s.Logoff()
	names, err := s.ListSharenames()
	log.Printf("Shares Available %s", names)

	if err != nil {
		log.Printf("Unable to list shares for  host: %s ", host)
		return nil, nil
	}

	file, err := os.Create("./logs/" + host + ".log")
	if err != nil {
		log.Printf("Unable to create file for host " + host)
		return nil, nil

	}
	log.Printf("File created for host %s ", host)

	defer file.Close()
	w := bufio.NewWriter(file)
	for _, name := range names {

		fs, err := s.Mount(name)
		if err != nil {

			log.Printf("File mount failed for host %s %s", host, err)
			if err == os.ErrPermission {
				log.Printf("Permission denied %s", err)
				continue
			}
		}
		defer fs.Umount()
		fmt.Println("Probing share ", name)
		// matches, err := iofs.Glob(fs.DirFS("."), "*")
		// if err != nil {
		// 	log.Printf("Unable to glob fs for host %s " + host)
		// 	continue
		// }
		// for _, match := range matches {
		// 	log.Printf("%s", match)
		// }

		err = iofs.WalkDir(fs.DirFS("."), ".", func(path string, d iofs.DirEntry, err error) error {

			fmt.Fprintln(w, name+":/"+path)
			return nil
		})
		if err != nil {
			log.Printf("File walker failed for host %s %s", host, err)
			continue
		}
	}
	w.Flush()
	return nil, nil
}

// func startReaping(host string, port int) -> {

// }

func main() {

	// Read masscan file and probe for the smbconnection.

	content, err := os.ReadFile("test.txt")

	if err != nil {
		//Do something
		log.Printf("Unable to read input file... ")
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	// for _, host := range lines {
	// 	smbConnect(host, 445)
	// }

	pool := gopool.NewGoPool(1000)
	defer pool.Release()

	for _, host := range lines {
		fmt.Println(host)
		pool.AddTask(func() (interface{}, error) {
			return smbConnect(host, 445)

		})
	}

	pool.Wait()

	// smbConnect("175.101.19.102", 445, dialer)

}
