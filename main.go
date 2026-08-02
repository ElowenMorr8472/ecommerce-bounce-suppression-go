package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

func main() {
	messageID := flag.String("message-id", "", "email message_id to audit")
	address := flag.String("email", "", "hard-bounced recipient address")
	flag.Parse()
	if *messageID == "" || *address == "" {
		log.Fatal("both -message-id and -email are required")
	}

	client, err := newClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	events, err := client.ListEvents(ctx, *messageID)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.AddHardBounce(ctx, *address); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("audited %s and suppressed %s\n", string(events.Data), *address)
}
