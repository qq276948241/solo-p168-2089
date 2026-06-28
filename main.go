package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	MapSize = 8
)

type TileType int

const (
	TileEmpty TileType = iota
	TileMonster
	TileChest
	TileTrap
	TileFountain
)

type Tile struct {
	Type    TileType
	Visited bool
}

type Player struct {
	HP, MaxHP   int
	ATK, DEF    int
	Gold        int
	Kills       int
	X, Y        int
	Defending   bool
	HeavyActive bool
	Skills      *SkillManager
	Level       int
}

type Monster struct {
	Name    string
	HP      int
	Max     int
	ATK     int
	BaseATK int
	Gold    int
	Rage    int
}

var reader = bufio.NewReader(os.Stdin)

func main() {
	rand.Seed(time.Now().UnixNano())
	for {
		player := Player{
			HP: 50, MaxHP: 50,
			ATK: 10, DEF: 5,
			Gold: 0, Kills: 0,
			X: 0, Y: 0,
			Skills: NewSkillManager(),
			Level:  1,
		}
		gameOver := false
		for !gameOver {
			won := runLevel(&player)
			if !won {
				printLose(player)
				gameOver = true
			} else {
				PrintGrade(player, player.Level)
				fmt.Print("  按回车进入下一关，输入q退出: ")
				input, _ := reader.ReadString('\n')
				if len(input) > 0 && input[0] == 'q' {
					gameOver = true
				} else {
					player.Level++
					player.X = 0
					player.Y = 0
					player.Kills = 0
				}
			}
		}
		fmt.Print("\n按回车重新开始，输入q退出: ")
		input, _ := reader.ReadString('\n')
		if len(input) > 0 && input[0] == 'q' {
			break
		}
	}
}

func runLevel(player *Player) bool {
	gameMap := generateMap(player.Level)
	totalMonsters := countMonsters(gameMap)
	origMonsters := totalMonsters

	for {
		clearScreen()
		printStatus(*player, totalMonsters, origMonsters)
		printMap(gameMap, *player)

		if totalMonsters == 0 {
			return true
		}
		if player.HP <= 0 {
			return false
		}
		printControls()

		move := readMove()
		if move == 'q' {
			return false
		}

		nx, ny := player.X, player.Y
		switch move {
		case 'w', 'W':
			ny--
		case 's', 'S':
			ny++
		case 'a', 'A':
			nx--
		case 'd', 'D':
			nx++
		}

		if nx < 0 || nx >= MapSize || ny < 0 || ny >= MapSize {
			continue
		}

		player.X, player.Y = nx, ny
		player.Defending = false

		tile := &gameMap[ny][nx]
		if !tile.Visited {
			tile.Visited = true
			switch tile.Type {
			case TileMonster:
				monster := generateMonster()
				killed := Combat(player, monster)
				if killed {
					totalMonsters--
				}
			case TileChest:
				openChest(player)
			case TileTrap:
				triggerTrap(player)
			case TileFountain:
				UseFountain(player)
			}
		}
	}
}

func generateMap(level int) [MapSize][MapSize]Tile {
	var m [MapSize][MapSize]Tile
	for y := 0; y < MapSize; y++ {
		for x := 0; x < MapSize; x++ {
			m[y][x] = Tile{Type: TileEmpty, Visited: false}
		}
	}
	m[0][0].Visited = true

	numMonsters := GetMonsterCount(level)
	numChests := 5
	numTraps := 5
	numFountains := 2

	placeRandom(&m, TileMonster, numMonsters)
	placeRandom(&m, TileChest, numChests)
	placeRandom(&m, TileTrap, numTraps)
	placeRandom(&m, TileFountain, numFountains)
	return m
}

func placeRandom(m *[MapSize][MapSize]Tile, t TileType, count int) {
	placed := 0
	for placed < count {
		x := rand.Intn(MapSize)
		y := rand.Intn(MapSize)
		if x == 0 && y == 0 {
			continue
		}
		if m[y][x].Type == TileEmpty {
			m[y][x].Type = t
			placed++
		}
	}
}

func countMonsters(m [MapSize][MapSize]Tile) int {
	cnt := 0
	for y := 0; y < MapSize; y++ {
		for x := 0; x < MapSize; x++ {
			if m[y][x].Type == TileMonster && !m[y][x].Visited {
				cnt++
			}
		}
	}
	return cnt
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printStatus(p Player, remaining, total int) {
	hpBar := makeBar(p.HP, p.MaxHP)
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ 第%d关  HP: %3d/%3d %s  ATK:%2d  DEF:%2d  │ 金币:%3d  │ 杀怪:%2d/%2d ║\n",
		p.Level, p.HP, p.MaxHP, hpBar, p.ATK, p.DEF, p.Gold, p.Kills, total)
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
}

