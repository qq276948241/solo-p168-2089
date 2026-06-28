package main

import (
	"fmt"
	"math/rand"
)

const red = "\033[31m"
const reset = "\033[0m"

type MonsterEx struct {
	Name   string
	HP     int
	Max    int
	ATK    int
	BaseATK int
	Gold   int
	Rage   int
}

func NewMonsterEx(m *Monster) *MonsterEx {
	return &MonsterEx{
		Name:    m.Name,
		HP:      m.HP,
		Max:     m.Max,
		ATK:     m.ATK,
		BaseATK: m.ATK,
		Gold:    m.Gold,
		Rage:    0,
	}
}

func (m *MonsterEx) TryRage() bool {
	if m.Rage > 0 {
		return false
	}
	if m.HP < m.Max/2 && rand.Intn(100) < 30 {
		m.Rage = 2
		m.ATK = m.BaseATK * 2
		return true
	}
	return false
}

func (m *MonsterEx) DecRage() {
	if m.Rage > 0 {
		m.Rage--
		if m.Rage == 0 {
			m.ATK = m.BaseATK
		}
	}
}

func DrawMonsterRage(name string, raging bool) {
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
		if raging {
			// 修复4: 颜色通过color()函数降级
			fmt.Printf("       %s%s%s\n", color(red), line, color(reset))
		} else {
			fmt.Printf("       %s\n", line)
		}
	}
	if raging {
		fmt.Printf("       %s【狂暴中！攻击力x2】%s\n", color(red), color(reset))
	}
}

func UseFountain(p *Player) {
	if p.HP >= p.MaxHP {
		fmt.Println()
		fmt.Println("  ══════════════════════════════════════════════")
		fmt.Println("  💧 发现清澈的泉水...但你已是满血状态")
		fmt.Println("  ══════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("  按回车继续...")
		reader.ReadString('\n')
		return
	}
	healed := p.MaxHP - p.HP
	p.HP = p.MaxHP
	fmt.Println()
	fmt.Println("  ══════════════════════════════════════════════")
	fmt.Println("  💧 发现清澈的泉水！沁人心脾！")
	fmt.Printf("  💧 生命值完全恢复！(+%d HP)\n", healed)
	fmt.Println("  ══════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  按回车继续...")
	reader.ReadString('\n')
}

func GetMonsterCount(level int) int {
	return 8 + (level-1)*2
}

func CalculateGrade(kills int, remainingHP, maxHP int) string {
	hpRatio := float64(remainingHP) / float64(maxHP)
	score := float64(kills)*10 + hpRatio*50
	switch {
	case score >= 120:
		return "S"
	case score >= 90:
		return "A"
	case score >= 60:
		return "B"
	case score >= 30:
		return "C"
	default:
		return "D"
	}
}

func PrintGrade(p Player, level int) {
	grade := CalculateGrade(p.Kills, p.HP, p.MaxHP)
	gradeColor := ""
	switch grade {
	case "S":
		gradeColor = color("\033[33m")
	case "A":
		gradeColor = color("\033[32m")
	case "B":
		gradeColor = color("\033[36m")
	case "C":
		gradeColor = color("\033[34m")
	default:
		gradeColor = color("\033[37m")
	}
	resetCode := color("\033[0m")

	fmt.Println()
	fmt.Println("  ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★")
	fmt.Println("  ★                                       ★")
	fmt.Printf("  ★           第 %d 关 通 关 ！               ★\n", level)
	fmt.Println("  ★                                       ★")
	fmt.Printf("  ★  击杀怪物: %2d 只                       ★\n", p.Kills)
	fmt.Printf("  ★  剩余血量: %d/%d                       ★\n", p.HP, p.MaxHP)
	fmt.Printf("  ★  获得金币: %3d 枚                       ★\n", p.Gold)
	fmt.Println("  ★                                       ★")
	fmt.Printf("  ★              评价: %s%s%s                  ★\n", gradeColor, grade, resetCode)
	fmt.Println("  ★                                       ★")
	fmt.Println("  ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★")
	fmt.Println()
}
