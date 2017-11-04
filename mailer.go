package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Configuration struct {
	Subject      string   `json:"subject"`
	Sender       string   `json:"sender"`
	Domain       string   `json:"domain"`
	SMTPHost     string   `json:"smtpHost"`
	Recepients   []string `json:"recepients"`
	Team1Users   string   `json:"team1Users"`
	Team2Users   string   `json:"team2Users"`
	BodyFilePath string   `json:"bodyFilePath"`
}

var (
	confFile = kingpin.Flag(
		"config",
		"Configuration for Mailer",
	).Required().String()

	config Configuration
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	configFile, err := os.Open(*confFile)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	sendMail()
}

func getEmail(user string) string {
	return user + "@" + config.Domain
}

func getEmails(users []string) string {
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = getEmail(u)
	}
	return strings.Join(emails, ",")
}

func sendMail() {
	m := gomail.NewMessage()
	m.SetHeader("From", getEmail(config.Sender))
	m.SetHeader("To", getEmails(config.Recepients))
	user1 := getFirstEntryAndRotate(config.Team1Users)
	user2 := getFirstEntryAndRotate(config.Team2Users)

	m.SetAddressHeader("Cc", getEmail(user1), getEmail(user1))
	m.SetAddressHeader("Cc", getEmail(user2), getEmail(user2))

	m.SetHeader("Subject", fmt.Sprintf(config.Subject, user1, user2 ) )
	m.SetBody("text/html", slurpFile(config.BodyFilePath))

	pass := os.Getenv("SENDER_PASSWORD")
	d := gomail.NewDialer(config.SMTPHost, 587, getEmail(config.Sender), pass)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func getFirstEntryAndRotate(fPath string) string {
	tmpPath := fPath + ".tmp"
	f, err := os.Open(fPath)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	fo, err := os.Create(tmpPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	w := bufio.NewWriter(fo)
	for _, line := range lines[1:] {
		fmt.Println(line)
		w.WriteString(line + "\n")
		w.Flush()
	}

	w.WriteString(lines[0] + "\n")
	w.Flush()

	if stat, err := fo.Stat(); err != nil && stat.Size() < 1 {
		log.Fatal(err)
	}
	os.Rename(tmpPath, fPath)
	return lines[0]
}

func slurpFile(fPath string) string {
	b, err := ioutil.ReadFile(fPath) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	return string(b)

}
