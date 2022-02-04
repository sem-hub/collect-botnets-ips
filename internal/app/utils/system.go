package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func AppendFile(file string, line string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fmt.Fprintln(f, line)
	f.Close()

	return nil
}

func RewriteFile(file string, content []string) error {
	tempFileName := "/tmp/" + path.Base(file) + "." + randString(6)
	log.Print("Temp file: " + tempFileName)
	f, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	for _, line := range content {
		fmt.Fprintln(f, line)
	}
	f.Close()

	err = os.Rename(tempFileName, file)
	if err != nil {
		log.Print("Rename error")
	}

	return nil
}

func OsExec(cmd string, args string) ([]byte, error) {
	out, err := exec.Command(cmd, strings.Fields(args)...).CombinedOutput()
	if err != nil {
		return out, err
	}
	return nil, nil
}
