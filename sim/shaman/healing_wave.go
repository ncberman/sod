package shaman

import (
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
)

const HealingWaveRanks = 10

var HealingWaveSpellId = [HealingWaveRanks + 1]int32{0, 331, 332, 547, 913, 939, 959, 8005, 10395, 10396, 25357}
var HealingWaveBaseHealing = [HealingWaveRanks + 1][]float64{{0}, {36, 47}, {69, 83}, {136, 163}, {279, 328}, {378, 443}, {552, 639}, {759, 874}, {1026, 1177}, {1389, 1583}, {1620, 1850}}
var HealingWaveSpellCoef = [HealingWaveRanks + 1]float64{0, .123, .271, .5, .793, .857, .857, .857, .857, .857, .857}
var HealingWaveCastTime = [HealingWaveRanks + 1]int32{0, 1500, 2000, 2500, 3000, 3000, 3000, 3000, 3000, 3000, 3000}
var HealingWaveManaCost = [HealingWaveRanks + 1]float64{0, 25, 45, 80, 155, 200, 265, 340, 440, 560, 620}
var HealingWaveLevel = [HealingWaveRanks + 1]int{0, 1, 6, 12, 18, 24, 32, 40, 48, 56, 60}

func (shaman *Shaman) registerHealingWaveSpell() {
	overloadRuneEquipped := shaman.HasRune(proto.ShamanRune_RuneChestOverload)

	shaman.HealingWave = make([]*core.Spell, HealingWaveRanks+1)

	if overloadRuneEquipped {
		shaman.HealingWaveOverload = make([]*core.Spell, HealingWaveRanks+1)
	}

	for rank := 1; rank <= HealingWaveRanks; rank++ {
		config := shaman.newHealingWaveSpellConfig(rank, false)

		if config.RequiredLevel <= int(shaman.Level) {
			shaman.HealingWave[rank] = shaman.RegisterSpell(config)

			if overloadRuneEquipped {
				shaman.HealingWaveOverload[rank] = shaman.RegisterSpell(shaman.newHealingWaveSpellConfig(rank, true))
			}
		}
	}
}

func (shaman *Shaman) newHealingWaveSpellConfig(rank int, isOverload bool) core.SpellConfig {
	spellId := HealingWaveSpellId[rank]
	baseHealingMultiplier := 1 + shaman.purificationHealingModifier()
	baseHealingLow := HealingWaveBaseHealing[rank][0] * baseHealingMultiplier
	baseHealingHigh := HealingWaveBaseHealing[rank][1] * baseHealingMultiplier
	spellCoeff := HealingWaveSpellCoef[rank]
	castTime := HealingWaveCastTime[rank]
	manaCost := HealingWaveManaCost[rank]
	level := HealingWaveLevel[rank]

	flags := core.SpellFlagHelpful
	if !isOverload {
		flags |= core.SpellFlagAPL
	}

	spell := core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellId},
		ClassSpellMask: ClassSpellMask_ShamanHealingWave,
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          flags,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
				CastTime: time.Millisecond * time.Duration(
					castTime-(100*shaman.Talents.ImprovedHealingWave),
				),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Take Healing Way into account 6% stacking up to 3x
			result := spell.CalcAndDealHealing(sim, spell.Unit, sim.Roll(baseHealingLow, baseHealingHigh), spell.OutcomeHealingCrit)

			if !isOverload && shaman.procOverload(sim, "Healing Wave Overload", 1) {
				shaman.HealingWaveOverload[rank].Cast(sim, spell.Unit)
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				if shaman.HasRune(proto.ShamanRune_RuneFeetAncestralAwakening) {
					shaman.ancestralHealingAmount = result.Damage * AncestralAwakeningHealMultiplier

					// TODO: this should actually target the lowest health target in the raid.
					//  does it matter in a sim? We currently only simulate tanks taking damage (multiple tanks could be handled here though.)
					shaman.AncestralAwakening.Cast(sim, spell.Unit)
				}
			}
		},
	}

	if isOverload {
		shaman.applyOverloadModifiers(&spell)
	}

	return spell
}
