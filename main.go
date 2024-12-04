package main

import (
	"github.com/Rejna/gator/internal/database"
	"github.com/Rejna/gator/internal/config"

	"fmt"
	"os"
	"errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
	"net/http"
	"encoding/xml"
	"io"
	"html"
	"strconv"
	"strings"
)

import _ "github.com/lib/pq"

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
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

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if handler, ok := c.handlers[cmd.name]; ok {
		return handler(s, cmd)
	}
	return errors.New("Unknown command: " + cmd.name)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("no username provided")
	}
	if _, err := s.db.GetUser(context.Background(), cmd.args[0]); err != nil {
		return errors.New("Username " + cmd.args[0] + " doesn't exist!")
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Println("username set succesfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("no username provided")
	}
	if _, err := s.db.CreateUser(context.Background(), database.CreateUserParams{uuid.New(), time.Now(), time.Now(), cmd.args[0]}); err != nil {
		return errors.New("User " + cmd.args[0] + " already exists")
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Println("User " + cmd.args[0] + " added successfully")
	user, _ := s.db.GetUser(context.Background(), cmd.args[0])
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return errors.New("Reset unsuccessful")
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return errors.New("List all users failed")
	}
	for _, user := range users {
		line := "* " + user.Name
		if user.Name == s.cfg.CurrentUserName {
			line += " (current)"
		}
		fmt.Println(line)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	req.Header.Add("User-Agent", "gator")
	if err != nil {
		return nil, fmt.Errorf("Bad request: %w", err)
	}
	var httpClient http.Client
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Bad response: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading body: %w", err)
	}
	var rssFeed RSSFeed
	if err := xml.Unmarshal(bodyBytes, &rssFeed); err != nil {
		return nil, fmt.Errorf("Error parsing XML body: %w", err)
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i := 0; i < len(rssFeed.Channel.Item); i++ {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}
	return &rssFeed, nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("not enough arguments")
	}
	timeBetweenReqsStr := cmd.args[0]
	timeBetweenReqs, err := time.ParseDuration(timeBetweenReqsStr)
	if err != nil {
		return fmt.Errorf("Error while parsing duration %s: %w", timeBetweenReqsStr, err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Errorf("Error while scraping feeds: %w", err)
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("Not enough arguments")
	}
	feedRecord := database.CreateFeedParams{uuid.New(), time.Now(), time.Now(), cmd.args[0], cmd.args[1], user.ID}
	feed, err := s.db.CreateFeed(context.Background(), feedRecord)
	if err != nil {
		return fmt.Errorf("Error creating feed %s (%s): %w", cmd.args[0], cmd.args[1], err)
	}
	feedFollowRecord := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	if _, err := s.db.CreateFeedFollow(context.Background(), feedFollowRecord); err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}
	fmt.Println(feedRecord)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting feeds: %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Feed name: %s | Feed URL: %s | Created by: %s\n", feed.FeedName, feed.FeedUrl, feed.Username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("not enough arguments")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error getting feed by url: %w", err)
	}

	feedFollowRecord := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	if _, err := s.db.CreateFeedFollow(context.Background(), feedFollowRecord); err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("Error getting feed follows for user %s: %w", user.Name, err)
	}
	fmt.Printf("Current user %s follows the feeds:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("* %s (%s)\n", feed.FeedName, feed.FeedUrl)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("not enough arguments")
	}

	if err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{user.ID, cmd.args[0]}); err != nil {
		return fmt.Errorf("Error deleting feed follow %s for user $s: %w", cmd.args[0], user.Name, err)
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32
	if len(cmd.args) == 0 {
		limit = 2
	} else {
		limitT, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("Error parsing limit %s: %w", cmd.args[0], err)
		}
		limit = int32(limitT)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{user.ID, limit})
	if err != nil {
		return fmt.Errorf("Error getting posts for user $s: %w", user.Name, err)
	}
	for _, post := range posts {
		fmt.Printf("* %s (%s), published %v\n", post.Title, post.Url, post.PublishedAt.Time)
		if post.Description.Valid {
			fmt.Printf("%s\n", strings.TrimSpace(post.Description.String))
		}
		fmt.Println("------------------")
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Error getting user %s: %w", s.cfg.CurrentUserName, err)
		}
		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error while getting next feed to fetch: %w", err)
	}
	if err := s.db.MarkFeedFetched(context.Background(), nextFeed.ID); err != nil {
		return fmt.Errorf("Error marking feed %s fetched: %w", nextFeed.Name, err)
	}
	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching the feed %s: %w", nextFeed.Url, err)
	}

	for _, feedItem := range feed.Channel.Item {
		if feedItem.Title != "" {
			pubDate, err := time.Parse(time.RFC1123Z, feedItem.PubDate)
			if err != nil {
				fmt.Errorf("Error parsing pubDate %s: %w", feedItem.PubDate, err)
			}
			pubDateSql := sql.NullTime{pubDate, feedItem.PubDate != ""}
			desc := sql.NullString{feedItem.Description, feedItem.Description != ""}
			params := database.CreatePostParams{uuid.New(), time.Now(), time.Now(), feedItem.Title, feedItem.Link, desc, pubDateSql, nextFeed.ID}
			if _, err := s.db.CreatePost(context.Background(), params); err != nil {
				fmt.Errorf("Error adding post: %w", err)
			}
		}
	}
	fmt.Printf("%s scraped\n", nextFeed.Name)

	return nil
}

func main() {
	conf := config.Read()
	db, err := sql.Open("postgres", conf.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s := state{dbQueries, &conf}
	cmds := commands{}
	cmds.handlers = map[string]func(*state, command) error{}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("feeds", handlerFeeds)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}
	commandName := args[1]
	commandArgs := args[2:]
	if err := cmds.run(&s, command{commandName, commandArgs}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}