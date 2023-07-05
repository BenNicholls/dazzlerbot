package main

import (
	"bufio"
	"errors"
	"flag"
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

	return nil
}

// attempts to connect to discord api using Api Key provided in configuration.
// if no session can be created, session will be nil.
func setupDiscordSession() (err error) {
	session, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		return
	}
	session.AddHandler(onMessage)
	err = session.Open()
	if err != nil {
		return
	}

	return nil
}

func main() {

	interactive := flag.Bool("i", false, "begin dazzlerbot in interactive mode. no discord session will be created.")
	flag.Parse()

	err := startup()
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	cliInput := make(chan string, 1)

	if *interactive {
		fmt.Println("Making interactive session")
		cli := bufio.NewScanner(os.Stdin)
		go func() {
			for cli.Scan(); cli.Text() != "exit"; cli.Scan() {
				cliInput <- cli.Text()
			}
			cliInput <- "exit"
			close(cliInput)
			return
		}()
	} else {
		fmt.Println("Initializing Discord session.")
		err = setupDiscordSession()
		if err != nil {
			fmt.Println("Could not initialize discord session:  ", err.Error())
			fmt.Println("Shutting down...")
			return
		} else {
			fmt.Println("Discord session online.")
		}

	}

	masterVoice.outputStats()
	fmt.Println("Bot is operational! Press Ctrl-C to shutdown.")
	running = true
	if *interactive {
		fmt.Print("CMD > ")
	}

	for running {
		select {
		case <-sc: //os level signal to close
			running = false
		case input := <-cliInput: //handle input from command line
			processCommand(strings.TrimSpace(input))
		}
	}

	if !*interactive {
		fmt.Println("Closing discord session.")
		session.Close()
	}

	fmt.Println("Smell ya later")
}

// process commands from the cli interface in interactive mode.
func processCommand(rawCommand string) {
	if rawCommand == "" {
		return
	}

	splitCommand := strings.Split(rawCommand, " ")
	command := splitCommand[0]
	args := make([]string, 0)
	if len(splitCommand) > 1 {
		args = splitCommand[1:]
	}

	switch command {
	case "help":
		fmt.Println("DAZZLERBOT COMMANDS:")
		fmt.Println(" speak              Generates a sentence.")
		fmt.Println(" stats              Prints the stats for the bot's current brain.")
		fmt.Println(" output             Outputs the brain. WARNING: for large brains, this takes FOREVER.")
		fmt.Println(" respond <phrase>   Responds to a phrase, interpreting the phrase as some kind of bot command.")
		fmt.Println(" help               Prints a mysterious menu")
		fmt.Println(" exit               Shuts down dazzlerbot.")
		fmt.Println("All other input will be added as a sentence into the current brain. This input is NOT recorded permanently.")
	case "stats":
		masterVoice.outputStats()
	case "output":
		masterVoice.output()
	case "speak":
		fmt.Println(masterVoice.Generate(config.SentenceLen))
	case "respond":
		fmt.Println(InterpretCommand(args))
	case "exit":
		running = false
		return
	default:
		masterVoice.AddString(command)
		fmt.Println("Added \"" + command + "\" to brain.")
	}

	fmt.Print("CMD > ")
}

// Callback that runs whenever a new message is sent in a server/channel that dazzlerbot has access to
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
	response := ""

	//detect special response modes.
	if strings.HasPrefix(strings.ToLower(m.Message.Content), "dazzlerbot") {
		//interpret as a command to dazzlerbot. of course this won't always be the case though.
		var commandString []string = strings.Split(m.Message.Content, " ")
		commandString = commandString[1:]
		if len(commandString) != 0{
			response = InterpretCommand(commandString)	
		}		
	}

	//search for triggers
	if response == "" {
		for _, word := range config.TriggerWords {
			if strings.Contains(strings.ToLower(m.Content), word) {
				response = masterVoice.Generate(config.SentenceLen)
				break
			}
		}
	}

	//if not triggered, generate messages randomly according to config.ResponseFrequency
	if response == "" && config.ResponseFrequency > 0 && rand.Intn(config.ResponseFrequency) == 0 {
		response = masterVoice.Generate(config.SentenceLen)
	}

	if response != "" {
		s.ChannelMessageSend(m.ChannelID, response)
	}
}
