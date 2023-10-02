# localfs

## Goals

There are different types of files:
- Team Roster
- Lego Themes
- etc.

Some file types has no parameter (Lego Themes file)
Some file types has parameters (Roster has team id and date)

A file's path is determined by its type and its params:
/seasons/2022/team-12/rosters/roster-team-12-2021-11-25.timestamp.extension
/catalog/themes.timestamp.csv.gz

A file can have multiple generations

detectors -> check if filename is matching a file type
finders -> filter a directory content by file type
differs -> compare two files of the same type
