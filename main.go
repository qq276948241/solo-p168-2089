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
)

type Tile struct {
	Type    TileType
	Visited bool
}

type Player struct {
	HP, MaxHP int
	ATK, DEF  int
	Gold      int
	Kills     int
	X, Y      int
	Defending bool
}

type Monster struct {
	Name string
	HP   int
	Max  int
	ATK  int
	Gold int
}

var reader = bufio.NewReader(os.Stdin)

func main() {
	rand.Seed(time.Now().UnixNano())
	for {
		runGame()
		fmt.Print("\n按回车重新开始，输入q退出: ")
		input, _ := reader.ReadString('\n')
		if len(input) > 0 && input[0] == 'q' {
			break
		}
	}
}

func runGame() {
	player := Player{
		HP: 50, MaxHP: 50,
		ATK: 10, DEF: 5,
		Gold: 0, Kills: 0,
		X: 0, Y: 0,
	}

	gameMap := generateMap()
	totalMonsters := countMonsters(gameMap)
	origMonsters := totalMonsters

	for {
		clearScreen()
		printStatus(player, totalMonsters, origMonsters)
		printMap(gameMap, player)

		if totalMonsters == 0 {
			printWin(player)
			return
		}
		if player.HP <= 0 {
			printLose(player)
			return
		}
		printControls()

		move := readMove()
		if move == 'q' {
			return
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
				killed := combat(&player, monster)
				if killed {
					totalMonsters--
				}
			case TileChest:
				openChest(&player)
			case TileTrap:
				triggerTrap(&player)
			}
		}
	}
}

func generateMap() [MapSize][MapSize]Tile {
	var m [MapSize][MapSize]Tile
	for y := 0; y < MapSize; y++ {
		for x := 0; x < MapSize; x++ {
			m[y][x] = Tile{Type: TileEmpty, Visited: false}
		}
	}
	m[0][0].Visited = true

	numMonsters := 10
	numChests := 6
	numTraps := 6

	placeRandom(&m, TileMonster, numMonsters)
	placeRandom(&m, TileChest, numChests)
	placeRandom(&m, TileTrap, numTraps)
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
	fmt.Printf("║ HP: %3d/%3d %s  ATK:%2d  DEF:%2d  │ 金币:%3d  │ 杀怪:%2d/%2d ║\n",
		p.HP, p.MaxHP, hpBar, p.ATK, p.DEF, p.Gold, p.Kills, total)
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
	fmt.Println("  @=你   ?=未知   ·=空地   ✗=怪物   □=宝箱   ~=陷阱")
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
	return &Monster{
		Name: n,
		HP:   hp,
		Max:  hp,
		ATK:  5 + lvl*3 + rand.Intn(4),
		Gold: 5 + lvl*5 + rand.Intn(10),
	}
}

func combat(p *Player, m *Monster) bool {
	skillCD := 0
	for {
		clearScreen()
		p.Defending = false
		printCombatStatus(p, m, skillCD)
		if m.HP <= 0 {
			fmt.Printf("  你击败了%s！获得%d金币！\n", m.Name, m.Gold)
			p.Gold += m.Gold
			p.Kills++
			fmt.Println("  按回车继续...")
			reader.ReadString('\n')
			return true
		}
		if p.HP <= 0 {
			return false
		}
		fmt.Print("  选择行动 (1=攻击 2=防御 3=技能): ")
		choice, _ := reader.ReadString('\n')
		playerDmg := 0
		playerMsg := ""
		switch {
		case len(choice) > 0 && choice[0] == '1':
			playerDmg = max(1, p.ATK+rand.Intn(4)-1)
			playerMsg = fmt.Sprintf("  你挥剑砍向%s，造成%d点伤害！", m.Name, playerDmg)
		case len(choice) > 0 && choice[0] == '2':
			p.Defending = true
			playerMsg = "  你举起盾牌进行防御！"
		case len(choice) > 0 && choice[0] == '3':
			if skillCD > 0 {
				fmt.Printf("  技能冷却中！(%d回合)\n", skillCD)
				fmt.Println("  按回车继续...")
				reader.ReadString('\n')
				continue
			}
			playerDmg = max(1, p.ATK*2+rand.Intn(6))
			skillCD = 3
			playerMsg = fmt.Sprintf("  ⚡ 发动必杀技！对%s造成%d点暴击伤害！", m.Name, playerDmg)
		default:
			continue
		}
		if playerDmg > 0 {
			m.HP -= playerDmg
			if m.HP < 0 {
				m.HP = 0
			}
		}
		fmt.Println(playerMsg)
		if m.HP <= 0 {
			fmt.Printf("  %s倒下了！\n", m.Name)
			continue
		}
		monsterDmg := max(1, m.ATK-p.DEF/2+rand.Intn(4)-1)
		if p.Defending {
			monsterDmg = max(0, monsterDmg/2)
		}
		p.HP -= monsterDmg
		if p.HP < 0 {
			p.HP = 0
		}
		defText := ""
		if p.Defending {
			defText = " (防御减半)"
		}
		fmt.Printf("  %s攻击你，造成%d点伤害！%s\n", m.Name, monsterDmg, defText)
		if skillCD > 0 {
			skillCD--
		}
		fmt.Println("  按回车继续...")
		reader.ReadString('\n')
	}
}

func printCombatStatus(p *Player, m *Monster, cd int) {
	hpBar := makeBar(p.HP, p.MaxHP)
	mhpBar := makeBar(m.HP, m.Max)
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %s  HP: %2d/%2d %s  ATK:%2d         ║\n", pad(m.Name, 6), m.HP, m.Max, mhpBar, m.ATK)
	fmt.Printf("║ 你的 HP: %3d/%3d %s  ATK:%2d  DEF:%2d  技能CD:%d        ║\n",
		p.HP, p.MaxHP, hpBar, p.ATK, p.DEF, cd)
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	drawMonster(m.Name)
	fmt.Println()
}

func drawMonster(name string) {
	arts := map[string][]string{
		"史莱姆": {
			"     .-.     ",
			"    (o o)    ",
			"   > ^ <    ",
			"   '---'    ",
		},
		"哥布林": {
			"    .--.    ",
			"   / 0  \\   ",
			"  |  ^  |   ",
			"   \\ - /   ",
			"    'v'     ",
		},
		"骷髅兵": {
			"    +-+     ",
			"   (o.o)    ",
			"    |=|     ",
			"   /   \\   ",
		},
		"蝙蝠群": {
			"  \\/\\/\\/\\  ",
			"  /\\/\\/\\/  ",
			"  \\/\\/\\/\\  ",
		},
		"狼人": {
			"   ,;;;,    ",
			"  ( o o )   ",
			"   ) ^ (    ",
			"  /|   |\\  ",
		},
	}
	art, ok := arts[name]
	if !ok {
		art = arts["史莱姆"]
	}
	for _, line := range art {
		fmt.Printf("       %s\n", line)
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

func printWin(p Player) {
	fmt.Println()
	fmt.Println("  ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★")
	fmt.Println("  ★                                       ★")
	fmt.Println("  ★           通 关 胜 利 ！               ★")
	fmt.Println("  ★                                       ★")
	fmt.Println("  ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★")
	fmt.Printf("  ★  击杀怪物: %2d 只                       ★\n", p.Kills)
	fmt.Printf("  ★  获得金币: %3d 枚                       ★\n", p.Gold)
	fmt.Println("  ★                                       ★")
	fmt.Println("  ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★")
	fmt.Println()
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
