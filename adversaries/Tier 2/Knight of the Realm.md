---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Knight of the Realm"
tier: "2"
type: "Leader"
description: "A decorated soldier with heavy armor and a powerful steed."
motives_and_tactics: "Run down, seek glory, show dominance"
difficulty: "15"
thresholds: "13/26"
hp: "6"
stress: "4"
atk: "+4"
attack: "Longsword"
range: "Melee"
damage: "2d10+4 phy"
experience: "Ancient Knowledge +3, High Society +2, Tactics +2"
feats:
 - name: "Chevalier - Passive"
   desc: "While the Knight is on a mount, they gain a +2 bonus to their Difficulty. When they take Severe damage, they’re knocked from their mount and lose this benefit until they’re next spotlighted."
 - name: "Heavily Armored - Passive"
   desc: "When the Knight takes physical damage, reduce it by 3."
 - name: "Cavalry Charge - Action"
   desc: "If the Knight is mounted, move up to Far range and make a standard attack against a target. On a success, deal 2d8+4 physical damage and the target must mark a Stress."
 - name: "For the Realm! - Action"
   desc: "Mark a Stress to spotlight 1d4+1 allies. Attacks they make while spotlighted in this way deal half damage."
```