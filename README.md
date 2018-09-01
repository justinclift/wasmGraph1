### A simple experiment with Go Wasm on the HTML5 Canvas (2D).

Online demo: https://justinclift.github.io/wasmGraph1/

This renders several basic objects from 3D space onto the 2D canvas as wireframe
solids, performing various rotation, scale, and transform operations on them via
matrix transforms.

Use the wasd, arrow, and numpad keys (including + and -) to rotate the objects
around the origin.

The code for this started from https://github.com/stdiopt/gowasm-experiments,
and has been fairly radically reworked from there. :smile:
