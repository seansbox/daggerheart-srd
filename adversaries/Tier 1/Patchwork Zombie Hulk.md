---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Patchwork Zombie Hulk"
tier: "1"
type: "Solo"
description: "A towering gestalt of corpses moving as one, with torso-sized limbs and fists as large as a grown halfling."
motives_and_tactics: "Absorb corpses, flail, hunger, terrify"
difficulty: "13"
thresholds: "8/15"
hp: "10"
stress: "3"
atk: "+4"
attack: "Too Many Arms"
range: "Very Close"
damage: "1d20 phy"
experience: "Intimidation +2, Tear Things Apart +2"
feats:
 - name: "Destructive - Passive"
   desc: "When the Zombie takes Major or greater damage, they mark an additional HP."
 - name: "Flailing Limbs - Passive"
   desc: "When the Zombie makes a standard attack, they can attack all targets within Very Close range."
 - name: "Another for the Pile - Action"
   desc: "When the Zombie is within Very Close range of a corpse, they can incorporate it into themselves, clearing a HP and a Stress."
 - name: "Tormented Screams - Action"
   desc: "Mark a Stress to cause all PCs within Far range to make a Presence Reaction Roll (13). Targets who fail lose a Hope and you gain a Fear for each. Targets who succeed must mark a Stress."
```