func makeBar(cur, max int) string {
	width := 12
	filled := 0
	if max > 0 {
		filled = int(float64(cur) / float64(max) * float64(width))
	}
	if cur > 0 && filled == 0 && cur < max {
		filled = 1
	}
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	return bar
}

func printMap(m [MapSize][MapSize]Tile, p Player) {
	fmt.Println()
	for y := 0; y < MapSize; y++ {
		fmt.Print("  ")
		for x := 0; x < MapSize; x++ {
			fmt.Print("+---")
		}
		fmt.Println("+")
		fmt.Print("  ")
		for x := 0; x < MapSize; x++ {
			tile := m[y][x]
			var c = " "
			if p.X == x && p.Y == y {
				c = "@"
			} else if tile.Visited {
				switch tile.Type {
				case TileMonster:
					c = "✗"
				case TileChest:
					c = "□"
				case TileTrap:
					c = "~"
				case TileFountain:
					c = "≈"
				case TileEmpty:
					c = "·"
				}
			} else {
				c = "?"
			}
			fmt.Printf("| %s ", c)
		}
		fmt.Println("|")
	}
	fmt.Print("  ")
	for x := 0; x < MapSize; x++ {
		fmt.Print("+---")
	}
	fmt.Println("+")
	fmt.Println()
	fmt.Println("  @=你   ?=未知   ·=空地   ✗=怪物   □=宝箱   ~=陷阱   ≈=泉水")
	fmt.Println()
}

func printControls() {
	fmt.Println("  WASD = 移动   Q = 退出游戏")
	fmt.Println()
	fmt.Print("  请输入移动方向: ")
}

func readMove() byte {
	s, _ := reader.ReadString('\n')
	if len(s) == 0 {
		return 0
	}
	return s[0]
}

func generateMonster() *Monster {
	names := []string{
		"史莱姆", "哥布林", "骷髅兵", "蝙蝠群", "狼人",
	}
	n := names[rand.Intn(len(names))]
	lvl := rand.Intn(3)
	hp := 8 + lvl*6 + rand.Intn(6)
	atk := 5 + lvl*3 + rand.Intn(4)
	return &Monster{
		Name:    n,
		HP:      hp,
		Max:     hp,
		ATK:     atk,
		BaseATK: atk,
		Gold:    5 + lvl*5 + rand.Intn(10),
		Rage:    0,
	}
}

func pad(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	if len(s) > n {
		s = s[:n]
	}
	return s
}

func openChest(p *Player) {
	heal := 15 + rand.Intn(15)
	atkUp := 2 + rand.Intn(3)
	fmt.Println()
	fmt.Println("  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓")
	fmt.Println("  ▓  发现宝箱！金光闪闪！                      ▓")
	fmt.Printf("  ▓  恢复 %d 点生命值！                       ▓\n", heal)
	fmt.Printf("  ▓  攻击力永久提升 %d 点！                   ▓\n", atkUp)
	fmt.Println("  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓")
	p.HP = min(p.MaxHP, p.HP+heal)
	p.ATK += atkUp
	fmt.Println()
	fmt.Println("  按回车继续...")
	reader.ReadString('\n')
}

func triggerTrap(p *Player) {
	dmg := 8 + rand.Intn(8)
	fmt.Println()
	fmt.Println("  ══════════════════════════════════════════════")
	fmt.Println("  ⚠  踩到陷阱！机关启动！")
	fmt.Printf("  ⚠  受到 %d 点伤害！\n", dmg)
	fmt.Println("  ══════════════════════════════════════════════")
	p.HP -= dmg
	if p.HP < 0 {
		p.HP = 0
	}
	fmt.Println()
	fmt.Println("  按回车继续...")
	reader.ReadString('\n')
}

func printLose(p Player) {
	fmt.Println()
	fmt.Println("  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	fmt.Println("  x                                       x")
	fmt.Println("  x          你 已 阵 亡 ！                x")
	fmt.Println("  x                                       x")
	fmt.Println("  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("  x  击杀怪物: %2d 只                       x\n", p.Kills)
	fmt.Printf("  x  获得金币: %3d 枚                       x\n", p.Gold)
	fmt.Println("  x                                       x")
	fmt.Println("  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	fmt.Println()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
