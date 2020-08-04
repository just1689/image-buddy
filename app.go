package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const PULL = "PULL"
const PROMPT = "PROMPT"
const TAG = "TAG"
const PUSH = "PUSH"

var imagesFile = flag.String("images", "images", "the images file list to use")
var buddyFile = flag.String("buddyFile", "buddy.yaml", "the yaml file use for configuration")

type Handler func()

var c *Config

var ActionMap = map[string]Handler{
	PULL:   PullHandler,
	TAG:    TagHandler,
	PUSH:   PushHandler,
	PROMPT: PromptHandler,
}

func main() {
	flag.Parse()
	logrus.Println("Starting...")
	b, err := ioutil.ReadFile(*buddyFile)
	if err != nil {
		logrus.Panic(err)
	}
	c = &Config{}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		logrus.Panic(err)
	}
	for _, action := range c.Actions {
		f, found := ActionMap[action]
		if !found {
			logrus.Panic("could not understand action: ", action)
		}
		f()
	}
}

func PullHandler() {
	files := FileReader(*imagesFile)
	for f := range files {
		s := "docker pull " + f
		fmt.Println(s)
		Cmd(s, true)
	}
}

func TagHandler() {
	files := FileReader(*imagesFile)
	for f := range files {
		s := "docker tag " + f + " " + c.Registry + "/" + f
		fmt.Println(s)
		Cmd(s, true)
	}
}

func PushHandler() {
	files := FileReader(*imagesFile)
	for f := range files {
		s := "docker push " + c.Registry + "/" + f
		fmt.Println(s)
		Cmd(s, true)
	}
}

func PromptHandler() {
	logrus.Println("Waiting for user - press enter")
	bufio.NewReader(os.Stdin).ReadRune()
}

func FileReader(filename string) chan string {
	result := make(chan string)
	go func() {
		defer close(result)
		f, err := os.Open(filename)
		if err != nil {
			logrus.Errorln("error opening file ", err)
			return
		}
		defer f.Close()
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString(10) // 0x0A separator = newline
			if line != "" {
				line = strings.TrimSuffix(line, "\n")
				result <- line
			}
			if err == io.EOF {
				break
			}
		}
	}()
	return result
}

func Cmd(cmd string, shell bool) []byte {
	if shell {
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			logrus.Errorln(err)
			logrus.Errorln(string(out))
			panic("some error found")
		}
		return out
	} else {
		out, err := exec.Command(cmd).Output()
		if err != nil {
			logrus.Errorln(err)
			logrus.Errorln(string(out))
			panic("some error found")
		}
		return out
	}
}

type Config struct {
	Registry string   `yaml:"registry"`
	Actions  []string `yaml:"actions"` // From: PULL, PROMPT, TAG, PUSH
}
