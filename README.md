A simple experiment with Go Wasm on the HTML5 Canvas (2D).

This renders several basic objects from 3D space onto the 2D canvas, performing
various rotation, scale, and transform operations on them (via matrix transforms).

No backface culling (yet), so the rear edges can be seen through the objects, this is
just a super simple test for the moment. :wink:

Use the wasd, arrow, or numpad keys to move the square around.  <-- not yet implemented

Note - The code for this started from https://github.com/stdiopt/gowasm-experiments,
and has been fairly radically reworked from there. :smile:
