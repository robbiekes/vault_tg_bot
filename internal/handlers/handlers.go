package handlers

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	tele "gopkg.in/telebot.v4"
	"time"
)

func RunHandlers(b *tele.Bot, r *redis.Client) {
	ctx := context.Background()

	b.Handle("/greet", func(c tele.Context) error {
		return c.Send("Hello, beautiful :)")
	})

	b.Handle(tele.OnReply, func(c tele.Context) error {
		replyToText := c.Message().ReplyTo.Text
		timestamp := time.Now().Unix()
		key := fmt.Sprintf("%v:%v", c.Sender().ID, timestamp)
		err := r.Set(ctx, key, replyToText, 0).Err()
		if err != nil {
			fmt.Printf("error saving message to redis: %s", err.Error())
		}

		return nil
	})

	b.Handle("/pinned", func(c tele.Context) error {
		pinnedMessages, err := getAllReplies(ctx, r)
		if err != nil {
			return errors.WithMessage(err, "fail getting all pinned messages")
		}

		formattedList := "*🌟🌟🌟 Pinned messages 🌟🌟🌟*\n\n"
		for i, msg := range pinnedMessages {
			formattedList += fmt.Sprintf("%d. _%s_\n\n", i+1, msg)
		}

		chatID := tele.ChatID(c.Chat().ID)
		_, err = b.Send(chatID, formattedList, &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
		})
		if err != nil {
			return errors.WithMessage(err, "fail sending message")
		}

		return nil
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		timestamp := time.Now().Unix()
		key := fmt.Sprintf("%v:%v", c.Sender().ID, timestamp)
		err := r.Set(ctx, key, c.Text(), 0).Err()
		if err != nil {
			fmt.Printf("error saving message to redis: %s", err.Error())
		}

		return nil
	})
}

func getAllKeysAndValues(ctx context.Context, rdb *redis.Client) (map[string]string, error) {
	var cursor uint64
	keysValues := make(map[string]string)

	// loop until the cursor is 0 (indicating the end of the scan)
	for {
		// use the SCAN command to fetch keys
		var err error
		keys, newCursor, err := rdb.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			return nil, errors.WithMessage(err, "fail scanning keys")
		}

		for _, k := range keys {
			v, err := rdb.Get(ctx, k).Result()
			if err == redis.Nil {
				continue // key does not exist (shouldn't happen in SCAN results)
			} else if err != nil {
				return nil, errors.WithMessage(err, "fail getting keys and values")
			}
			keysValues[k] = v
		}

		cursor = newCursor

		// break the loop if the cursor is 0
		if cursor == 0 {
			break
		}
	}

	return keysValues, nil
}

func getAllReplies(ctx context.Context, r *redis.Client) ([]string, error) {
	keysValues, err := getAllKeysAndValues(ctx, r)
	if err != nil {
		return nil, errors.WithMessage(err, "fail getting replies")
	}

	replies := lo.MapToSlice(keysValues, func(k string, v string) string {
		return v
	})

	return replies, nil
}
