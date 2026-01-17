---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Vault Guardian Turret"
tier: "3"
type: "Ranged"
description: "A massive hulking turret with reinforced armor and twelve piston-driven mechanical legs."
motives_and_tactics: "Concentrate fire, lock down, mark, protect"
difficulty: "16"
thresholds: "20/32"
hp: "5"
stress: "4"
atk: "+3"
attack: "Magitech Cannon"
range: "Far"
damage: "3d10+3 mag"
feats:
 - name: "Slow Firing - Passive"
   desc: "When you spotlight the Turret and they don’t have a token on their stat block, they can’t make a standard attack. Place a token on their stat block and describe what they’re preparing to do. When you spotlight the Turret and they have a token on their stat block, clear the token and they can attack."
 - name: "Mark Target - Action"
   desc: "Spend a Fear to Mark a target within Far range until the Turret is destroyed or the Marked target becomes Hidden. While the target is Marked, their Evasion is halved."
 - name: "Concentrate Fire - Reaction"
   desc: "When another adversary deals damage to a target within Far range of the Turret, you can mark a Stress to add the Turret’s standard attack damage to the damage roll."
 - name: "Detonation - Reaction"
   desc: "When the Turret is destroyed, they explode. All targets within Close range must make an Agility Reaction Roll. Targets who fail take 3d20 physical damage. Targets who succeed take half damage."
```