---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Kraken"
tier: "4"
type: "Solo"
description: "A legendary beast of the sea, bigger than the largest galleon, with sucker-laden tentacles and a terrifying maw."
motives_and_tactics: "Consume, crush, drown, grapple"
difficulty: "20"
thresholds: "35/70"
hp: "11"
stress: "8"
atk: "+7"
attack: "Tentacles"
range: "Close"
damage: "4d12+10 phy"
experience: "Swimming +3"
feats:
 - name: "Relentless (3) - Passive"
   desc: "The Kraken can be spotlighted up to three times per GM turn. Spend Fear as usual to spotlight them."
 - name: "Many Tentacles - Passive"
   desc: "While the Kraken has 7 or fewer marked HP, they can make their standard attack against two targets within range."
 - name: "Grapple and Drown - Action"
   desc: "Make an attack roll against a target within Close range. On a success, mark a Stress to grab them with a tentacle and drag them beneath the water. The target is Restrained and Vulnerable until they break free with a successful Strength Roll or the Kraken takes Major or greater damage. While Restrained and Vulnerable in this way, a target must mark a Stress when they make an action roll."
 - name: "Boiling Blast - Action"
   desc: "Spend a Fear to spew a line of boiling water at any number of targets in a line up to Far range. All targets must succeed on an Agility Reaction Roll or take 4d6+9 physical damage. If a target marks an Armor Slot to reduce the damage, they must also mark a Stress."
 - name: "Momentum - Reaction"
   desc: "When the Kraken makes a successful attack against a PC, you gain a Fear."
```