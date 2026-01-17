---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Huge Green Ooze"
tier: "3"
type: "Skulk"
description: "A translucent green mound of acid taller than most humans."
motives_and_tactics: "Camouflage, creep up, envelop, multiply"
difficulty: "15"
thresholds: "15/30"
hp: "7"
stress: "4"
atk: "+3"
attack: "Ooze Appendage"
range: "Melee"
damage: "3d8+1 mag"
experience: "Blend In +3"
feats:
 - name: "Slow - Passive"
   desc: "When you spotlight the Ooze and they don’t have a token on their stat block, they can’t act yet. Place a token on their stat block and describe what they’re preparing to do. When you spotlight the Ooze and they have a token on their stat block, clear the token and they can act."
 - name: "Acidic Form - Passive"
   desc: "When the Ooze makes a successful attack, the target must mark an Armor Slot without receiving its benefits (they can still use armor to reduce the damage). If they can’t mark an Armor Slot, they must mark an additional HP."
 - name: "Envelop - Action"
   desc: "Make an attack against a target within Melee range. On a success, the Ooze Envelops them and the target must mark 2 Stress. While Enveloped, the target must mark an additional Stress every time they make an action roll. When the Ooze takes Severe damage, all Enveloped targets are freed and the condition is cleared."
 - name: "Split - Reaction"
   desc: "When the Ooze has 4 or more HP marked, you can spend a Fear to split them into two Green Oozes (with no marked HP or Stress). Immediately spotlight both of them."
```