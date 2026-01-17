---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Arch-Necromancer"
tier: "4"
type: "Leader"
description: "A decaying mage adorned in dark, tattered robes."
motives_and_tactics: "Corrupt, decay, flee to fight another day, resurrect"
difficulty: "21"
thresholds: "33/66"
hp: "9"
stress: "8"
atk: "+6"
attack: "Necrotic Blast"
range: "Far"
damage: "4d12+8 mag"
experience: "Forbidden Knowledge +3, Wisdom of Centuries +3"
feats:
 - name: "Dance of Death - Action"
   desc: "Mark a Stress to spotlight 1d4 allies. Attacks they make while spotlighted in this way deal half damage, or full damage if you spend a Fear."
 - name: "Beam of Decay - Action"
   desc: "Mark 2 Stress to cause all targets within Far range to make a Strength Reaction Roll. Targets who fail take 2d20+12 magic damage and you gain a Fear. Targets who succeed take half damage. A target who marks 2 or more HP must also mark 2 Stress and becomes Vulnerable until they roll with Hope."
 - name: "Open the Gates of Death - Action"
   desc: "Spend a Fear to summon a Zombie Legion, which appears at Close range and immediately takes the spotlight."
 - name: "Not Today, My Dears - Reaction"
   desc: "When the Necromancer has marked 7 or more of their HP, you can spend a Fear to have them teleport away to a safe location to recover. A PC who succeeds on an Instinct Roll can trace the teleportation magic to their destination."
 - name: "Your Demise is Near - Reaction"
   desc: "Countdown (2d6). When the Necromancer has marked 6 or more of their HP, activate the countdown. When it triggers, deal 2d10+6 direct magic damage to a target within Close range. The Necromancer then clears a number of Stress or HP equal to the number of HP marked by the target from this attack."
```