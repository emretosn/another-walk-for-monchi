# Another Walk for Monchi
This is the code that accompanies the paper "Another Walk for Monchi"

- _Languages_: Go (1.23), C (99+), Python (3.8+)
- _Platforms_: Linux, MacOS
- _Dependencies_: `tuneinsight/lattigo/v4 v4.1.0`
- _Code Author_: Anonymous
- _License_: GNU GPLv3 (any code derived from this must be **open source**)

### Description
Monchi is a privacy-focused biometric identification protocol using homomorphic encryption for score computation and function secret sharing for threshold comparisons.
We build on Bassit et al.'s method, replacing homomorphic multiplications with lookup tables, and extend this by applying function secret sharing for score comparison.
Additionally, we introduce a two-party computation for scores using lookup tables, which integrates seamlessly with function secret sharing.

### Usage
The tests have been conducted in Linux and MacOS.

To run the project:
- First follow the instructions in `funshade/README.md` to generate the shared object files.
- Generate the borders and mfip-tables by running `mfip-tables/genBorderMFIP.py`.
- Place any 128 feature biometric data in `monchi-lut/data/`
- Use `go run .` inside both `monchi-lut` and `monchichi` to execute.

