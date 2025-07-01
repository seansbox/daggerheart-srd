import os
import json

w_dir = os.path.dirname(os.path.abspath(__file__))
os.chdir(w_dir)

# map the json directory
json_dir = os.path.join(w_dir, 'json')

# Check for the JSON file for adversaries
json_file_path = os.path.join(json_dir, 'adversaries.json')
if not os.path.exists(json_file_path):
    raise FileNotFoundError(f"JSON file not found: {json_file_path}")

with open(json_file_path, 'r', encoding='utf-8-sig') as file:
    data = json.load(file)

# Check for or set up the directory structure for statblocks output up one level from .build
statblock_dir = os.path.join('..','adversaries','statblocks')
if not os.path.exists(statblock_dir):
  os.makedirs(statblock_dir)

for adversary in data:
    markdown_output =  f"""---
layout: Daggerheart
image:
name: {adversary['name']}
desc: {adversary['description']}
exp: {adversary['experience'] if 'experience' in adversary else '0'}
mt: {adversary['motives_and_tactics']}
tier: {adversary['tier']}
type: {adversary['type']}
scores: [{adversary['difficulty']}, {adversary['thresholds']}, {adversary['hp']}, {adversary['stress']}]
atk: 1d20{adversary['atk']}
atk_roll: 1d20{adversary['atk']}
atk_dice: 1d20{adversary['atk']}
dmg: {adversary['attack']} - {adversary['range']}
dmg_roll: {adversary['damage']}
dmg_dice: {adversary['damage']}
feats:
  - name: {adversary['feats'][0]['name'] if len(adversary['feats']) > 0 else ''}
    desc: {adversary['feats'][0]['text'] if len(adversary['feats']) > 0 else ''}
  - name: {adversary['feats'][1]['name'] if len(adversary['feats']) > 1 else ''}
    desc: {adversary['feats'][1]['text'] if len(adversary['feats']) > 1 else ''}
  - name: {adversary['feats'][2]['name'] if len(adversary['feats']) > 2 else ''}
    desc: {adversary['feats'][2]['text'] if len(adversary['feats']) > 2 else ''}
  - name: {adversary['feats'][3]['name'] if len(adversary['feats']) > 3 else ''}
    desc: {adversary['feats'][3]['text'] if len(adversary['feats']) > 3 else ''}
  - name: {adversary['feats'][4]['name'] if len(adversary['feats']) > 4 else ''}
    desc: {adversary['feats'][4]['text'] if len(adversary['feats']) > 4 else ''}
  - name: {adversary['feats'][5]['name'] if len(adversary['feats']) > 5 else ''}
    desc: {adversary['feats'][5]['text'] if len(adversary['feats']) > 5 else ''}
  - name: {adversary['feats'][6]['name'] if len(adversary['feats']) > 6 else ''}
    desc: {adversary['feats'][6]['text'] if len(adversary['feats']) > 6 else ''}
  - name: {adversary['feats'][7]['name'] if len(adversary['feats']) > 7 else ''}
    desc: {adversary['feats'][7]['text'] if len(adversary['feats']) > 7 else ''}
---
"""

    tier_dir = os.path.join(statblock_dir, f'Tier {adversary["tier"]}')
    if not os.path.exists(tier_dir):
      os.makedirs(tier_dir)

    statblock_output = os.path.join(w_dir, statblock_dir, f'Tier {adversary['tier']}',f'{adversary['name'].replace(':','')}.md')
    with open(statblock_output, 'w') as statblock:
        statblock.write(markdown_output)
    print(f"Tier {adversary['tier']}/{adversary['name'].replace(':','')}.md'")
