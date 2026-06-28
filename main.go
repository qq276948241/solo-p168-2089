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
	SkillCD     map[string]int
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
			SkillCD: make(map[string]int),
			Level:   1,
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
				killed := combat(player, monster)
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

func combat(p *Player, m *Monster) bool {
	for {
		clearScreen()
		p.Defending = false

		DecreaseCooldowns(p)
		if m.Rage > 0 {
			m.Rage--
			if m.Rage == 0 {
				m.ATK = m.BaseATK
			}
		}

		printCombatStatus(p, m)
		if m.HP <= 0 {
			fmt.Printf("  你击败了%s！获得%d金币！\n", m.Name, m.Gold)
			p.Gold += m.Gold
			p.Kills++
			p.HeavyActive = false
			fmt.Println("  按回车继续...")
			reader.ReadString('\n')
			return true
		}
		if p.HP <= 0 {
			return false
		}

		fmt.Println("  1=攻击  2=防御  3=技能")
		fmt.Print("  选择行动: ")
		choice, _ := reader.ReadString('\n')
		if len(choice) == 0 {
			continue
		}

		playerMsg := ""
		turnDamage := 0
		skipMonsterTurn := false

		switch choice[0] {
		case '1':
			atk := p.ATK
			if p.HeavyActive {
				atk *= Skills["heavy"].Value
				p.HeavyActive = false
			}
			turnDamage = max(1, atk+rand.Intn(4)-1)
			m.HP -= turnDamage
			if m.HP < 0 {
				m.HP = 0
			}
			playerMsg = fmt.Sprintf("  你挥剑砍向%s，造成%d点伤害！", m.Name, turnDamage)
		case '2':
			p.Defending = true
			playerMsg = "  你举起盾牌进行防御！"
		case '3':
			skillKey := showSkillMenu(p)
			if skillKey == "" {
				continue
			}
			result := UseSkill(p, m, skillKey)
			if result.Damage == 0 && result.Heal == 0 && !result.HeavyNextTurn {
				fmt.Println(" ", result.Message)
				fmt.Println("  按回车继续...")
				reader.ReadString('\n')
				continue
			}
			playerMsg = result.Message
			if result.HeavyNextTurn {
				p.HeavyActive = true
			}
			if result.PierceThisTurn {
				skipMonsterTurn = false
			}
		default:
			continue
		}

		fmt.Println(playerMsg)
		if m.HP <= 0 {
			fmt.Printf("  %s倒下了！\n", m.Name)
			fmt.Println("  按回车继续...")
			reader.ReadString('\n')
			continue
		}

		if skipMonsterTurn {
			fmt.Println("  按回车继续...")
			reader.ReadString('\n')
			continue
		}

		justRaged := false
		if m.Rage == 0 && m.HP < m.Max/2 && rand.Intn(100) < 30 {
			m.Rage = 2
			m.ATK = m.BaseATK * 2
			justRaged = true
		}

		monsterDmg := max(1, m.ATK-p.DEF/2+rand.Intn(4)-1)
		if p.Defending {
			monsterDmg = max(0, monsterDmg/2)
		}
		p.HP -= monsterDmg
		if p.HP < 0 {
			p.HP = 0
		}

		if justRaged {
			fmt.Printf("  %s%s进入狂暴状态！攻击力翻倍！%s\n", red, m.Name, reset)
		}
		defText := ""
		if p.Defending {
			defText = " (防御减半)"
		}
		fmt.Printf("  %s攻击你，造成%d点伤害！%s\n", m.Name, monsterDmg, defText)

		fmt.Println("  按回车继续...")
		reader.ReadString('\n')
	}
}

func showSkillMenu(p *Player) string {
	fmt.Println()
	fmt.Println("  ── 技能列表 ──")
	keys := []string{"heavy", "heal", "pierce"}
	labels := []string{"1", "2", "3"}
	for i, k := range keys {
		cfg := Skills[k]
		cd := p.SkillCD[k]
		cdStr := ""
		if cd > 0 {
			cdStr = fmt.Sprintf(" [CD:%d]", cd)
		}
		fmt.Printf("  %s) %s - %s%s\n", labels[i], cfg.Name, cfg.Description, cdStr)
	}
	fmt.Println("  0) 返回")
	fmt.Print("  选择技能: ")
	choice, _ := reader.ReadString('\n')
	if len(choice) == 0 {
		return ""
	}
	switch choice[0] {
	case '1':
		return "heavy"
	case '2':
		return "heal"
	case '3':
		return "pierce"
	default:
		return ""
	}
}

func printCombatStatus(p *Player, m *Monster) {
	hpBar := makeBar(p.HP, p.MaxHP)
	mhpBar := makeBar(m.HP, m.Max)
	rageText := ""
	if m.Rage > 0 {
		rageText = fmt.Sprintf(" %s[狂暴:%d]%s", red, m.Rage, reset)
	}
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %s  HP: %2d/%2d %s  ATK:%2d %s║\n", pad(m.Name, 6), m.HP, m.Max, mhpBar, m.ATK, rageText)
	skillCdLine := ""
	for k, cfg := range Skills {
		cd := p.SkillCD[k]
		if cd > 0 {
			skillCdLine += fmt.Sprintf("%s:%d ", cfg.Name, cd)
		}
	}
	if skillCdLine == "" {
		skillCdLine = "就绪"
	}
	if p.HeavyActive {
		skillCdLine += " [重击就绪]"
	}
	fmt.Printf("║ 你的 HP: %3d/%3d %s  ATK:%2d  DEF:%2d  CD:%s║\n",
		p.HP, p.MaxHP, hpBar, p.ATK, p.DEF, skillCdLine)
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	DrawMonsterRage(m.Name, m.Rage > 0)
	fmt.Println()
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
