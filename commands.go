package main

import "strings"

//signature for command functions. all command functions take a list of arguments and return a generated string.
type commandFunc func([]string) string

//map commands to the corresponding functions
var commandMap map[string]commandFunc

func init() {
	commandMap = make(map[string]commandFunc)
	commandMap["what"] = WhatCommand
	commandMap["who"] = WhatCommand
}

// figures out what kind of command has been given, if any
// shuttles the command arguments to the command function
func InterpretCommand(command []string) (response string) {
	if commandfunc, ok := commandMap[command[0]]; ok {
		response = commandfunc(command[1:])
	}
	return
}

//Handles "What is" or "What are" queries.
//Ex: INPUT: dazzlerbot, what is love?
//    OUTPUT: Love is incredible!
func WhatCommand(args []string) (response string) {
	if (strings.ToLower(args[0]) == "is" || strings.ToLower(args[0]) == "are") && len(args) > 1 {
		var responsePrefix []string = make([]string, 0)
		for _, word := range args[1:] {
			responsePrefix = append(responsePrefix, strings.Trim(word, "?!.\"()"))
		}
		responsePrefix = append(responsePrefix, args[0])

		response = masterVoice.GenerateWithPrefix(config.SentenceLen, responsePrefix)
	}
	return
}
