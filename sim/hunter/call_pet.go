package hunter

import (
	"time"

	"github.com/wowsims/sod/sim/core"
)

func (hunter *Hunter) getSummonPetConfig() core.SpellConfig {

	summonMetrics := hunter.pet.NewFocusMetrics(core.ActionID{SpellID: 883})

	return core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 883},
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL,
		RequiredLevel:  10,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if hunter.pet == nil {
				return
			}

			restartSwingAt := sim.CurrentTime
			spell.Unit.AutoAttacks.StopMeleeUntil(sim, restartSwingAt, false)
			spell.Unit.AutoAttacks.StopRangedUntil(sim, restartSwingAt)

			hunter.pet.DistanceFromTarget = hunter.DistanceFromTarget
			hunter.pet.AddFocus(sim, 1000, summonMetrics)
		},
	}
}

func (hunter *Hunter) registerSummonPetSpell() {
	config := hunter.getSummonPetConfig()

	if config.RequiredLevel <= int(hunter.Level) {
		hunter.SummonPet = hunter.GetOrRegisterSpell(config)
	}
}