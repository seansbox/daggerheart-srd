# SRD 2.0 migration audit

Source of truth: `.build/01_pdf/DH-SRD-2.0-2026-08-25.pdf`.

## Classification

- **Unchanged / retained:** the nine SRD 1.0 domains, nine original classes,
  18 original ancestries, nine original communities, and all pre-existing
  machine-facing paths and field names are retained. Existing records remain
  available to downstream consumers.
- **Changed:** all source rules text is replaced by the 224-page SRD 2.0 text
  extraction. Domain descriptions now include their SRD 2.0 class access:
  Blade (Assassin), Bone and Valor (Brawler), Grace (Warlock), Sage (Witch),
  and the new Dread domain (Warlock and Witch).
- **Renamed / superseded:** no supported entity identifier was renamed. The
  pre-existing `adversaries/Outer Realms Corrupter.md` identifier remains the
  compatibility spelling; SRD 2.0 uses the displayed spelling “Outer Realms
  Corruptor.” The former `Loot` heading is superseded in SRD 2.0 by `Loot &
  Items`; existing `items/` paths remain intact.
- **Removed:** no existing content files are removed by this migration.
- **New:** Dread (21 cards), Assassin, Brawler, Warlock, Witch, their eight
  subclasses, Aetheris, Earthkin, Emberkin, Gnome, Skykin, Tidekin, Duneborne,
  Freeborne, Frostborne, Hearthborne, Reborne, Warborne, and the six
  Transformations (Demigod, Ghost, Reanimated, Shapeshifter, Vampire, and
  Werewolf).

## Coverage and provenance

`README.md` and `.build/01_pdf/DH-SRD-2.0-2026-08-25.md` contain the complete
page-preserving extraction of the authoritative SRD 2.0 PDF. The importer
creates `.build/02_csv/srd_2_content.csv` and the corresponding JSON file with
one lossless record per PDF page. This preserves the complete new adversary,
environment, equipment, campaign, and appendix text while the original
1.0-specific parser is incrementally adapted to the changed PDF layout.

The normal category CSV, JSON, and generated Markdown layers contain the new
character-facing entities listed above. `transformations/` is a first-class
category and Transformation cards do not reuse ancestry identifiers.
