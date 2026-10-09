package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/dylanrclee/gator/internal/config"
	"github.com/dylanrclee/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db         *database.Queries
	conpointer *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	list map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("No arguments provided\n")
	}

	_, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}

	err = s.conpointer.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User has been set to %s\n", cmd.arguments[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("No arguments provided\n")
	}

	cur_time := time.Now()
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: cur_time,
		UpdatedAt: cur_time,
		Name:      cmd.arguments[0],
	}

	_, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err == nil {
		return fmt.Errorf("Name %s already exists", cmd.arguments[0])
	}
	if err != sql.ErrNoRows {
		return err
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return err
	}

	err = s.conpointer.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %s was successfully created", cmd.arguments[0])
	log.Printf("%+v", user)
	return nil
}

func handlerDeleteAll(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("To many arguments provided\n")
	}

	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("Database was successfully reset\n")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == s.conpointer.Username {
			fmt.Printf("%s (current)", user)
		} else {
			fmt.Printf("%s\n", user)
		}
	}
	return nil
}

func agg(s *state, cmd command) error {
	res_feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", res_feed)
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var res_feed RSSFeed
	err = xml.Unmarshal(body, &res_feed)
	if err != nil {
		return nil, err
	}

	res_feed.Channel.Title = html.UnescapeString(res_feed.Channel.Title)
	res_feed.Channel.Description = html.UnescapeString(res_feed.Channel.Description)
	for _, item := range res_feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
	return &res_feed, err
}

func main() {
	var mstate state

	read_result, err := config.Read()
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}
	mstate.conpointer = &read_result

	db, err := sql.Open("postgres", config.Postgresfilepath)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	dbQueries := database.New(db)
	mstate.db = dbQueries

	var mcommands commands
	mcommands.list = make(map[string]func(*state, command) error)
	mcommands.register("login", handlerLogin)
	mcommands.register("register", handlerRegister)
	mcommands.register("reset", handlerDeleteAll)
	mcommands.register("users", handlerUsers)
	mcommands.register("agg", agg)

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
