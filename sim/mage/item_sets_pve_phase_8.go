package mage

import (
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
)

var ItemSetFireleafRegalia = core.NewItemSet(core.ItemSet{
	Name: "Fireleaf Regalia",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			mage := agent.(MageAgent).GetMage()
			mage.applyScarletEnclaveDamage2PBonus()
		},
		4: func(agent core.Agent) {
			mage := agent.(MageAgent).GetMage()
			mage.applyScarletEnclaveDamage4PBonus()
		},
		6: func(agent core.Agent) {
			mage := agent.(MageAgent).GetMage()
			mage.applyScarletEnclaveDamage6PBonus()
		},
	},
})

// Living Bomb ticks every 1 second and when it explodes it spreads Living Bomb to all targets struck that don't have an active Living Bomb.
// Glaciate now stacks to 8 and Spellfrost Bolt grants 2 stacks per hit.
func (mage *Mage) applyScarletEnclaveDamage2PBonus() {
	label := "S03 - Item - Scarlet Enclave - Mage - Damage 2P Bonus"
	if mage.HasAura(label) {
		return
	}

	aura := core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: label,
	}))

	if mage.HasRune(proto.MageRune_RuneHandsLivingBomb) {
		ticksDelta := LivingBombBaseTickLength - time.Second

		aura.AttachSpellMod(core.SpellModConfig{
			Kind:      core.SpellMod_DotTickLength_Flat,
			ClassMask: ClassSpellMask_MageLivingBomb,
			TimeValue: -ticksDelta,
		}).AttachSpellMod(core.SpellModConfig{
			Kind:      core.SpellMod_DotNumberOfTicks_Flat,
			ClassMask: ClassSpellMask_MageLivingBomb,
			IntValue:  LivingBombBaseNumTicks * int64(ticksDelta.Seconds()),
		}).ApplyOnSpellHitDealt(func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if dot := mage.LivingBomb.Dot(result.Target); !dot.IsActive() && spell.Matches(ClassSpellMask_MageLivingBombExplosion) && result.Landed() {
				dot.Apply(sim)
			}
		})
	}

	if mage.HasRune(proto.MageRune_RuneHandsIceLance) {
		aura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			for _, aura := range mage.GlaciateAuras {
				if aura == nil {
					continue
				}

				aura.MaxStacks += 3
			}
		}).ApplyOnSpellHitDealt(func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Matches(ClassSpellMask_MageSpellfrostBolt) && result.Landed() {
				mage.GlaciateAuras.Get(result.Target).Activate(sim)
				mage.GlaciateAuras.Get(result.Target).AddStack(sim)
			}
		})
	}
}

// Casting Deep Freeze increases the remaining duration of your Icy Veins spell by 8 sec.
// Casting Pyroblast cancels 1 stack of the effect from your Balefire Bolt.
func (mage *Mage) applyScarletEnclaveDamage4PBonus() {
	label := "S03 - Item - Scarlet Enclave - Mage - Damage 4P Bonus"
	if mage.HasAura(label) {
		return
	}

	aura := core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: label,
	}))

	if mage.HasRune(proto.MageRune_RuneBracersBalefireBolt) && mage.Talents.Pyroblast {
		aura.ApplyOnCastComplete(func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(ClassSpellMask_MagePyroblast) && mage.BalefireAura.IsActive() && mage.BalefireAura.GetStacks() > 0 {
				mage.BalefireAura.RemoveStack(sim)
			}
		}, false)
	}

	if mage.HasRune(proto.MageRune_RuneHelmDeepFreeze) && mage.HasRune(proto.MageRune_RuneLegsIcyVeins) {
		aura.ApplyOnCastComplete(func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(ClassSpellMask_MageDeepFreeze) && mage.IcyVeinsAura.IsActive() {
				mage.IcyVeinsAura.UpdateExpires(sim, mage.IcyVeinsAura.ExpiresAt()+time.Second*8)
			}
		}, false)
	}
}

// Reduces the cooldown on your Frozen Orb spell by 20 sec.
// Each time Glaciate is consumed, the cooldown on your Deep Freeze is reduced by 1.0 sec per stack consumed.
// Increases the chance for Arcane Blast to trigger Missile Barrage by 10% and for Fireball and Frostbolt to trigger Missile Barrage by 5%.
// Reduces the cooldown on Fire Blast by 5 sec and Fire Blast now refreshes the duration of your Living Bomb on the target.
func (mage *Mage) applyScarletEnclaveDamage6PBonus() {
	label := "S03 - Item - Scarlet Enclave - Mage - Damage 6P Bonus"
	if mage.HasAura(label) {
		return
	}

	aura := core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: label,
	}))

	if mage.HasRune(proto.MageRune_RuneCloakFrozenOrb) {
		aura.AttachSpellMod(core.SpellModConfig{
			Kind:      core.SpellMod_Cooldown_Flat,
			ClassMask: ClassSpellMask_MageFrozenOrb,
			TimeValue: -time.Second * 20,
		})
	}

	if mage.HasRune(proto.MageRune_RuneHelmDeepFreeze) && mage.HasRune(proto.MageRune_RuneHandsIceLance) {
		aura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			for _, aura := range mage.GlaciateAuras {
				if aura == nil {
					continue
				}

				aura.ApplyOnStacksChange(func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
					if newStacks == 0 && !mage.DeepFreeze.CD.Cooldown.IsReady(sim) {
						mage.DeepFreeze.ModifyRemainingCooldown(sim, -time.Second*time.Duration(oldStacks))
					}
				})
			}
		})
	}

	if mage.HasRune(proto.MageRune_RuneBeltMissileBarrage) {
		if mage.HasRune(proto.MageRune_RuneHandsArcaneBlast) {
			mage.ArcaneBlastMissileBarrageChance += 0.10
			mage.FireballFrostboltMissileBarrageChance += 0.05
		}
	}

	aura.AttachSpellMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: ClassSpellMask_MageFireBlast,
		TimeValue: -time.Second * 5,
	})

	if mage.HasRune(proto.MageRune_RuneHandsLivingBomb) {
		aura.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if dot := mage.LivingBomb.Dot(result.Target); dot.IsActive() && spell.Matches(ClassSpellMask_MageFireBlast) && result.Landed() {
				dot.Refresh(sim)
			}
		}
	}
}

var ItemSetFireleafVestments = core.NewItemSet(core.ItemSet{
	Name: "Fireleaf Vestments",
	Bonuses: map[int32]core.ApplyEffect{
		// Your Arcane Blast has a 10% chance to cause Arcane Tunneling. Arcane Tunneling prevents your Arcane Blast effect from being consumed by the next other Arcane damage spell you cast.
		// In addition, activating Arcane Power resets the cooldown on your Mass Regeneration.
		2: func(agent core.Agent) {
		},
		// Rewind Time also reduces all damage taken by your target by 20% for 8 sec.
		4: func(agent core.Agent) {
		},
		// TODO: Reduces the cooldown of your Arcane Power by 90 sec and increases its duration by 10 sec.
		// While Arcane Power is active, your chance to gain Arcane Tunneling is increased by 10% and each cast of Arcane Blast reduces the remaining cooldown on Mass Regeneration by 1.0 sec.
		6: func(agent core.Agent) {
		},
	},
})
