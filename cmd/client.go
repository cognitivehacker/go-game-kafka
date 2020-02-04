package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/veandco/go-sdl2/sdl"
)

type Player struct {
	x      int32
	y      int32
	width  int32
	height int32
}

func (p Player) draw(window **sdl.Window) {
	w := *window
	pl1 := sdl.Rect{p.x, p.y, p.width, p.height}
	surface, err := w.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(&pl1, 0xffee0000)
	// w.UpdateSurface()
}

func randomPlayer() Player {
	return Player{
		x:      int32(rand.Intn(800)),
		y:      int32(rand.Intn(600)),
		width:  10,
		height: 10}
}

func main() {
	var window *sdl.Window
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	log.Println("Creating window")
	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	log.Println("Start running")
	running := true

	var players = []Player{}

	players = append(players, randomPlayer())
	players = append(players, randomPlayer())
	players = append(players, randomPlayer())
	players = append(players, randomPlayer())

	for running {
		// Clear window for each frame
		surface.FillRect(nil, 0)

		for _, p := range players {
			p.draw(&window)
		}
		window.UpdateSurface()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				// Modifier keys
				switch t.Keysym.Sym {
				case sdl.K_UP:
					if t.State == sdl.PRESSED {
						players[0].y -= 5
					}
				case sdl.K_DOWN:
					if t.State == sdl.PRESSED {
						players[0].y += 5
					}
				case sdl.K_LEFT:
					if t.State == sdl.PRESSED {
						players[0].x -= 5
					}
				case sdl.K_RIGHT:
					if t.State == sdl.PRESSED {
						players[0].x += 5
					}
				}
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func consumer() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"myTopic", "^aRegex.*[Tt]opic"}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}

func producer() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "myTopic"
	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
}
