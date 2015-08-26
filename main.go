package main

import (
        "flag"
        "fmt"
        "log"
        "net"
        "os/exec"
        "sync/atomic"
)

func main() {
        flag.Usage = func() {
                fmt.Println("\ninet [-p port] [-m max-connections] [-c command]\n")
                flag.PrintDefaults()
                fmt.Println()
        }
        cmdflag := flag.String("c", "cat", "command")
        portflag := flag.String("p", "4321", "port number")
        maxflag := flag.Int64("m", 10, "max concurrent connections")
        flag.Parse()
        Serve(*portflag, *cmdflag, *maxflag)
}

func Serve(port string, command string, limit int64) error {
        ln, err := net.Listen("tcp", ":"+port)
        if err != nil {
                return err
        }

        var counter int64
        for {
                conn, err := ln.Accept()
                if err != nil {
                        log.Println("accept error ", err)
                        continue
                }
                log.Println("accept ", conn.RemoteAddr())
                if atomic.LoadInt64(&counter) >= limit {
                        conn.Close()
                        log.Println("too many connections")
                        log.Println("drop", conn.RemoteAddr())
                        continue
                }

                atomic.AddInt64(&counter, 1)
                cmd := exec.Command("sh", "-c", command)
                cmd.Stdout = conn
                cmd.Stdin = conn
                cmd.Stderr = conn
                cmd.Start()

                go func(cmd *exec.Cmd) {
                        defer func() {
                                conn.Close()
                                atomic.AddInt64(&counter, -1)
                                log.Println("close", conn.RemoteAddr())
                        }()
                        cmd.Wait()
                }(cmd)
        }

        log.Println("Server is shutting down")
        return nil
}
