package main

import (
	"fmt"
	"math/rand"
)

const (
	SkillHeavyID  = "heavy"
	SkillHealID   = "heal"
	SkillPierceID = "pierce"

	SkillHeavyName  = "重击"
	SkillHealName   = "治疗"
	SkillPierceName = "破甲"

	SkillHeavyDesc  = "下回合伤害翻倍"
	SkillHealDesc   = "恢复20点生命值"
	SkillPierceDesc = "无视防御造成真实伤害"

	SkillHeavyCD  = 3
	SkillHealCD   = 4
	SkillPierceCD = 3

	SkillHeavyValue  = 2
	SkillHealValue   = 20
	SkillPierceValue = 1
)

type SkillResult struct {
	Message       string
	Damage        int
	Heal          int
	HeavyNextTurn bool
	OnCooldown    bool
}

type Skill interface {
	ID() string
	Name() string
	Description() string
	Cooldown() int
	Execute(p *Player, m *Monster) SkillResult
}

type BaseSkill struct {
	id          string
	name        string
	description string
	cooldown    int
	value       int
}

func (s *BaseSkill) ID() string          { return s.id }
func (s *BaseSkill) Name() string        { return s.name }
func (s *BaseSkill) Description() string { return s.description }
func (s *BaseSkill) Cooldown() int       { return s.cooldown }

type HeavySkill struct{ BaseSkill }

func NewHeavySkill() *HeavySkill {
	return &HeavySkill{BaseSkill{
		id:          SkillHeavyID,
		name:        SkillHeavyName,
		description: SkillHeavyDesc,
		cooldown:    SkillHeavyCD,
		value:       SkillHeavyValue,
	}}
}

func (s *HeavySkill) Execute(p *Player, m *Monster) SkillResult {
	return SkillResult{
		Message:       fmt.Sprintf("⚔  你摆出重击姿态，下一次攻击伤害翻倍！"),
		HeavyNextTurn: true,
	}
}

type HealSkill struct{ BaseSkill }

func NewHealSkill() *HealSkill {
	return &HealSkill{BaseSkill{
		id:          SkillHealID,
		name:        SkillHealName,
		description: SkillHealDesc,
		cooldown:    SkillHealCD,
		value:       SkillHealValue,
	}}
}

func (s *HealSkill) Execute(p *Player, m *Monster) SkillResult {
	healAmount := s.value
	if p.HP+healAmount > p.MaxHP {
		healAmount = p.MaxHP - p.HP
	}
	p.HP += healAmount
	return SkillResult{
		Message: fmt.Sprintf("✨  你默念咒语，恢复了%d点生命值！", healAmount),
		Heal:    healAmount,
	}
}

type PierceSkill struct{ BaseSkill }

func NewPierceSkill() *PierceSkill {
	return &PierceSkill{BaseSkill{
		id:          SkillPierceID,
		name:        SkillPierceName,
		description: SkillPierceDesc,
		cooldown:    SkillPierceCD,
		value:       SkillPierceValue,
	}}
}

func (s *PierceSkill) Execute(p *Player, m *Monster) SkillResult {
	dmg := max(1, p.ATK+rand.Intn(4)-1)
	m.HP -= dmg
	if m.HP < 0 {
		m.HP = 0
	}
	return SkillResult{
		Message: fmt.Sprintf("🗡  破甲刺击！无视防御对%s造成%d点真实伤害！", m.Name, dmg),
		Damage:  dmg,
	}
}

type SkillManager struct {
	skills    map[string]Skill
	cooldowns map[string]int
}

func NewSkillManager() *SkillManager {
	sm := &SkillManager{
		skills:    make(map[string]Skill),
		cooldowns: make(map[string]int),
	}
	sm.Register(NewHeavySkill())
	sm.Register(NewHealSkill())
	sm.Register(NewPierceSkill())
	return sm
}

func (sm *SkillManager) Register(s Skill) {
	sm.skills[s.ID()] = s
	sm.cooldowns[s.ID()] = 0
}

func (sm *SkillManager) Use(skillID string, p *Player, m *Monster) SkillResult {
	s, ok := sm.skills[skillID]
	if !ok {
		return SkillResult{Message: "  无效的技能！", OnCooldown: false}
	}
	if sm.cooldowns[skillID] > 0 {
		return SkillResult{
			Message:    fmt.Sprintf("[%s]冷却中！还需%d回合", s.Name(), sm.cooldowns[skillID]),
			OnCooldown: true,
		}
	}
	result := s.Execute(p, m)
	if !result.OnCooldown {
		sm.cooldowns[skillID] = s.Cooldown()
	}
	return result
}

func (sm *SkillManager) DecreaseCooldowns() {
	for k := range sm.cooldowns {
		if sm.cooldowns[k] > 0 {
			sm.cooldowns[k]--
		}
	}
}

func (sm *SkillManager) GetCooldown(skillID string) int {
	return sm.cooldowns[skillID]
}

func (sm *SkillManager) GetSkill(skillID string) (Skill, bool) {
	s, ok := sm.skills[skillID]
	return s, ok
}

func (sm *SkillManager) SkillIDs() []string {
	return []string{SkillHeavyID, SkillHealID, SkillPierceID}
}

func (sm *SkillManager) HeavyMultiplier() int {
	return SkillHeavyValue
}
