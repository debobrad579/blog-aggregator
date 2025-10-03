package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/debobrad579/blog-aggregator/internal/config"
	"github.com/debobrad579/blog-aggregator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	cmdFunc, ok := c.commands[cmd.name]
	if !ok {
		return errors.New("Command '" + cmd.name + "' does not exist")
	}

	if err := cmdFunc(s, cmd); err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("There was an error reading the config file")
		os.Exit(1)
	}

	s := state{cfg: &cfg}

	cmds := commands{commands: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddfeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Please enter a command")
		os.Exit(1)
	}

	cmdName := args[1]
	var cmdArgs []string

	if len(args) > 2 {
		cmdArgs = args[2:]
	}

	db, err := sql.Open("postgres", s.cfg.DBUrl)
	if err != nil {
		fmt.Println("There was an error opening the database")
		os.Exit(1)
	}

	s.db = database.New(db)

	if err := cmds.run(&s, command{name: cmdName, args: cmdArgs}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
