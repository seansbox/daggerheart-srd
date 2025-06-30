import os
import json

w_dir = os.path.dirname(os.path.abspath(__file__))
os.chdir(w_dir)

# Load the JSON file and print the details of the Acid Burrower adversary
# Ensure the JSON file is in the same directory as this script
json_file_path = os.path.join(w_dir, 'adversaries.json')
if not os.path.exists(json_file_path):
    raise FileNotFoundError(f"JSON file not found: {json_file_path}")

with open(json_file_path, 'r') as file:
    data = json.load(file)

#loop through all adversaries in the JSON file
if not isinstance(data, list) or len(data) == 0:
    raise ValueError("JSON data is not in the expected format or is empty.")
if not all(isinstance(adversary, dict) for adversary in data):
    raise ValueError("JSON data contains invalid entries. Each adversary should be a dictionary.")

# Process each adversary and generate the markdown output
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

    # Ensure the output directory exists
    # Ensure the output directory exists
    output_dir = os.path.join(w_dir, 'adversaries\statblock')
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    # Create a subdirectory for the tier if it doesn't exist
    tier_dir = os.path.join(output_dir, f'Tier {adversary["tier"]}')
    if not os.path.exists(tier_dir):
        os.makedirs(tier_dir)
        
    # Write the markdown output to a file
    output_file_path = os.path.join(w_dir, f'adversaries\statblock\Tier {adversary['tier']}\{adversary['name'].replace(':','')}.md')
    with open(output_file_path, 'w') as output_file:
        output_file.write(markdown_output)
    print(f"Tier {adversary['tier']}\{adversary['name'].replace(':','')}.md'")
