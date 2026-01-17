---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Tangle Bramble Swarm"
tier: "1"
type: "Horde (3/HP)"
description: "A cluster of animated, blood-drinking tumbleweeds, each the size of a large gourd."
motives_and_tactics: "Digest, entangle, immobilize"
difficulty: "12"
thresholds: "6/11"
hp: "6"
stress: "3"
atk: "0"
attack: "Thorns"
range: "Melee"
damage: "1d6+3 phy"
experience: "Camouflage +2"
feats:
 - name: "Horde (1d4+2) - Passive"
   desc: "When the Swarm has marked half or more of their HP, their standard attack deals 1d4+2 physical damage instead."
 - name: "Crush - Action"
   desc: "Mark a Stress to deal 2d6+8 direct physical damage to a target with 3 or more bramble tokens."
 - name: "Encumber - Reaction"
   desc: "When the Swarm succeeds on an attack, give the target a bramble token. If a target has any bramble tokens, they are Restrained. If a target has 3 or more bramble tokens, they are also Vulnerable. All bramble tokens can be removed by succeeding on a Finesse Roll (12 + the number of bramble tokens) or dealing Major or greater damage to the Swarm. If bramble tokens are removed from a target using a Finesse Roll, a number of Tangle Bramble Minions spawn within Melee range equal to the number of tokens removed."
```