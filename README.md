# Trailpkg


minimal package manager


to set it up please follow these steps:
    - set TRAILPKG_ROOT env variable in your system that should point to an existing dir for this manager
    to store its state and shit.
    - run `make setup` and run the executable `setup` inside the bin dir.
    - then you can run `make build` and there you go. You got your executable in bin dir.

NOTE that packages insalled with trailpkg aren't accessible systemwide and that they are stored in 
$TRAILPKG_ROOT/bin directory (if they are executable binaries) or in the $TRAILPKG_ROOT/store/[package]-[version] 
(if they are libraries).

NOTE that trailpkg is highly dependant on your system's tooling. 
Runtime dependencies are `tar`, `unzip` and a C (prefferably `gcc`) compiler.


and if you find any issues, feel free NOT to reach out to me.
This project is experimental and isn't taken seriously. 
Let it die in the Elephant graveyard of Github.

Enjoy your fucking packages.


