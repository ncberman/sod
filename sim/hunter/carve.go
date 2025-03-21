package hunter

import (
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
)

func (hunter *Hunter) registerCarveSpell() {
	if !hunter.HasRune(proto.HunterRune_RuneHandsCarve) {
		return
	}

	hunter.CarveMH = hunter.newCarveHitSpell(true)
	hunter.CarveOH = hunter.newCarveHitSpell(false)

	hunter.RegisterSpell(core.SpellConfig{
		ClassSpellMask: ClassSpellMask_HunterCarve,
		ActionID:       core.ActionID{SpellID: 425711},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		MaxRange: core.MaxMeleeAttackRange,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.04,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
				hunter.CarveMH.Cast(sim, aoeTarget)
				if hunter.AutoAttacks.IsDualWielding {
					hunter.CarveOH.Cast(sim, aoeTarget)
				}
			}
		},
	})
}

func (hunter *Hunter) newCarveHitSpell(isMH bool) *core.Spell {
	procMask := core.ProcMaskMeleeMHSpecial
	damageMultiplier := 0.65
	damageFunc := hunter.MHWeaponDamage

	if !isMH {
		procMask = core.ProcMaskMeleeOHSpecial
		damageMultiplier = hunter.AutoAttacks.OHConfig().DamageMultiplier * 0.65
		damageFunc = hunter.OHWeaponDamage
	}

	return hunter.RegisterSpell(core.SpellConfig{
		ClassSpellMask: ClassSpellMask_HunterCarveHit,
		ActionID:       core.ActionID{SpellID: 425711}.WithTag(core.TernaryInt32(isMH, 1, 2)),
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       procMask,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := damageFunc(sim, spell.MeleeAttackPower())
			if target == hunter.CurrentTarget {
				baseDamage *= 1.5
			}
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})
}
