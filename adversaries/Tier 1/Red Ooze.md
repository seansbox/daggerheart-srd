---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Red Ooze"
tier: "1"
type: "Skulk"
description: "A moving mound of translucent flaming red slime."
motives_and_tactics: "Camouflage, consume and multiply, ignite, start fires"
difficulty: "10"
thresholds: "6/11"
hp: "5"
stress: "3"
atk: "+1"
attack: "Ooze Appendage"
range: "Melee"
damage: "1d8+3 mag"
experience: "Camouflage +3"
feats:
 - name: "Creeping Fire - Passive"
   desc: "The Ooze can only move within Very Close range as their normal movement. They light any flammable object they touch on fire."
 - name: "Ignite - Action"
   desc: "Make an attack against a target within Very Close range. On a success, the target takes 1d8 magic damage and is ignited until they’re extinguished with a successful Finesse Roll (14). While ignited, the target takes 1d4 magic damage when they make an action roll."
 - name: "Split - Reaction"
   desc: "When the Ooze has 3 or more HP marked, you can spend a Fear to split them into two Tiny Red Oozes (with no marked HP or Stress). Immediately spotlight both of them."
```