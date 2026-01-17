---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
name: "Head Guard"
tier: "1"
type: "Leader"
description: "A seasoned guard with a mace, a whistle, and a bellowing voice."
motives_and_tactics: "Arrest, close gates, pin down, seek glory"
difficulty: "15"
thresholds: "7/13"
hp: "7"
stress: "3"
atk: "+4"
attack: "Mace"
range: "Melee"
damage: "1d10+4 phy"
experience: "Commander +2, Local Knowledge +2"
feats:
 - name: "Rally Guards - Action"
   desc: "Spend 2 Fear to spotlight the Head Guard and up to 2d4 allies within Far range."
 - name: "On My Signal - Reaction"
   desc: "Countdown (5). When the Head Guard is in the spotlight for the first time, activate the countdown. It ticks down when a PC makes an attack roll. When it triggers, all Archer Guards within Far range make a standard attack with advantage against the nearest target within their range. If any attacks succeed on the same target, combine their damage."
 - name: "Momentum - Reaction"
   desc: "When the Head Guard makes a successful attack against a PC, you gain a Fear."
```