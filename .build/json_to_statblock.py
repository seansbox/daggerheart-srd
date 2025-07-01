import os
import json

# Set the working directory to the script's directory
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

# Iterate through each adversary in the JSON data and generate the markdown output
for adversary in data:
  
  for feat in adversary['feats']:
    # dynamically create a list of feats by iterating through the feats
    feats = []
    for feat in adversary['feats']:
      # Replace colons with dashes in the name and text fields
      if 'name' in feat and 'text' in feat:
        # Append the formatted feat to the feats list
        feats.append(f"- name: {feat['name'].replace(':',' - ')}\n    desc: {feat['text'].replace(':',' - ')}")
    # Join the feats list into a single string
    feats = '\n  '.join(feats)
      
    markdown_output =  f"""```statblock
layout: Daggerheart
image:
name: {adversary['name'].replace(':',' - ')}
desc: {adversary['description'].replace(':',' - ')}
exp: {adversary['experience'] if 'experience' in adversary else ''}
mt: {adversary['motives_and_tactics'].replace(':',' - ')}
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
  {feats}
```
"""

    # Ensure the tier directory exists
    tier_dir = os.path.join(statblock_dir, f'Tier {adversary["tier"]}')
    if not os.path.exists(tier_dir):
      os.makedirs(tier_dir)

    # Write the markdown output to a file
    statblock_output = os.path.join(statblock_dir, f'Tier {adversary['tier']}',f'{adversary['name'].replace(':','_')}.md')
    with open(statblock_output, 'w') as statblock:
        statblock.write(markdown_output)
    
    # Print the output file path for confirmation
    print(statblock_output)
