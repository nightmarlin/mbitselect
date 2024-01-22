# mbitselect

The mbitselect command automatically chooses the target platform for the tinygo
toolchain based on the device(s) currently connected to your machine. Where
possible, it will target the microbit on the first available volume (mimicking
tinygo)- otherwise it will fall back to a specified default.
