package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dylanrclee/gator/internal/config"
)

type state struct {
	conpointer *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	list map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("No arguments provided")
	}
	err := s.conpointer.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User has been set to %s\n", cmd.arguments[0])
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	value, ok := c.list[cmd.name]
	if !ok {
		return fmt.Errorf("Command not found in command list")
	}
	err := value(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.list[name] = f
}

func main() {
	var mstate state

	read_result, err := config.Read()
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}
	mstate.conpointer = &read_result

	var mcommands commands
	mcommands.list = make(map[string]func(*state, command) error)

	mcommands.register("login", handlerLogin)

	userarguments := os.Args
	if len(userarguments) < 2 {
		log.Fatal("Less than 2 arguments provided")
		return
	}
	var mcommand command
	mcommand.name = userarguments[1]
	mcommand.arguments = userarguments[2:]
	err = mcommands.run(&mstate, mcommand)
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}
}
