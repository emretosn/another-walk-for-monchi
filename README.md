# Another Walk for Monchi

- _Languages_: Go (1.23), C (99+), Python (3.8+, for the FSS wrapper)
- _Platforms_: Linux, MacOS
- _Dependencies_: `tuneinsight/lattigo/v4 v4.1.0`
- _Code Author_: Anonymous
- _License_: GNU GPLv3 (any code derived from this must be **open source**)
- _Version_: 1.0.0

[//]: # (### Description)

### Usage
The tests have been conducted in Linux and MacOS
To run the project:
- First you have to follow the instructions in `funshade/README.md` to generate the shared object files.
- Generate the borders and mfip-tables with `mfip-tables/genBorderMFIP.py`.
- You can use any 128 feature biometric data to be placed in `monchi/data/`, be sure change any additional subdirectory path in the Go code.
- Use `go run .` to execute.

