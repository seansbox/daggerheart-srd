---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Archer Squadron"
tier: "2"
type: "Horde (2/HP)"
description: "A group of trained archers bearing massive bows."
motives_and_tactics: "Stick together, survive, volley fire"
difficulty: "13"
thresholds: "8/16"
hp: "4"
stress: "3"
atk: "0"
attack: "Longbow"
range: "Far"
damage: "2d6+3 phy"
feats:
 - name: "Horde (1d6+3) - Passive"
   desc: "When the Squadron has marked half or more of their HP, their standard attack deals 1d6+3 physical damage instead."
 - name: "Focused Volley - Action"
   desc: "Spend a Fear to target a point within Far range. Make an attack with advantage against all targets within Close range of that point. Targets the Squadron succeeds against take 1d10+4 physical damage."
 - name: "Suppressing Fire - Action"
   desc: "Mark a Stress to target a point within Far range. Until the next roll with Fear, a creature who moves within Close range of that point must make an Agility Reaction Roll. On a failure, they take 2d6+3 physical damage. On a success, they take half damage."
```