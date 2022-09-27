package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/tikinang/sims/list"

	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	seed  int64
	speed time.Duration
}

func main() {

	var c Config
	var logFile string
	flag.StringVar(&logFile, "log-file", "log", "log file path")
	flag.Int64Var(&c.seed, "seed", 0, "seed of the simulation")
	flag.DurationVar(&c.speed, "speed", 100*time.Millisecond, "time between each beat, speed of the simulation")
	flag.Parse()

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)

	w := &World{
		config:   c,
		noise:    rand.New(rand.NewSource(c.seed)),
		entities: list.NewLinkedList[Entity](),
	}

	log.Println("----- starting simulation -----")
	p := tea.NewProgram(w, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		panic(err)
	}
	log.Println("----- simulation ended -----")
}

type World struct {
	config   Config
	noise    *rand.Rand
	age      uint64
	entities *list.LinkedList[Entity]
	w, h     int
	render   bool
}

func (r *World) Resize(w, h int) {
	r.w, r.h = w, h
	r.render = true
}

func (r *World) Beat() {
	r.age++
	r.entities.IterateRemove(func(e Entity) bool {
		if !e.Beat() {
			return true
		}
		p := e.Position()
		if p.x >= r.w || p.y >= r.h {
			return true
		}
		return false
	})
	if r.noise.Intn(SlugSpawnChance) == 0 {
		r.entities.PushBack(&Slug{
			position: PreciseCoordinates{
				x: float64(r.noise.Intn(r.w)),
				y: float64(r.noise.Intn(r.h)),
			},
			movementDirection: Direction(r.noise.Intn(4)),
		})
	}
}

type Entity interface {
	Beat() bool
	Render() rune
	Position() Coordinates
}

type Direction uint

const (
	North Direction = iota
	East
	South
	West
)

type Coordinates struct{ x, y int }
type PreciseCoordinates struct{ x, y float64 }

func (r PreciseCoordinates) Move(direction Direction, distance float64) PreciseCoordinates {
	switch direction {
	case North:
		r.y -= distance
	case East:
		r.x += distance
	case South:
		r.y += distance
	case West:
		r.x -= distance
	default:
		panic("invalid direction")
	}
	return r
}

type TickMsg time.Time

func (r *World) TickCmd() tea.Cmd {
	return tea.Tick(r.config.speed, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (r *World) Init() tea.Cmd {
	return r.TickCmd()
}

func (r *World) Update(raw tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := raw.(type) {
	case tea.WindowSizeMsg:
		r.Resize(msg.Width, msg.Height)
		return r, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		default:
			log.Printf("update: [key] %s\n", msg.String())
			return r, nil
		}
	case TickMsg:
		r.Beat()
		return r, r.TickCmd()
	default:
		log.Printf("update: [%T] %v\n", raw, raw)
		return r, nil
	}
}

const (
	RenderEmpty   = ' '
	RenderNewline = '\n'
)

func (r *World) View() string {
	if !r.render {
		return "not initialized"
	}

	canvas := make([][]rune, 0, r.h)
	for i := 0; i < r.h; i++ {
		canvas = append(canvas, make([]rune, 0, r.w))
		for j := 0; j < r.w; j++ {
			canvas[i] = append(canvas[i], RenderEmpty)
		}
	}

	r.entities.Iterate(func(e Entity) {
		p := e.Position()
		canvas[p.y][p.x] = e.Render()
	})

	s := new(strings.Builder)
	for _, row := range canvas {
		for _, cell := range row {
			s.WriteRune(cell)
		}
		s.WriteRune(RenderNewline)
	}
	return strings.TrimSuffix(s.String(), "\n")
}

const (
	SlugSpeed        float64 = 0.12
	SlugAgeThreshold uint64  = 256
	SlugRender               = 'S'
	SlugSpawnChance          = 32
)

type Slug struct {
	age               uint64
	position          PreciseCoordinates
	movementDirection Direction
}

func (r *Slug) Beat() bool {
	r.age++
	if r.age > SlugAgeThreshold {
		return false
	}
	r.position = r.position.Move(r.movementDirection, SlugSpeed)
	return true
}

func (r *Slug) Render() rune {
	return SlugRender
}

func (r *Slug) Position() Coordinates {
	return Coordinates{
		x: int(math.Round(r.position.x)),
		y: int(math.Round(r.position.y)),
	}
}
