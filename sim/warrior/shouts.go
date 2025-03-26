package warrior

import (
	"time"

	"github.com/wowsims/sod/sim/core"
)

const ShoutExpirationThreshold = time.Second * 3

func (warrior *Warrior) newShoutSpellConfig(actionID core.ActionID, rank int32, allyAuras core.AuraArray) *WarriorSpell {
	// Use extra hits to simulate buffing your party for threat
	extraHits := 5 - len(allyAuras)

	return warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL | core.SpellFlagHelpful,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		FlatThreatBonus: float64(core.BattleShoutLevel[rank]),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aura := range allyAuras {
				spell.CalcAndDealOutcome(sim, aura.Unit, spell.OutcomeAlwaysHit)
				aura.Activate(sim)
			}

			for i := 0; i < extraHits; i++ {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			}
		},

		RelatedAuras: []core.AuraArray{allyAuras},
	})
}

func (warrior *Warrior) registerBattleShout() {
	rank := core.LevelToBuffRank[core.BattleShout][warrior.Level]
	actionId := core.BattleShoutSpellId[rank]

	warrior.BattleShout = warrior.newShoutSpellConfig(core.ActionID{SpellID: actionId}, rank, warrior.NewPartyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.BattleShoutAura(unit, warrior.Talents.ImprovedBattleShout, warrior.Talents.BoomingVoice)
	}))
}

func (warrior *Warrior) registerShouts() {
	warrior.registerBattleShout()
}
