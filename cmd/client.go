package main

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
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
	surface.FillRect(nil, 0)
	log.Println("Start running")
	running := true
	x := 1
	y := 1
	for running {
		x += 1
		y += 1
		pl1 := sdl.Rect{x, y, 20, 20}
		pl2 := sdl.Rect{400, 400, 20, 20}
		surface.FillRect(&pl1, 0xffee0000)
		surface.FillRect(&pl2, 0xffeeff00)
		window.UpdateSurface()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				// keyCode := t.Keysym.Sym
				keys := ""
				// Modifier keys
				switch t.Keysym.Sym {
				case sdl.K_UP:
					if t.State == sdl.RELEASED {
						println("Foo Balr")
					}
					keys += "Left Alt"
				}
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}
