package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func WaitSig() {
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT)
	sig := <-c
	log.Println("Exit Signal:", sig)
}
