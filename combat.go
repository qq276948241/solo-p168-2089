package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

var combatReader = bufio.NewReader(os.Stdin)

func Combat(p *Player, m *Monster) bool {
	for {
		clearScreen()
		p.Defending = false

		p.Skills.DecreaseCooldowns()
		if m.Rage > 0 {
			m.Rage--
			if m.Rage == 0 {
				m.ATK = m.BaseATK
			}
		}

		printCombatStatus(p, m)
		if m.HP <= 0 {
			// 修复3: 金币显示后主动flush，避免某些终端时序问题导致不显示
			fmt.Printf("  💰  你击败了%s！获得%d金币！\n", m.Name, m.Gold)
			os.Stdout.Sync()
			p.Gold += m.Gold
			p.Kills++
			p.HeavyActive = false
			fmt.Println("  按回车继续...")
			combatReader.ReadString('\n')
			return true
		}
		if p.HP <= 0 {
			return false
		}

		fmt.Println("  1=攻击  2=防御  3=技能")
		fmt.Print("  选择行动: ")
		choice, _ := combatReader.ReadString('\n')
		if len(choice) == 0 {
			continue
		}

		playerMsg := ""
		turnDamage := 0

		switch choice[0] {
		case '1':
			atk := p.ATK
			if p.HeavyActive {
				atk *= p.Skills.HeavyMultiplier()
				p.HeavyActive = false
			}
			turnDamage = max(1, atk+rand.Intn(4)-1)
			m.HP -= turnDamage
			if m.HP < 0 {
				m.HP = 0
			}
			playerMsg = fmt.Sprintf("  你挥剑砍向%s，造成%d点伤害！", m.Name, turnDamage)
		case '2':
			// 修复1: 防御是本回合生效，提示说清楚避免误解
			p.Defending = true
			playerMsg = "  🛡  你举起盾牌防御！本回合受到伤害减半"
		case '3':
			skillKey := showSkillMenu(p)
			if skillKey == "" {
				continue
			}
			result := p.Skills.Use(skillKey, p, m)
			if result.OnCooldown {
				fmt.Println(" ", result.Message)
				fmt.Println("  按回车继续...")
				combatReader.ReadString('\n')
				continue
			}
			playerMsg = result.Message
			if result.HeavyNextTurn {
				p.HeavyActive = true
			}
		default:
			continue
		}

		fmt.Println(playerMsg)
		if m.HP <= 0 {
			fmt.Printf("  %s倒下了！\n", m.Name)
			fmt.Println("  按回车继续...")
			combatReader.ReadString('\n')
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
			// 修复4: 颜色通过color()函数，不支持ANSI的终端自动降级
			fmt.Printf("  %s%s进入狂暴状态！攻击力翻倍！%s\n", color(red), m.Name, color(reset))
		}
		defText := ""
		if p.Defending {
			defText = " (防御减半)"
		}
		fmt.Printf("  %s攻击你，造成%d点伤害！%s\n", m.Name, monsterDmg, defText)

		fmt.Println("  按回车继续...")
		combatReader.ReadString('\n')
	}
}

func showSkillMenu(p *Player) string {
	fmt.Println()
	fmt.Println("  ── 技能列表 ──")
	ids := p.Skills.SkillIDs()
	for i, id := range ids {
		s, _ := p.Skills.GetSkill(id)
		cd := p.Skills.GetCooldown(id)
		cdStr := ""
		if cd > 0 {
			cdStr = fmt.Sprintf(" [CD:%d]", cd)
		}
		fmt.Printf("  %d) %s - %s%s\n", i+1, s.Name(), s.Description(), cdStr)
	}
	fmt.Println("  0) 返回")
	fmt.Print("  选择技能: ")
	choice, _ := combatReader.ReadString('\n')
	if len(choice) == 0 {
		return ""
	}
	switch choice[0] {
	case '1':
		return SkillHeavyID
	case '2':
		return SkillHealID
	case '3':
		return SkillPierceID
	default:
		return ""
	}
}

func printCombatStatus(p *Player, m *Monster) {
	hpBar := makeBar(p.HP, p.MaxHP)
	mhpBar := makeBar(m.HP, m.Max)
	rageText := ""
	if m.Rage > 0 {
		// 修复4: 颜色通过color()函数降级
		rageText = fmt.Sprintf(" %s[狂暴:%d]%s", color(red), m.Rage, color(reset))
	}
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %s  HP: %2d/%2d %s  ATK:%2d %s║\n", pad(m.Name, 6), m.HP, m.Max, mhpBar, m.ATK, rageText)
	skillCdLine := ""
	for _, id := range p.Skills.SkillIDs() {
		s, _ := p.Skills.GetSkill(id)
		cd := p.Skills.GetCooldown(id)
		if cd > 0 {
			skillCdLine += fmt.Sprintf("%s:%d ", s.Name(), cd)
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
