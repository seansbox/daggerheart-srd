---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "War Wizard"
tier: "2"
type: "Ranged"
description: "A battle-hardened mage trained in destructive magic."
motives_and_tactics: "Develop new spells, seek power, conquer"
difficulty: "16"
thresholds: "11/23"
hp: "5"
stress: "6"
atk: "+4"
attack: "Staff"
range: "Far"
damage: "2d10+4 mag"
experience: "Magical Knowledge +2, Strategize +2"
feats:
 - name: "Battle Teleport - Passive"
   desc: "Before or after making a standard attack, you can mark a Stress to teleport to a location within Far range."
 - name: "Refresh Warding Sphere - Action"
   desc: "Mark a Stress to refresh the Wizard’s “Warding Sphere” reaction."
 - name: "Eruption - Action"
   desc: "Spend a Fear and choose a point within Far range. A Very Close area around that point erupts into impassable terrain. All targets within that area must make an Agility Reaction Roll (14). Targets who fail take 2d10 physical damage and are thrown out of the area. Targets who succeed take half damage and aren’t moved."
 - name: "Arcane Artillery - Action"
   desc: "Spend a Fear to unleash a precise hail of magical blasts. All targets in the scene must make an Agility Reaction Roll. Targets who fail take 2d12 magic damage. Targets who succeed take half damage."
 - name: "Warding Sphere - Reaction"
   desc: "When the Wizard takes damage from an attack within Close range, deal 2d6 magic damage to the attacker. This reaction can’t be used again until the Wizard refreshes it with their “Refresh Warding Sphere” action."
```