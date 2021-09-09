package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

var session *discordgo.Session
var masterVoice Chain
var config configuration
var running bool

func startup() error {
	//load config
	configFile, err := os.Open("config.txt")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			werr := writeDefaultConfig()
			if werr != nil {
				return werr
			}

			return errors.New("No config file detected. A new config file config.txt has been created. You will have to put your discord api token there before dazzbot will work.")
		}
		return errors.New("Error opening config file: " + err.Error())
	}
	defer configFile.Close()

	data, err := ioutil.ReadAll(configFile)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return errors.New("Could not read config file. Could be a formatting problem. If this problem persists, delete config.txt and have run dazzbot again to generate a new default config file.")
	}

	//validate configuration
	if valid := config.validate(); valid != nil {
		return valid
	}

	//load archive and build markov chain
	masterVoice.init()
	err = loadArchive("archive")
	if err != nil {
		return errors.New("Could not load archive: " + err.Error())
	}

	//initialize discord connection
	session, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		return errors.New("Could not create discord session: " + err.Error())
	}
	session.AddHandler(onMessage)
	err = session.Open()
	if err != nil {
		return errors.New("Could not open discord connection: " + err.Error())
	}

	return nil
}

func main() {
	err := startup()
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	masterVoice.outputStats()

	fmt.Println("Bot is operational! Press Ctrl-C to shutdown.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	cli := bufio.NewScanner(os.Stdin)
	cliInput := make(chan string, 1)
	go func() {
		for cli.Scan(); cli.Text() != "exit"; cli.Scan() {
			cliInput <- cli.Text()
		}
		cliInput <- "exit"
		close(cliInput)
		return
	}()

	running = true
	fmt.Print("CMD > ")

	for running {
		select {
		case <-sc: //os level signal to close
			running = false
		case input := <-cliInput: //handle input from command line
			processCommand(strings.TrimSpace(input))
		}
	}

	session.Close()
	fmt.Println("Smell ya later")
}

//process commands from the cli interface. could also be used for the eventual processing of
//discord slash commands, if those ever make it in
func processCommand(command string) {
	if command == "" {
		return
	}

	switch command {
	case "help":
		fmt.Println("DAZZLERBOT COMMANDS:")
		fmt.Println(" speak     Generates a sentence.")
		fmt.Println(" stats     Prints the stats for the bot's current brain.")
		fmt.Println(" output    Outputs the brain. WARNING: for large brains, this takes FOREVER.")
		fmt.Println(" help      Prints a mysterious menu")
		fmt.Println(" exit      Shuts down dazzlerbot.")
		fmt.Println("All other input will be added as a sentence into the current brain. This input is NOT recorded permanently.")
	case "stats":
		masterVoice.outputStats()
	case "output":
		masterVoice.output()
	case "speak":
		fmt.Println(masterVoice.Generate(config.SentenceLen))
	case "exit":
		running = false
		return
	default:
		masterVoice.AddString(command)
		fmt.Println("Added \"" + command + "\" to brain.")
	}

	fmt.Print("CMD > ")
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	//keep dazzlerbot from responding to itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Message.Content != "" {
		//add new message to master voice
		writeStringToUserArchive(m.Author.ID, m.Message.Content)
		masterVoice.AddString(m.Message.Content)
	}

	//generate and send response (if necessary)
	message := ""

	//search for triggers
	for _, word := range config.TriggerWords {
		if strings.Contains(strings.ToLower(m.Content), word) {
			message = masterVoice.Generate(config.SentenceLen)
			break
		}
	}

	//if not triggered, generate messages randomly according to config.ResponseFrequency
	if message == "" && config.ResponseFrequency > 0 && rand.Intn(config.ResponseFrequency) == 0 {
		message = masterVoice.Generate(config.SentenceLen)
	}

	if message != "" {
		s.ChannelMessageSend(m.ChannelID, message)
	}
}
