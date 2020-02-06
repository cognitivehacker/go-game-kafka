package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/veandco/go-sdl2/sdl"
)

type Player struct {
	Uuid   uuid.UUID
	X      int32
	Y      int32
	Width  int32
	Height int32
	Color  uint32
	Date   string
}

type KfkPlayerData struct {
	Uuid   uuid.UUID `json:"uuid"`
	X      int32     `json:"x"`
	Y      int32     `json:"y"`
	Width  int32     `json:"Width"`
	Height int32     `json:"Height"`
	Color  uint32    `json:"Color"`
	Date   string	 `json:"Date"`
}

func main() {

	log.Println("Start running")

	var wg sync.WaitGroup

	messages := make(chan KfkPlayerData, 1000)

	// INITIATE KAFKA CONSUMER
	wg.Add(1)
	go consumer(messages)

	// MAP OF PLAYERS
	players := make(map[uuid.UUID]*Player)

	// CREATE MY PLAYER
	var playerUUID uuid.UUID = uuid.New()

	player := Player{
		Uuid:   playerUUID,
		X:      int32(rand.Intn(800)),
		Y:      int32(rand.Intn(600)),
		Width:  10,
		Height: 10,
		Color:  0xffee0000}

	players[playerUUID] = &player

	go produce(player)

	wg.Add(1)
	go gameLoop(&players, playerUUID)

	// CONSUME KAFKA MESSAGES
	for {
		updatedData, ok := <-messages
		if ok == false {
			break
		}

		fmt.Printf("Updated Data: %+v\n", updatedData)
		// Player already in map
		if val, ok := players[updatedData.Uuid]; ok {
			fmt.Printf("Player ALREADY Exists %v \n", updatedData.Uuid)
			println("", val)
			player := players[updatedData.Uuid]
			player.Uuid = updatedData.Uuid
			player.X = updatedData.X
			player.Y = updatedData.Y
			player.Width = updatedData.Width
			player.Height = updatedData.Height
			player.Color = updatedData.Color

		} else {
			// Player is new
			fmt.Printf("Player IS NEW %v \n", updatedData.Uuid)
			players[updatedData.Uuid] = &Player{
				Uuid:   updatedData.Uuid,
				X:      updatedData.X,
				Y:      updatedData.Y,
				Width:  updatedData.Width,
				Height: updatedData.Height,
				Color:  updatedData.Color}
		}
	}
	wg.Wait()
}

func gameLoop(players *map[uuid.UUID]*Player, playerUUID uuid.UUID) {
	// Create Game window
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
	// GAME LOOP
	running := true
	for running {
		// Clear window for each frame
		surface.FillRect(nil, 0)

		// DRAW PLAYERS
		for _, p := range *players {
			p.draw(&window)
		}
		window.UpdateSurface()

		myPlayer := (*players)[playerUUID]

		// KEYBOARD EVENTS
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				// Modifier keys
				switch t.Keysym.Sym {
				case sdl.K_UP:
					if t.State == sdl.PRESSED {
						myPlayer.Y -= 5
					}
				case sdl.K_DOWN:
					if t.State == sdl.PRESSED {
						myPlayer.Y += 5
					}
				case sdl.K_LEFT:
					if t.State == sdl.PRESSED {
						myPlayer.X -= 5
					}
				case sdl.K_RIGHT:
					if t.State == sdl.PRESSED {
						myPlayer.X += 5
					}
				}

				// PRODUCE KAFKA MESSAGE
				go produce(*myPlayer)
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func (p Player) draw(window **sdl.Window) {
	w := *window
	pl1 := sdl.Rect{p.X, p.Y, p.Width, p.Height}
	surface, err := w.GetSurface()
	if err != nil {
		panic(err)
	}
	if p.Color == 0 {
		p.Color = 0xffee0000
	}
	surface.FillRect(&pl1, p.Color)
}

func randomPlayer() Player {
	return Player{
		X:      int32(rand.Intn(800)),
		Y:      int32(rand.Intn(600)),
		Width:  10,
		Height: 10,
		Color:  0xffee0000}
}

func consumer(messages chan<- KfkPlayerData) {
	// get kafka reader using environment variables.
	kafkaURL := "localhost:29092"
	topic := "mytopic"
	partition := 0

	reader := getKafkaReader(kafkaURL, topic, partition)

	defer reader.Close()

	fmt.Println("start consuming ... !!")
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		var response KfkPlayerData
		json.Unmarshal([]byte(m.Value), &response)

		fmt.Printf("consumer response: %+v\n\n#################\n", response)
		messages <- response
	}
}

func produce(player Player) {

	player.Date = time.Now().String()
	playerJson, err := json.Marshal(player)

	if err != nil {
		fmt.Println(err)
		return
	}

	kafkaURL := "localhost:29092"
	topic := "mytopic"
	writer := newKafkaWriter(kafkaURL, topic)
	defer writer.Close()
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("Key-%d", player.Uuid)),
		Value: []byte(fmt.Sprint(string(playerJson))),
	}
	err2 := writer.WriteMessages(context.Background(), msg)
	if err2 != nil {
		fmt.Println(err2)
	}
	time.Sleep(1 * time.Second)
}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func getKafkaReader(kafkaURL, topic string, partition int) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Partition:  partition,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		ReadBackoffMin: time.Microsecond * 1,
		ReadBackoffMax: time.Microsecond * 2,
		ReadLagInterval: -1,

	})
}
