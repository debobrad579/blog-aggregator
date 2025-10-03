package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/debobrad579/blog-aggregator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Please enter a username")
	}

	user, err := s.db.GetUser(context.Background(), cmd.args[0])

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("User not found")
		} else {
			return err
		}
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return err
	}

	fmt.Println("User has been set to '" + user.Name + "'")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Please enter a username")
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return errors.New("User '" + cmd.args[0] + "' already exists")
		}
		return err
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return err
	}

	fmt.Println("User '" + user.Name + "' has been created")

	return nil
}

func handlerReset(s *state, _ command) error {
	if err := s.db.Reset(context.Background()); err != nil {
		return err
	}

	fmt.Println("Reset successful")
	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		currentText := ""
		if user.Name == s.cfg.CurrentUserName {
			currentText = " (current)"
		}

		fmt.Println("* " + user.Name + currentText)
	}

	return nil
}

func handlerAddfeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("Please enter a name and a url")
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID})

	if err != nil {
		return err
	}

	fmt.Printf("Feed '%s' ('%s') has been added\n", feed.Name, feed.Url)

	handlerFollow(s, command{name: "follow", args: []string{feed.Url}}, user)

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("* %s (%s) by %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("Please enter a url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}

	fmt.Printf("User '%s' is now following feed '%s'\n", feedFollow.UserName, feedFollow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFollowingFeeds(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println("* " + feed.Name)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("Please enter a url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	if err := s.db.Unfollow(context.Background(), database.UnfollowParams{UserID: user.ID, FeedID: feed.ID}); err != nil {
		return err
	}

	fmt.Printf("User '%s' unfollowed feed '%s'", user.Name, feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32 = 2

	if len(cmd.args) != 0 {
		i64, err := strconv.ParseInt(cmd.args[0], 10, 32)

		if err != nil {
			return errors.New("Please enter an integer limit")
		}

		limit = int32(i64)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: limit})
	if err != nil {
		return err
	}

	for i, post := range posts {
		fmt.Println("Title: " + post.Title)
		fmt.Println("Description: " + post.Description)
		fmt.Printf("Publish Date: %v\n", post.PublishedAt)
		fmt.Println("Link: " + post.Url)

		if i != len(posts)-1 {
			fmt.Println("")
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Please enter a time between requests")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return errors.New("Please enter a valid time (1s, 1m, 1h, etc.)")
	}

	ticker := time.NewTicker(timeBetweenReqs)

	for ; ; <-ticker.C {
		feed, err := s.db.GetNextFeedToFetch(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}

		if err := s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{UpdatedAt: time.Now(), ID: feed.ID}); err != nil {
			fmt.Println(err)
			continue
		}

		rssFeed, err := fetchFeed(context.Background(), feed.Url)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, item := range rssFeed.Channel.Item {
			pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
			if err != nil {
				fmt.Println(err)
				pubDate = time.Now()
			}

			post, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       item.Title,
				Url:         item.Link,
				Description: item.Description,
				PublishedAt: pubDate,
				FeedID:      feed.ID,
			})
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					continue
				}

				fmt.Println("unexpected error inserting post:", err)
				continue
			}

			fmt.Println("Created post '" + post.Title + "'")
		}
	}
}
