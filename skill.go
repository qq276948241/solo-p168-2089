package main

import (
	"fmt"
	"math/rand"
)

type SkillConfig struct {
	Name        string
	Description string
	Cooldown    int
	Value       int
}

var Skills = map[string]SkillConfig{
	"heavy": {
		Name:        "重击",
		Description: "下回合伤害翻倍",
		Cooldown:    3,
		Value:       2,
	},
	"heal": {
		Name:        "治疗",
		Description: "恢复20点生命值",
		Cooldown:    4,
		Value:       20,
	},
	"pierce": {
		Name:        "破甲",
		Description: "无视防御造成真实伤害",
		Cooldown:    3,
		Value:       1,
	},
}

type SkillResult struct {
	Message      string
	Damage       int
	Heal         int
	HeavyNextTurn bool
	PierceThisTurn bool
}

func DecreaseCooldowns(p *Player) {
	for k := range p.SkillCD {
		if p.SkillCD[k] > 0 {
			p.SkillCD[k]--
		}
	}
}

func SkillHeavy(p *Player, m *Monster) SkillResult {
	cfg := Skills["heavy"]
	if p.SkillCD["heavy"] > 0 {
		return SkillResult{Message: fmt.Sprintf("[%s]冷却中！还需%d回合", cfg.Name, p.SkillCD["heavy"])}
	}
	p.SkillCD["heavy"] = cfg.Cooldown
	return SkillResult{
		Message:       fmt.Sprintf("⚔  你摆出重击姿态，下一次攻击伤害翻倍！"),
		HeavyNextTurn: true,
	}
}

func SkillHeal(p *Player, m *Monster) SkillResult {
	cfg := Skills["heal"]
	if p.SkillCD["heal"] > 0 {
		return SkillResult{Message: fmt.Sprintf("[%s]冷却中！还需%d回合", cfg.Name, p.SkillCD["heal"])}
	}
	healAmount := cfg.Value
	if p.HP+healAmount > p.MaxHP {
		healAmount = p.MaxHP - p.HP
	}
	p.HP += healAmount
	p.SkillCD["heal"] = cfg.Cooldown
	return SkillResult{
		Message: fmt.Sprintf("✨  你默念咒语，恢复了%d点生命值！", healAmount),
		Heal:    healAmount,
	}
}

func SkillPierce(p *Player, m *Monster) SkillResult {
	cfg := Skills["pierce"]
	if p.SkillCD["pierce"] > 0 {
		return SkillResult{Message: fmt.Sprintf("[%s]冷却中！还需%d回合", cfg.Name, p.SkillCD["pierce"])}
	}
	dmg := max(1, p.ATK+rand.Intn(4)-1)
	m.HP -= dmg
	if m.HP < 0 {
		m.HP = 0
	}
	p.SkillCD["pierce"] = cfg.Cooldown
	return SkillResult{
		Message:        fmt.Sprintf("🗡  破甲刺击！无视防御对%s造成%d点真实伤害！", m.Name, dmg),
		Damage:         dmg,
		PierceThisTurn: true,
	}
}

func UseSkill(p *Player, m *Monster, skillKey string) SkillResult {
	switch skillKey {
	case "heavy":
		return SkillHeavy(p, m)
	case "heal":
		return SkillHeal(p, m)
	case "pierce":
		return SkillPierce(p, m)
	default:
		return SkillResult{Message: "  无效的技能！"}
	}
}
