package main

import (
	//"bytes"
	"fmt"
	"log"
	"time"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

type Hub struct {
	Recv      chan clientMsg
	Me        kitp.UserID
	journal   *journal.Journal
	clients   map[*client]bool
	subscribe chan *client
	//unsubscribe chan *client
}

type clientMsg struct {
	client *client
	msg    *kitp.Raw
}

type client struct {
	user        kitp.UserID
	sync        kitp.Sync
	follow      bool
	send        chan []*kitp.Raw
	pending     []*kitp.Raw
	delayed     *kitp.Raw
	delayedTime time.Time
}

func NewHub(dir string) (*Hub, error) {
	journal, err := journal.Open(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open journal: %v", err)
	}
	if len(journal.Messages()) == 0 {
		return nil, fmt.Errorf("not initialized")
	}
	hub := &Hub{
		Recv:      make(chan clientMsg, 100),
		Me:        journal.Messages()[0].User,
		journal:   journal,
		clients:   make(map[*client]bool),
		subscribe: make(chan *client, 10),
		//unsubscribe:make(chan *client, 1),
	}
	return hub, err
}

func (hub *Hub) Register(user kitp.UserID, sync kitp.Sync) *client {
	client := &client{
		user:   user,
		sync:   sync,
		follow: sync.Flags&kitp.SyncFollow != 0,
		send:   make(chan []*kitp.Raw, 2),
	}
	hub.subscribe <- client
	return client
}

func (hub *Hub) Unregister(c *client) {
	if c.follow {
		hub.subscribe <- c
	}
}

func (hub *Hub) GetCursor(user *kitp.UserID) *kitp.MsgID {
	// TODO: ok to access concurrnetly?
	cursor, err := hub.journal.PeerCursor(*user)
	if err != nil {
		log.Printf("failed to read peer cursor: %v", err)
	}
	return cursor
}

func (hub *Hub) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case cm := <-hub.Recv:
			// TODO: who does copies? where?
			if ok, err := hub.journal.AppendDup(cm.msg, &cm.client.user); err != nil {
				log.Printf("failed to append msg: %v\n%v", err, cm.msg)
			} else if !ok {
				break
			}
			for client := range hub.clients {
				if client == cm.client /*|| bytes.Equal(cm.msg.ID[:], client.user[:])*/ {
					if client.delayed == nil {
						client.delayedTime = time.Now()
					}
					client.delayed = cm.msg
					continue
				}
				client.delayed = nil
				client.pending = append(client.pending, cm.msg)
				select {
				case client.send <- client.pending:
					client.pending = nil
				default:
				}
			}
		case client := <-hub.subscribe:
			if hub.clients[client] {
				if !client.follow {
					panic("unsubscribing non-following client")
				}
				delete(hub.clients, client)
				close(client.send)
				break
			}
			msgs := hub.journal.SelectCursor(client.user, client.sync)
			if len(msgs) != 0 {
				client.send <- msgs
			}
			if client.follow {
				hub.clients[client] = true
			} else {
				close(client.send)
			}
		case <-ticker.C:
			for client := range hub.clients {
				if client.delayed != nil && time.Since(client.delayedTime) > 3*time.Second {
					client.pending = append(client.pending, client.delayed)
					client.delayed = nil
				}
				if len(client.pending) == 0 {
					continue
				}
				select {
				case client.send <- client.pending:
					client.pending = nil
				default:
				}
			}
		}
	}
}
