package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/veandco/go-sdl2/sdl"
)

type Player struct {
	uuid   uuid.UUID
	x      int32
	y      int32
	width  int32
	height int32
	color  uint32
}

func (p Player) draw(window **sdl.Window) {
	w := *window
	pl1 := sdl.Rect{p.x, p.y, p.width, p.height}
	surface, err := w.GetSurface()
	if err != nil {
		panic(err)
	}
	if p.color == 0 {
		p.color = 0xffee0000
	}
	surface.FillRect(&pl1, p.color)
	// w.UpdateSurface()
}

func randomPlayer() Player {
	return Player{
		x:      int32(rand.Intn(800)),
		y:      int32(rand.Intn(600)),
		width:  10,
		height: 10,
		color:  0xffee0000}
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

	// MAP OF PLAYERS
	players := make(map[uuid.UUID]*Player)
	var playerUUID uuid.UUID = uuid.New()
	players[playerUUID] = &Player{
		uuid:   playerUUID,
		x:      int32(rand.Intn(800)),
		y:      int32(rand.Intn(600)),
		width:  10,
		height: 10,
		color:  0xffee0000}

	// GAME LOOP
	for running {
		// Clear window for each frame
		surface.FillRect(nil, 0)

		for _, p := range players {
			p.draw(&window)
		}
		window.UpdateSurface()

		player := players[playerUUID]

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				// Modifier keys
				switch t.Keysym.Sym {
				case sdl.K_UP:
					if t.State == sdl.PRESSED {
						player.y -= 5
					}
				case sdl.K_DOWN:
					if t.State == sdl.PRESSED {
						player.y += 5
					}
				case sdl.K_LEFT:
					if t.State == sdl.PRESSED {
						player.x -= 5
					}
				case sdl.K_RIGHT:
					if t.State == sdl.PRESSED {
						player.x += 5
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
	// get kafka reader using environment variables.
	kafkaURL := "localhost:29092"
	topic := "mytopic"
	groupID := "groupID"

	reader := getKafkaReader(kafkaURL, topic, groupID)

	defer reader.Close()

	fmt.Println("start consuming ... !!")
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}

func producer() {
	kafkaURL := "localhost:29092"
	topic := "mytopic"
	writer := newKafkaWriter(kafkaURL, topic)
	defer writer.Close()
	fmt.Println("start producing ... !!")
	for i := 0; ; i++ {
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("Key-%d", i)),
			Value: []byte(fmt.Sprint(uuid.New())),
		}
		err := writer.WriteMessages(context.Background(), msg)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
