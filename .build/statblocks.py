import json

adversaries = 0
with open("json/adversaries.json", "r", encoding="utf-8-sig") as file:
    adversaries = json.load(file)
environments = 0
with open("json/environments.json", "r", encoding="utf-8-sig") as file:
    environments = json.load(file)

def convert_environment(json):
    md = """---
statblock: inline
---

```statblock
layout: Daggerheart Environment
"""
    for key, value in json.items():
        if key == "feats":
            md += "feats:\n"
            for name, desc in [feat.values() for feat in value]:
                desc_raw = repr(desc)[1:-1]
                md += f' - name: "{name}"\n   desc: "{desc_raw}"\n'
        else:
            md += f'{key}: "{value}"\n'
    md += "```"
    return md

def convert_adversary(json):
    md = """---
statblock: inline
---

```statblock
layout: Daggerheart Adversary
"""
    for key, value in json.items():
        if key == "feats":
            md += "feats:\n"
            for name, desc in [feat.values() for feat in value]:
                md += f' - name: "{name}"\n   desc: "{desc}"\n'
        else:
            md += f'{key}: "{value}"\n'
    md += "```"
    return md

def get_file(json):
    folder = "adversaries" if "atk" in json else "environments"
    name = json["name"]
    if ":" in name:
        name = name.replace(": ", " - ")
    return f"../{folder}/Tier {json['tier']}/{name}.md"

for adversary in adversaries:
    with open(get_file(adversary), "w") as file:
        file.write(convert_adversary(adversary))

for environment in environments:
    with open(get_file(environment), "w") as file:
        file.write(convert_environment(environment))

